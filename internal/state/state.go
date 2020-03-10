// MIT License
//
// Copyright 2018 Canonical Ledgers, LLC
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
// IN THE SOFTWARE.

package state

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"sync"

	"github.com/Factom-Asset-Tokens/factom"
	"github.com/Factom-Asset-Tokens/fatd/internal/db"
	"github.com/Factom-Asset-Tokens/fatd/internal/log"
	"github.com/nightlyone/lockfile"
	"golang.org/x/sync/errgroup"
)

type State struct {
	Chains     map[factom.Bytes32]Chain
	issuedIDs  []*factom.Bytes32
	trackedIDs []*factom.Bytes32
	sync.RWMutex

	DBPath    string
	NetworkID factom.NetworkID

	IgnoreNewChains bool

	g   *errgroup.Group
	ctx context.Context

	c *factom.Client

	SyncHeight  uint32
	SyncDBKeyMR *factom.Bytes32

	Log log.Log

	Lockfile lockfile.Lockfile
}

func (state *State) track(id *factom.Bytes32, chain Chain) {
	state.Lock()
	defer state.Unlock()

	if _, ok := state.Chains[*id]; ok {
		panic(fmt.Errorf("Chain already tracked: %v", id))
	}

	state.trackedIDs = append(state.trackedIDs, id)

	state.Chains[*id] = chain
}

func (state *State) ignore(id *factom.Bytes32) {
	state.Lock()
	defer state.Unlock()
	state.Chains[*id] = nil
}

func (state *State) get(id *factom.Bytes32) (Chain, bool) {
	state.RLock()
	defer state.RUnlock()
	chain, ok := state.Chains[*id]
	return chain, ok
}

func (state *State) IssuedIDs() []*factom.Bytes32 {
	state.RLock()
	defer state.RUnlock()
	return state.issuedIDs
}

func (state *State) TrackedIDs() []*factom.Bytes32 {
	state.RLock()
	defer state.RUnlock()
	return state.trackedIDs
}

func (state *State) SetSync(ctx context.Context,
	height uint32, dbKeyMR *factom.Bytes32) error {

	state.Lock()
	defer state.Unlock()
	for id, chain := range state.Chains {
		if chain == nil {
			continue
		}

		if err := chain.SetSync(height, dbKeyMR); err != nil {
			return fmt.Errorf("Chain{%v}.SetSync(): %w", id, err)
		}
	}
	state.SyncHeight = height
	state.SyncDBKeyMR = dbKeyMR
	return nil
}

func (state *State) GetSync() uint32 {
	state.RLock()
	defer state.RUnlock()

	return state.SyncHeight
}

func (state *State) Close() {
	state.Lock()
	defer state.Unlock()
	for _, chain := range state.Chains {
		if chain != nil {
			chain.Close()
		}
	}
	if err := state.g.Wait(); err != nil {
		if !errors.Is(err, context.Canceled) {
			state.Log.Errorf("state.State.g.Wait(): %v", err)
		}
	}
	if err := state.Lockfile.Unlock(); err != nil {
		state.Log.Errorf("lockfile.Lockfile.Unlock(): %w", err)
	}
}

func Open(ctx context.Context, c *factom.Client,
	dbPath string,
	networkID factom.NetworkID,
	whitelist, blacklist []factom.Bytes32,
	skipDBValidation, repair bool,
) (_ *State, _ context.Context, err error) {
	log := log.New("pkg", "state")
	// Try to create the database directory.
	if err := os.Mkdir(dbPath, 0755); err != nil {
		if !os.IsExist(err) {
			return nil, nil,
				fmt.Errorf("os.Mkdir(%q): %w", dbPath, err)
		}
		log.Debugf("Loading state from %q...", dbPath)
	} else {
		log.Debugf("New database directory created at %q.", dbPath)
	}

	log.Debugf("Locking database directory...")
	// Try to create a lockfile
	lockFilePath := dbPath + "db.lock"
	lockFile, err := lockfile.New(lockFilePath)
	if err != nil {
		return nil, nil,
			fmt.Errorf("lockfile.New(%q): %w", lockFilePath, err)
	}
	if err = lockFile.TryLock(); err != nil {
		return nil, nil,
			fmt.Errorf("lockfile.Lockfile.TryLock(): %w", err)
	}
	// Always clean up the lockfile if Start fails.
	defer func() {
		if err != nil {
			if err := lockFile.Unlock(); err != nil {
				log.Errorf("lockfile.Lockfile.Unlock(): %v", err)
			}
		}
	}()

	g, ctx := errgroup.WithContext(ctx)
	state := State{
		Log:      log,
		Lockfile: lockFile,
		Chains: map[factom.Bytes32]Chain{
			factom.Bytes32{31: 0x0a}: nil,
			factom.Bytes32{31: 0x0c}: nil,
			factom.Bytes32{31: 0x0f}: nil,
		},
		DBPath:    dbPath,
		NetworkID: networkID,

		g: g, ctx: ctx, c: c,
	}

	if err := state.loadFATChains(dbPath,
		whitelist, blacklist,
		skipDBValidation, repair); err != nil {
		return nil, nil, err
	}

	return &state, ctx, nil
}

