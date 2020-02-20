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
	"fmt"

	"crawshaw.io/sqlite/sqlitex"
	"github.com/Factom-Asset-Tokens/factom"
	"github.com/subchen/go-trylock/v2"
)

type dbKeyMREBlock struct {
	dbKeyMR *factom.Bytes32
	factom.EBlock
}

type ParallelChain struct {
	Chain
	eblocks    chan dbKeyMREBlock
	pending    chan []factom.Entry
	syncHeight uint32
	trylock.TryLocker

	issued bool
}

func (chain *ParallelChain) Close() error {
	close(chain.eblocks)
	close(chain.pending)
	return nil
}

func ToParallelChain(chain Chain) *ParallelChain {
	return chain.(*ParallelChain)
}

func (chain *ParallelChain) SetSync(height uint32, dbKeyMR *factom.Bytes32) error {
	if height <= chain.syncHeight {
		return nil
	}
	// SetSync is called inside of State.SetSync on all chains.
	// State.SetSync acquires the State.Lock. In order to avoid deadlock,
	// this function MUST return so that the State.Lock can be released.
	chain.syncHeight = height
	select {
	case eb := <-chain.eblocks:
		// If there is already an eblock in the channel, check to see
		// if its populated.
		if eb.EBlock.IsPopulated() {
			// Just put it back.
			chain.eblocks <- eb
			return nil
		}
		// Otherwise it was simply a marker for the sync height, so we
		// discard it and replace it with the new sync height.
	default:
	}

	chain.eblocks <- dbKeyMREBlock{dbKeyMR, factom.EBlock{Height: height}}

	return nil
}

func (chain *ParallelChain) ApplyEBlockCtx(ctx context.Context,
	dbKeyMR *factom.Bytes32, eb factom.EBlock) error {

	if eb.Height <= chain.syncHeight {
		return nil
	}

	select {
	case chain.eblocks <- dbKeyMREBlock{dbKeyMR, eb}:
		chain.syncHeight = eb.Height
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (chain *ParallelChain) ApplyPendingEntries(
	ctx context.Context, es []factom.Entry) error {

	select {
	case chain.pending <- es:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (state *State) NewParallelChain(ctx context.Context,
	chainID *factom.Bytes32,
	init func(context.Context, *factom.Client,
		factom.EBlock) (Chain, error)) error {

	head := factom.EBlock{ChainID: chainID}
	if err := head.Get(ctx, state.c); err != nil {
		return fmt.Errorf("factom.EBlock.Get(): %w", err)
	}

	pChain := ParallelChain{
		eblocks:    make(chan dbKeyMREBlock, 1),  // DO NOT INCREASE BUFFER SIZE
		pending:    make(chan []factom.Entry, 1), // DO NOT INCREASE BUFFER SIZE
		syncHeight: head.Height,
		TryLocker:  trylock.New(),
	}

	pChain.Lock()

	state.g.Go(func() (err error) {

		defer func() {
			if err != nil {
				err = fmt.Errorf("state.ParallelChain{%v}: %w",
					chainID, err)
			}
		}()

		pChain.Chain, err = init(state.ctx, state.c, head)
		if err != nil {
			return fmt.Errorf("init(): %w", err)
		}

		defer func() {
			pChain.Lock() // lock forever on exit.
			if err := pChain.Chain.Close(); err != nil {
				pChain.ToFactomChain().Log.Errorf(
					"state.Chain.Close(): %v", err)
			}
		}()

		if fatChain, ok := ToFATChain(pChain.Chain); ok &&
			fatChain.IsIssued() {

			state.Lock()
			state.issuedIDs = append(state.issuedIDs, fatChain.ID)
			state.Unlock()
			pChain.issued = true
		}

		pChain.Unlock()

		return pChain.run(state)
	})

	state.track(chainID, &pChain)

	return nil
}
func (chain *ParallelChain) run(state *State) (err error) {
	for {
		select {
		case eb, ok := <-chain.eblocks:
			if !ok {
				return nil
			}
			if err := chain.processEBlock(state, eb); err != nil {
				return err
			}
		case es, ok := <-chain.pending:
			if !ok {
				return nil
			}
			if err := chain.processPending(state, es); err != nil {
				return err
			}
		case <-state.ctx.Done():
			return state.ctx.Err()
		}
	}
}
func (chain *ParallelChain) processEBlock(state *State, eb dbKeyMREBlock) error {

	chain.Lock()
	defer chain.Unlock()

	// Rollback any pending entries on the chain.
	if pending, ok := ToPendingChain(chain.Chain); ok {
		pending.LoadFromCache(&eb.EBlock)
		var err error
		chain.Chain, err = pending.Revert()
		if err != nil {
			return fmt.Errorf("state.PendingChain.Revert(): %w", err)
		}
	}

	if !eb.EBlock.IsPopulated() {
		// An unpopulated EBlock is sent by ParallelChain.SetSync to
		// indicate that we should simply advance the sync height.
		if err := chain.Chain.SetSync(eb.EBlock.Height, eb.dbKeyMR); err != nil {
			return fmt.Errorf("state.Chain.SetSync(): %w",
				err)

		}
		return nil
	}

	if err := eb.GetEntries(state.ctx, state.c); err != nil {
		return fmt.Errorf("factom.EBlock.GetEntries(): %w", err)
	}

	if err := chain.UpdateSidechainData(state.ctx, state.c); err != nil {
		return fmt.Errorf("state.Chain.UpdateSidechainData(): %w", err)
	}
	if err := Apply(chain.Chain, eb.dbKeyMR, eb.EBlock); err != nil {
		return fmt.Errorf("state.Apply(): %w", err)
	}

	if err := sqlitex.ExecScript(chain.ToFactomChain().Conn,
		`PRAGMA main.wal_checkpoint;`); err != nil {
		chain.ToFactomChain().Log.Error(err)
	}

	if !chain.issued {
		if fatChain, ok := ToFATChain(chain); ok &&
			fatChain.IsIssued() {

			state.Lock()
			state.issuedIDs = append(state.issuedIDs, fatChain.ID)
			state.Unlock()
			chain.issued = true
		}
	}

	return nil
}
func (chain *ParallelChain) processPending(state *State, es []factom.Entry) error {

	chain.Lock()
	defer chain.Unlock()

	// Initialize Pending if Entries is not yet populated.
	pending, ok := ToPendingChain(chain.Chain)
	if !ok {
		var err error
		if pending, err = NewPendingChain(state.ctx, state.c,
			chain.Chain); err != nil {
			return fmt.Errorf("state.NewPendingChain(): %w", err)
		}
		chain.Chain = pending
	}

	// startLenEntries tracks the initial size of our cache so we can
	// detect if any new pending entries get applied.
	startLenEntries := len(pending.Entries)

	if err := pending.ApplyPendingEntries(es); err != nil {
		return fmt.Errorf("state.PendingChain.ApplyEntry(): %w", err)
	}

	// Check if any no new entries were added.
	if startLenEntries != len(pending.Entries) {
		pending.ToFactomChain().Log.Debugf("Applied %v new pending entries.",
			len(pending.Entries)-startLenEntries)
	}

	return nil
}
