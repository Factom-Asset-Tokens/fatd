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
	"time"

	"crawshaw.io/sqlite"
	jsonrpc2 "github.com/AdamSLevy/jsonrpc2/v12"
	"github.com/Factom-Asset-Tokens/factom"
	"github.com/Factom-Asset-Tokens/fatd/flag"
)

func Process(ctx context.Context, dbKeyMR *factom.Bytes32, eb factom.EBlock) error {
	chain := Chains.Get(eb.ChainID)

	// Skip ignored chains and if we are ignoring new chain, also skip
	// unknown chains.
	if chain.IsIgnored() ||
		(flag.IgnoreNewChains() && chain.IsUnknown()) {
		return nil
	}
	if chain.IsUnknown() {
		// Attempt to open a new chain.
		var err error
		chain, err = OpenNew(ctx, c, dbKeyMR, eb)
		if err != nil {
			return err
		}

		// Ignore if the chain was not opened.
		if chain.IsUnknown() {
			Chains.ignore(eb.ChainID)
			return nil
		}

		// Fully sync the chain, so that it is up to date immediately.
		if err := chain.Sync(c); err != nil {
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
	if chain.Pending.OfficialSnapshot != nil {
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

			// Delete the entry from the pending cache, rather than
			// nil the entire map, because this allows the cache
			// for various chains to grow according to how active
			// they are. But this is not optimal for chains with
			// only very occasional bursts of pending entries.
			//
			// A mechanism to eventually free rarely used pending
			// entry maps would be a good improvement later.
			delete(chain.Pending.Entries, *e.Hash)
		}

		// Revert all of the pending transactions by applying the
		// inverse of the changeset tracked by the session.
		changeset := &bytes.Buffer{}
		if err := chain.Pending.Session.Changeset(changeset); err != nil {
			return err
		}
		inverse := bytes.NewBuffer(make([]byte, 0, changeset.Cap()))
		if err := sqlite.ChangesetInvert(inverse, changeset); err != nil {
			return err
		}
		if err := chain.Conn.ChangesetApply(
			inverse, nil, chain.conflictFn); err != nil {
			return err
		}

		// Clean up.
		chain.Pending.Session.Delete()
		chain.Pending.Session = nil

		chain.Pending.OfficialSnapshot.Free()
		chain.Pending.OfficialSnapshot = nil
	}

	// prevStatus saves the initial ChainStatus so we can detect if the
	// chain goes from Tracked to Issued.
	prevStatus := chain.ChainStatus

	// Apply this EBlock to the chain.
	if err := chain.Apply(c, dbKeyMR, eb); err != nil {
		return err
	}

	// Save the chain back into the map.
	Chains.set(chain.ID, chain, prevStatus)
	return nil
}
func (chain Chain) conflictFn(
	cType sqlite.ConflictType, _ sqlite.ChangesetIter) sqlite.ConflictAction {
	chain.Log.Errorf("ChangesetApply Conflict: %v", cType)
	return sqlite.SQLITE_CHANGESET_ABORT
}

func ProcessPending(es ...factom.Entry) error {
	chain := Chains.Get(es[0].ChainID)

	// We can only apply pending entries to tracked chains.
	if !chain.IsTracked() {
		return nil
	}

	// Initialize Pending if there is no snapshot yet.
	if chain.Pending.OfficialSnapshot == nil {
		// Create the cache if it does not exist.
		if chain.Pending.Entries == nil {
			chain.Pending.Entries = make(map[factom.Bytes32]factom.Entry)
		}

		// Take a snapshot of the official state and copy the current
		// official Chain.
		s, err := chain.Conn.CreateSnapshot("")
		if err != nil {
			return err
		}
		chain.Pending.OfficialSnapshot = s
		chain.Pending.OfficialChain = chain.Chain

		// Start a new session so we can track all changes and later
		// rollback all pending entries.
		session, err := chain.Conn.CreateSession("")
		if err != nil {
			return err
		}
		if err := session.Attach(""); err != nil {
			return err
		}
		chain.Pending.Session = session

		// There is a chance the Identity is populated now but wasn't
		// before, so update it now.
		if err := chain.Identity.Get(context.TODO(), c); err != nil {
			// A jsonrpc2.Error indicates that the identity chain
			// doesn't yet exist, which we tolerate.
			if _, ok := err.(jsonrpc2.Error); !ok {
				return err
			}
		}
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
		if err := e.Get(context.TODO(), c); err != nil {
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

	// Save the chain back into the map.
	Chains.set(chain.ID, chain, chain.ChainStatus)
	return nil
}

// Get returns a threadsafe connection to the database, and a function to
// release the connection back to the pool. If pending is true, the chain will
// reflect the state with pending entries applied. Otherwise the chain will
// reflect the official state after the most recent EBlock.
func (chain *Chain) Get(pending bool) func() {
	// Pull a Conn off the Pool and set it as the main Conn.
	conn := chain.Pool.Get(nil)
	chain.Conn = conn

	// If pending or if there is no pending state, then use the chain as
	// is, and just return a function that returns the conn to the pool.
	if pending || chain.Pending.OfficialSnapshot == nil {
		return func() { chain.Pool.Put(conn) }
	}

	// Use the official chain state with the conn from the Pool.
	chain.Chain = chain.Pending.OfficialChain
	chain.Conn = conn

	// Start a read transaction on the conn that reflects the official
	// state.
	endRead, err := conn.StartSnapshotRead(chain.Pending.OfficialSnapshot)
	if err != nil {
		panic(err)
	}

	// Return a function that ends the read transaction and returns the
	// conn to the Pool.
	return func() { endRead(); chain.Pool.Put(conn) }
}
