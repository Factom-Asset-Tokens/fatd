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

package engine

import (
	"context"
	"fmt"
	"math"
	"sync"

	"crawshaw.io/sqlite"
	"github.com/Factom-Asset-Tokens/factom"
	"github.com/Factom-Asset-Tokens/fatd/internal/db"
	"github.com/Factom-Asset-Tokens/fatd/internal/flag"
)

var (
	Chains = ChainMap{m: map[factom.Bytes32]Chain{
		factom.Bytes32{31: 0x0a}: Chain{ChainStatus: ChainStatusIgnored},
		factom.Bytes32{31: 0x0c}: Chain{ChainStatus: ChainStatusIgnored},
		factom.Bytes32{31: 0x0f}: Chain{ChainStatus: ChainStatusIgnored},
	}, RWMutex: new(sync.RWMutex)}
)

type ChainMap struct {
	m          map[factom.Bytes32]Chain
	issuedIDs  []*factom.Bytes32
	trackedIDs []*factom.Bytes32
	*sync.RWMutex
}

func (cm *ChainMap) set(id *factom.Bytes32, chain Chain, prevStatus ChainStatus) {
	cm.Lock()
	defer cm.Unlock()
	cm.m[*id] = chain
	if chain.ChainStatus != prevStatus {
		switch chain.ChainStatus {
		case ChainStatusIssued:
			cm.issuedIDs = append(cm.issuedIDs, id)
			fallthrough
		case ChainStatusTracked:
			if prevStatus.IsUnknown() {
				cm.trackedIDs = append(cm.trackedIDs, id)
			}
		}
	}
}

func (cm *ChainMap) ignore(id *factom.Bytes32) {
	cm.set(id, Chain{ChainStatus: ChainStatusIgnored}, ChainStatusIgnored)
}

func (cm *ChainMap) get(id *factom.Bytes32) Chain {
	cm.RLock()
	defer cm.RUnlock()
	return cm.m[*id]
}

func (cm *ChainMap) GetIssued() []*factom.Bytes32 {
	cm.RLock()
	defer cm.RUnlock()
	return cm.issuedIDs
}

func (cm *ChainMap) GetTracked() []*factom.Bytes32 {
	cm.RLock()
	defer cm.RUnlock()
	return cm.trackedIDs
}

func (cm *ChainMap) setSync(height uint32, dbKeyMR *factom.Bytes32) error {
	cm.Lock()
	defer cm.Unlock()
	for _, chain := range cm.m {
		if !chain.IsTracked() {
			continue
		}
		if err := chain.SetSync(height, dbKeyMR); err != nil {
			chain.Log.Errorf("chain.SetSync(): %v", err)
			return err
		}
		cm.m[*chain.ID] = chain
	}
	return nil
}

func (cm *ChainMap) Close() {
	cm.Lock()
	defer cm.Unlock()
	for _, chain := range cm.m {
		if chain.IsTracked() {
			// Rollback any pending entries on the chain.
			if chain.Pending.Entries != nil {
				// Always clean up.
				if err := chain.revertPending(); err != nil {
					log.Error(err)
				}
			}
			chain.Close()
		}
	}
}

// loadChains loads all chains from the database that are not blacklisted, and
// syncs them. Any whitelisted chains that are not previously tracked are
// synced. The lowest sync height among all chain databases is returned.
func loadChains(ctx context.Context) (syncHeight uint32, err error) {
	dbChains, err := db.OpenAll(ctx, flag.DBPath)
	if err != nil {
		return
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
	Chains.Lock()
	defer Chains.Unlock()
	// Set whitelisted chains to Tracked.
	for _, chainID := range flag.Whitelist {
		Chains.m[chainID] = Chain{ChainStatus: ChainStatusTracked}
	}
	// Blacklist overrides whitelist. Set chains to Ignore.
	for _, chainID := range flag.Blacklist {
		Chains.m[chainID] = Chain{ChainStatus: ChainStatusIgnored}
	}

	if len(dbChains) > 0 {
		syncHeight = math.MaxUint32
	}
	for i, dbChain := range dbChains {
		chain := Chains.m[*dbChain.ID]

		// Close and skip any blacklisted chains or, if there was a
		// whitelist, any non-tracked chain.
		if chain.IsIgnored() || flag.HasWhitelist() && !chain.IsTracked() {
			dbChain.Close()
			// Prevent double close in defer on error.
			dbChains[i].Conn = nil
			continue
		}

		chain.Chain = dbChain

		syncHeight = min(syncHeight, chain.SyncHeight)

		if chain.NetworkID != flag.NetworkID {
			err = fmt.Errorf("invalid NetworkID: %v for Chain{%v}",
				chain.NetworkID, chain.ID)
			return
		}

		if !flag.SkipDBValidation {
			if err = chain.Validate(); err != nil {
				return
			}
		} else {
			// Ensure WAL file exists which is required for the
			// Snapshots used for pending transactions to work.
			var begin, commit *sqlite.Stmt
			begin, _, err = chain.Conn.PrepareTransient("BEGIN IMMEDIATE;")
			if err != nil {
				panic(err)
			}
			defer begin.Finalize()
			if _, err = begin.Step(); err != nil {
				return
			}
			commit, _, err = chain.Conn.PrepareTransient("COMMIT;")
			if err != nil {
				panic(err)
			}
			defer commit.Finalize()
			if _, err = commit.Step(); err != nil {
				return
			}
		}

		if err = chain.Sync(ctx, c); err != nil {
			return
		}

		chain.ChainStatus = ChainStatusTracked
		Chains.trackedIDs = append(Chains.trackedIDs, chain.ID)
		if chain.Issuance.IsPopulated() {
			chain.ChainStatus = ChainStatusIssued
			Chains.issuedIDs = append(Chains.issuedIDs, chain.ID)
		}

		Chains.m[*chain.ID] = chain
	}

	// Open any whitelisted chains that do not already have databases.
	for id, chain := range Chains.m {
		if !(chain.IsTracked() && chain.Chain.Conn == nil) {
			continue
		}
		id := id
		var chain Chain
		chain, err = OpenNewByChainID(ctx, c, &id)
		if err != nil {
			return
		}
		Chains.trackedIDs = append(Chains.trackedIDs, chain.ID)
		if chain.IsIssued() {
			Chains.issuedIDs = append(Chains.issuedIDs, chain.ID)
		}
		Chains.m[*chain.ID] = chain

		// Ensure that this new chain gets closed in the defer if an
		// error occurs.
		dbChains = append(dbChains, chain.Chain)
	}

	return
}
func min(a, b uint32) uint32 {
	if a <= b {
		return a
	}
	return b
}
