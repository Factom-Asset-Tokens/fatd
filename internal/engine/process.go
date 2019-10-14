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
	"bytes"
	"context"
	"fmt"
	"time"

	"crawshaw.io/sqlite"
	"crawshaw.io/sqlite/sqlitex"
	jsonrpc2 "github.com/AdamSLevy/jsonrpc2/v12"
	"github.com/Factom-Asset-Tokens/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat"
	"github.com/Factom-Asset-Tokens/fatd/internal/db"
	"github.com/Factom-Asset-Tokens/fatd/internal/flag"
)

func Process(ctx context.Context, dbKeyMR *factom.Bytes32, eb factom.EBlock) error {
	chain := Chains.get(eb.ChainID)

	// Skip ignored chains and if we are ignoring new chain, also skip
	// unknown chains.
	if chain.IsIgnored() ||
		(flag.IgnoreNewChains() && chain.IsUnknown()) {
		return nil
	}
	if chain.IsUnknown() {
		if err := eb.Get(ctx, c); err != nil {
			return fmt.Errorf("%#v.Get(): %w", eb, err)
		}
		if !eb.IsFirst() {
			Chains.ignore(eb.ChainID)
			return nil
		}

		// Load first entry of new chain.
		first := &eb.Entries[0]
		if err := first.Get(ctx, c); err != nil {
			return fmt.Errorf("%#v.Get(): %w", first, err)
		}
		// Ignore chains with NameIDs that don't match the fat pattern.
		nameIDs := first.ExtIDs
		if !fat.ValidTokenNameIDs(nameIDs) {
			Chains.ignore(eb.ChainID)
			return nil
		}

		// Attempt to open a new chain.
		var err error
		chain, err = OpenNew(ctx, c, dbKeyMR, eb)
		if err != nil {
			return err
		}

		log.Infof("Tracking new FAT chain: %v", chain.ID)

		// Fully sync the chain, so that it is up to date immediately.
		if err := chain.Sync(ctx, c); err != nil {
			return err
		}

		// Save the chain back into the map.
		Chains.set(chain.ID, chain, ChainStatusUnknown)
		return nil
	}

	// Ignore EBlocks earlier than the chain's current sync height.
	if eb.Height <= chain.Head.Height {
		return nil
	}

	// Rollback any pending entries on the chain.
	if chain.Pending.Entries != nil {
		// Load any cached entries that are pending and remove them
		// from the cache.
		for i := range eb.Entries {
			e := &eb.Entries[i]

			// Check if this entry is cached.
			cachedE, ok := chain.Pending.Entries[*e.Hash]
			if !ok {
				continue
			}

			// Use official Timestamp established by EBlock.
			cachedE.Timestamp = e.Timestamp
			*e = cachedE
		}

		err := chain.revertPending()
		// We must save the Chain back to the map at this point to
		// avoid a double free panic in the event of any further
		// errors.
		Chains.set(chain.ID, chain, chain.ChainStatus)
		if err != nil {
			return err
		}
	}

	// prevStatus saves the initial ChainStatus so we can detect if the
	// chain goes from Tracked to Issued.
	prevStatus := chain.ChainStatus

	// Apply this EBlock to the chain.
	chain.Log.Debugf("Applying EBlock %v...", eb.KeyMR)
	if err := chain.Apply(ctx, c, dbKeyMR, eb); err != nil {
		return err
	}
	if err := sqlitex.ExecScript(chain.Conn,
		`PRAGMA main.wal_checkpoint;`); err != nil {
		chain.Log.Error(err)
	}

	// Save the chain back into the map.
	Chains.set(chain.ID, chain, prevStatus)

	return nil
}

