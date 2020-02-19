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

	"github.com/Factom-Asset-Tokens/factom"
	"github.com/Factom-Asset-Tokens/factom/fat"
)

func Apply(chain Chain, dbKeyMR *factom.Bytes32, eb factom.EBlock) (err error) {
	defer chain.Save()(&err)

	//chain.ToFactomChain().Log.Debugf("Applying EBlock %v...", eb.KeyMR)

	if err := chain.ApplyEBlock(dbKeyMR, eb); err != nil {
		return fmt.Errorf("state.Chain.ApplyEBlock(): %w", err)
	}

	// Insert each entry and attempt to apply it...
	for _, e := range eb.Entries {
		if _, err := chain.ApplyEntry(e); err != nil {
			return fmt.Errorf("state.Chain.ApplyEntry(): %w", err)
		}
	}

	if err := chain.SetSync(eb.Height, dbKeyMR); err != nil {
		return fmt.Errorf("state.Chain.SetSync(): %w", err)
	}

	return nil
}

func (state *State) ApplyEBlock(ctx context.Context,
	dbKeyMR *factom.Bytes32, eb factom.EBlock) (err error) {
	chain, ok := state.get(eb.ChainID)

	// Skip ignored chains and if we are ignoring new chain, also skip
	// unknown chains.
	if chain == nil {
		if ok || state.IgnoreNewChains {
			return nil
		}
	}

	if err := eb.Get(ctx, state.c); err != nil {
		return fmt.Errorf("factom.EBlock.Get(): %w", err)
	}

	if chain == nil { // if Chain is unknown...
		if !eb.IsFirst() { // if Chain is not new...
			state.ignore(eb.ChainID)
			return nil
		}

		// Load first entry of new chain.
		first := &eb.Entries[0]
		if err := first.Get(ctx, state.c); err != nil {
			return fmt.Errorf("factom.Entry.Get(): %w", err)
		}

		// Ignore chains with NameIDs that don't match the fat pattern.
		nameIDs := first.ExtIDs
		if !fat.ValidNameIDs(nameIDs) {
			state.ignore(eb.ChainID)
			return nil
		}

		state.Log.Infof("Tracking new FAT chain: %v", eb.ChainID)

		init := func(ctx context.Context, c *factom.Client,
			head factom.EBlock) (_ Chain, err error) {

			tokenID, issuerID := fat.ParseTokenIssuer(nameIDs)

			// Attempt to open a new chain.
			fatChain, err := NewFATChain(ctx, state.c, state.DBPath,
				tokenID, &issuerID, eb.ChainID, state.NetworkID)
			if err != nil {
				return nil, fmt.Errorf("state.NewFATChain(): %w", err)
			}
			defer func() {
				if err != nil {
					chain.Close()
				}
			}()

			fatChain.Log.Info("Downloading all EBlocks...")
			eblocks, err := head.GetPrevN(ctx, c, head.Sequence)
			if err != nil {
				return nil, fmt.Errorf(
					"factom.EBlock.GetPrevBackTo(): %w", err)
			}

			fatChain.Log.Info("Syncing entries...")
			if err = eb.GetEntries(ctx, c); err != nil {
				err = fmt.Errorf("factom.EBlock.GetEntries(): %w", err)
				return
			}
			if err := chain.UpdateSidechainData(
				state.ctx, state.c); err != nil {
				return nil, fmt.Errorf(
					"state.Chain.UpdateSidechainData(): %w", err)
			}
			if err = Apply(&fatChain, dbKeyMR, eb); err != nil {
				err = fmt.Errorf("state.Apply(): %w", err)
				return
			}

			if err := SyncEBlocks(ctx, c, &fatChain, eblocks); err != nil {
				return nil, fmt.Errorf(
					"state.SyncEBlocks(): %w", err)
			}

			return &fatChain, nil
		}

		if err = state.NewParallelChain(ctx, eb.ChainID, init); err != nil {
			return fmt.Errorf("state.State.NewParallelChain(): %w", err)
		}

		return nil
	}

	// Apply this EBlock to the chain.
	if ToParallelChain(chain).ApplyEBlockCtx(ctx, dbKeyMR, eb); err != nil {
		return fmt.Errorf("chain.ApplyEBlockCtx(): %w", err)
	}

	return nil
}

func (state *State) ApplyPendingEntries(ctx context.Context,
	es []factom.Entry) error {

	chainID := es[0].ChainID
	chain, _ := state.get(chainID)

	// We can only apply pending entries to Tracked chains.
	if chain == nil {
		return nil
	}

	return ToParallelChain(chain).ApplyPendingEntries(ctx, es)
}

// Get returns a threadsafe connection to the database, and a function to
// release the connection back to the pool. If pending is true, the chain will
// reflect the state with pending entries applied. Otherwise the chain will
// reflect the official state after the most recent EBlock.
func (state *State) Get(ctx context.Context,
	id *factom.Bytes32, includePending bool) (_ Chain, _ func(), err error) {

	chain, _ := state.get(id)
	if chain == nil {
		return chain, nil, nil
	}

	pChain := ToParallelChain(chain)
	if ok := pChain.RTryLock(ctx); !ok {
		return nil, nil, ctx.Err()
	}
	defer pChain.RUnlock()

	// Pull a Conn off the Pool and set it as the main Conn.
	factomChain := chain.ToFactomChain()
	if ok := factomChain.CloseMtx.RTryLock(ctx); !ok {
		return nil, nil, ctx.Err()
	}
	defer func() {
		if err != nil {
			factomChain.CloseMtx.RUnlock()
		}
	}()
	read := factomChain.Pool.Get(ctx)
	if read == nil {
		return nil, nil, ctx.Err()
	}
	defer func() {
		if err != nil {
			factomChain.Pool.Put(read)
		}
	}()

	pending, ok := ToPendingChain(chain)
	if ok {
		chain = pending.Chain
	}

	// If includePending or if there is no pending state, then use the
	// chain as is, and just return a function that returns the conn to the
	// pool.
	if includePending || !ok {
		chain = chain.Copy()
		chain.ToFactomChain().Conn = read
		return chain, func() {
			factomChain.Pool.Put(read)
			factomChain.CloseMtx.RUnlock()
		}, nil
	}
	// There are pending entries, but we have been asked for the official
	// state.

	// Start a read transaction on the conn that reflects the official
	// state.
	endRead, err := read.StartSnapshotRead(pending.OfficialSnapshot)
	if err != nil {
		return nil, nil, fmt.Errorf("sqlite.Conn.StartSnapshotRead(): %w", err)
	}

	// Use the official chain state with the conn from the Pool.
	chain = pending.OfficialState.Copy()
	chain.ToFactomChain().Conn = read

	// Return a function that ends the read transaction and returns the
	// conn to the Pool.
	return chain, func() {
		// We must clear the interrupt to prevent endRead from
		// panicking.
		read.SetInterrupt(nil)
		endRead()
		factomChain.Pool.Put(read)
		factomChain.CloseMtx.RUnlock()
	}, nil
}