func (state *State) loadFATChains(dbPath string,
	whitelist, blacklist []factom.Bytes32,
	skipDBValidation, repair bool) (err error) {

	var synced sync.WaitGroup
	defer func() {
		synced.Wait()
		if state.ctx.Err() != nil {
			if e := state.g.Wait(); e != nil {
				err = e
			}
		}
	}()

	dbChains, err := db.OpenAllFATChains(state.ctx, dbPath)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			for _, chain := range dbChains {
				if chain.Conn != nil {
					chain.Close()
				}
			}
		}
	}()

	// Set whitelisted chains.
	for _, chainID := range whitelist {
		state.Chains[chainID] = UnknownChain{}
	}
	// Blacklist overrides whitelist. Set chains to Ignore.
	for _, chainID := range blacklist {
		state.Chains[chainID] = nil
	}

	if len(dbChains) > 0 {
		state.SyncHeight = math.MaxUint32
	}
	for i, dbChain := range dbChains {
		// Skip this chain if
		if saved, ok := state.Chains[*dbChain.ID]; true && // just formatting
			// - it is blacklisted,
			(ok && saved == nil) ||
			// - there is a whitelist, and it is not it
			(whitelist != nil && !ok) {

			// Close this chain since we won't use it.
			dbChain.Close()
			// Prevent double close in defer on error.
			dbChains[i].Conn = nil

			// Next dbChain...
			continue
		}

		chain := FATChain(dbChain)
		state.SyncHeight = min(state.SyncHeight, chain.SyncHeight)

		if dbChain.NetworkID != state.NetworkID {
			return fmt.Errorf("invalid NetworkID: %v for Chain{%v}",
				chain.NetworkID, chain.ID)
		}

		init := func(ctx context.Context, c *factom.Client,
			head factom.EBlock) (_ Chain, err error) {

			defer synced.Done()

			if !skipDBValidation {
				if err := chain.Validate(ctx, repair); err != nil {
					return nil, fmt.Errorf(
						"state.FATChain.Validate(): %w", err)
				}
			}

			eblocks, err := head.GetPrevN(ctx, c,
				head.Sequence-chain.Head.Sequence)
			if err != nil {
				return nil, fmt.Errorf(
					"factom.EBlock.GetPrevBackTo(): %w", err)
			}

			if err := SyncEBlocks(ctx, c, &chain, eblocks); err != nil {
				return nil, fmt.Errorf(
					"state.SyncEBlocks(): %w", err)
			}

			return &chain, nil
		}

		synced.Add(1)
		dbChains[i].Conn = nil          // Prevent double close.
		delete(state.Chains, *chain.ID) // Prevent double track.
		if err = state.NewParallelChain(state.ctx,
			chain.ID, init); err != nil {
			synced.Done()
			return fmt.Errorf("state.State.NewParallelChain(): %w", err)
		}
	}

	// Open any whitelisted chains that do not already have databases.
	for id, chain := range state.Chains {
		if _, ok := chain.(UnknownChain); !ok {
			continue
		}
		id := id
		init := func(ctx context.Context, c *factom.Client,
			head factom.EBlock) (_ Chain, err error) {

			defer synced.Done()

			chain, err := NewFATChainByEBlock(ctx, c,
				state.DBPath, head)
			if err != nil {
				return nil, fmt.Errorf(
					"state.NewFATChainByChainID(): %w", err)
			}
			return &chain, nil
		}

		synced.Add(1)
		delete(state.Chains, id) // prevent double track panic.
		if err = state.NewParallelChain(state.ctx, &id, init); err != nil {
			synced.Done()
			return fmt.Errorf("state.State.NewParallelChain(): %w", err)
		}
	}

	return
}
func min(a, b uint32) uint32 {
	if a <= b {
		return a
	}
	return b
}