func ProcessPending(ctx context.Context, es ...factom.Entry) error {
	chain := Chains.get(es[0].ChainID)

	// We can only apply pending entries to tracked chains.
	if !chain.IsTracked() {
		return nil
	}

	// Initialize Pending if Entries is not yet populated.
	if chain.Pending.Entries == nil {
		if err := chain.initPending(ctx); err != nil {
			return err
		}
		// Ensure the chain is saved back into the map.
		Chains.set(chain.ID, chain, chain.ChainStatus)
	}

	// startLenEntries tracks the initial size of our cache so we can
	// detect if any new pending entries get applied.
	startLenEntries := len(chain.Pending.Entries)

	// Apply any new pending entries.
	for _, e := range es {
		// Ignore entries we have seen before.
		if _, ok := chain.Pending.Entries[*e.Hash]; ok {
			continue
		}

		// Load the Entry data.
		if err := e.Get(ctx, c); err != nil {
			return err
		}

		// The timestamp won't be established until the next EBlock so
		// use the current time for now.
		e.Timestamp = time.Now()

		if _, err := chain.Chain.ApplyEntry(e); err != nil {
			return err
		}

		// Cache the entry.
		chain.Pending.Entries[*e.Hash] = e
	}

	// Check if any no new entries were added.
	if startLenEntries == len(chain.Pending.Entries) {
		return nil
	}

	chain.Log.Debugf("Applied %v new pending entries.",
		len(chain.Pending.Entries)-startLenEntries)

	// Save the chain back into the map.
	Chains.set(chain.ID, chain, chain.ChainStatus)
	return nil
}
func (chain *Chain) initPending(ctx context.Context) (err error) {
	chain.Log.Debug("Initializing pending...")

	s, err := chain.Pool.GetSnapshot(ctx)
	if err != nil {
		return
	}

	// Start a new session so we can track all changes and later rollback
	// all pending entries.
	session, err := chain.Conn.CreateSession("")
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			session.Delete()
		}
	}()
	if err := session.Attach(""); err != nil {
		return err
	}

	// There is a chance the Identity is populated now but wasn't before,
	// so update it now.
	if err := chain.Identity.Get(ctx, c); err != nil {
		// A jsonrpc2.Error indicates that the identity chain doesn't
		// yet exist, which we tolerate.
		if _, ok := err.(jsonrpc2.Error); !ok {
			return err
		}
	}

	chain.Pending.Entries = make(map[factom.Bytes32]factom.Entry)
	chain.Pending.OfficialChain = chain.Chain
	chain.Pending.Session = session
	chain.Pending.OfficialSnapshot = s

	return nil
}
func (chain *Chain) revertPending() error {
	chain.Log.Debug("Cleaning up pending state...")
	// We must clear the interrupt to prevent from panicking or being
	// interrupted while reverting.
	oldDone := chain.Conn.SetInterrupt(nil)
	defer func() {
		chain.Pending.Entries = nil
		// Always clean up our session and snapshots.
		chain.Pending.OfficialSnapshot = nil
		chain.Pending.OfficialChain = db.Chain{}

		chain.Pending.Session.Delete()
		chain.Pending.Session = nil
		chain.Conn.SetInterrupt(oldDone)

	}()
	// Revert all of the pending transactions by applying the inverse of
	// the changeset tracked by the session.
	var changeset bytes.Buffer
	if err := chain.Pending.Session.Changeset(&changeset); err != nil {
		return fmt.Errorf("chain.Pending.Session.Changeset(): %w", err)
	}
	inverse := bytes.NewBuffer(make([]byte, 0, changeset.Len()))
	if err := sqlite.ChangesetInvert(inverse, &changeset); err != nil {
		return fmt.Errorf("sqlite.ChangesetInvert(): %w", err)
	}
	if err := chain.Conn.ChangesetApply(inverse, nil, chain.conflictFn); err != nil {
		return fmt.Errorf("chain.Conn.ChangesetApply(): %w", err)

	}
	return nil
}

// Get returns a threadsafe connection to the database, and a function to
// release the connection back to the pool. If pending is true, the chain will
// reflect the state with pending entries applied. Otherwise the chain will
// reflect the official state after the most recent EBlock.
func (cm *ChainMap) Get(ctx context.Context,
	id *factom.Bytes32, pending bool) (Chain, func()) {

	chain := cm.get(id)

	// Pull a Conn off the Pool and set it as the main Conn.
	conn := chain.Pool.Get(ctx)
	if conn == nil {
		return Chain{}, nil
	}
	chain.Conn = conn

	// If pending or if there is no pending state, then use the chain as
	// is, and just return a function that returns the conn to the pool.
	if pending || chain.Pending.Entries == nil {
		return chain, func() {
			chain.Pool.Put(conn)
		}
	}
	// There are pending entries, but we have been asked for the official
	// state.

	// Start a read transaction on the conn that reflects the official
	// state.
	endRead, err := conn.StartSnapshotRead(chain.Pending.OfficialSnapshot)
	if err != nil {
		chain.Pool.Put(conn)
		panic(err)
	}

	// Use the official chain state with the conn from the Pool.
	chain.Chain = chain.Pending.OfficialChain
	chain.Conn = conn

	// Return a function that ends the read transaction and returns the
	// conn to the Pool.
	return chain, func() {
		// We must clear the interrupt to prevent endRead from
		// panicking.
		conn.SetInterrupt(nil)
		endRead()
		chain.Pool.Put(conn)
	}
}
