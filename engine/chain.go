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
	"sync"

	"crawshaw.io/sqlite"

	jsonrpc2 "github.com/AdamSLevy/jsonrpc2/v12"

	"github.com/Factom-Asset-Tokens/factom"
	"github.com/Factom-Asset-Tokens/fatd/db"
	"github.com/Factom-Asset-Tokens/fatd/fat"
	"github.com/Factom-Asset-Tokens/fatd/flag"
)

type Chain struct {
	ChainStatus
	db.Chain

	*sync.RWMutex

	Pending Pending
}

type Pending struct {
	Session          *sqlite.Session
	OfficialSnapshot *sqlite.Snapshot
	EndSnapshotRead  func()
	OfficialChain    db.Chain
	Entries          map[factom.Bytes32]factom.Entry
}

func (chain Chain) String() string {
	return fmt.Sprintf("{ChainStatus:%v, ID:%v, "+
		"fat.Identity:%+v, fat.Issuance:%+v}",
		chain.ChainStatus, chain.ID,
		chain.Identity, chain.Issuance)
}

func OpenNew(ctx context.Context, c *factom.Client,
	dbKeyMR *factom.Bytes32, eb factom.EBlock) (chain Chain, err error) {
	if err := eb.Get(ctx, c); err != nil {
		return chain, fmt.Errorf("%#v.Get(): %w", eb, err)
	}
	// Load first entry of new chain.
	first := &eb.Entries[0]
	if err := first.Get(ctx, c); err != nil {
		return chain, fmt.Errorf("%#v.Get(): %w", first, err)
	}
	if !eb.IsFirst() {
		return
	}

	// Ignore chains with NameIDs that don't match the fat pattern.
	nameIDs := first.ExtIDs
	if !fat.ValidTokenNameIDs(nameIDs) {
		return
	}

	chain.RWMutex = new(sync.RWMutex)

	var identity factom.Identity
	identity.ChainID = new(factom.Bytes32)
	_, *identity.ChainID = fat.TokenIssuer(nameIDs)
	if err = identity.Get(ctx, c); err != nil {
		// A jsonrpc2.Error indicates that the identity chain
		// doesn't yet exist, which we tolerate.
		if _, ok := err.(jsonrpc2.Error); !ok {
			return
		}
	}

	if err := eb.GetEntries(ctx, c); err != nil {
		return chain, fmt.Errorf("%#v.GetEntries(): %w", eb, err)
	}

	chain.Chain, err = db.OpenNew(ctx,
		flag.DBPath, dbKeyMR, eb, flag.NetworkID, identity)
	if err != nil {
		return chain, fmt.Errorf("db.OpenNew(): %w", err)
	}
	if chain.Issuance.IsPopulated() {
		chain.ChainStatus = ChainStatusIssued
	} else {
		chain.ChainStatus = ChainStatusTracked
	}
	return
}

func (chain *Chain) OpenNewByChainID(ctx context.Context,
	c *factom.Client, chainID *factom.Bytes32) error {
	eblocks, err := factom.EBlock{ChainID: chainID}.GetPrevAll(ctx, c)
	if err != nil {
		return fmt.Errorf("factom.EBlock.GetPrevAll(): %w", err)
	}

	first := eblocks[len(eblocks)-1]
	// Get DBlock Timestamp and KeyMR
	var dblock factom.DBlock
	dblock.Height = first.Height
	if err := dblock.Get(ctx, c); err != nil {
		return fmt.Errorf("factom.DBlock.Get(): %w", err)
	}
	first.SetTimestamp(dblock.Timestamp)

	*chain, err = OpenNew(ctx, c, dblock.KeyMR, first)
	if err != nil {
		return err
	}
	if chain.IsUnknown() {
		return fmt.Errorf("not a valid FAT chain: %v", chainID)
	}

	// We already applied the first EBlock. Sync the remaining.
	return chain.SyncEBlocks(ctx, c, eblocks[:len(eblocks)-1])
}

func (chain *Chain) Sync(ctx context.Context, c *factom.Client) error {
	eblocks, err := factom.EBlock{ChainID: chain.ID}.
		GetPrevUpTo(ctx, c, *chain.Head.KeyMR)
	if err != nil {
		return fmt.Errorf("factom.EBlock.GetPrevUpTo(): %w", err)
	}
	return chain.SyncEBlocks(ctx, c, eblocks)
}

func (chain *Chain) SyncEBlocks(ctx context.Context, c *factom.Client, ebs []factom.EBlock) error {
	for i := range ebs {
		eb := ebs[len(ebs)-1-i] // Earliest EBlock first.

		// Get DBlock Timestamp and KeyMR
		var dblock factom.DBlock
		dblock.Height = eb.Height
		if err := dblock.Get(ctx, c); err != nil {
			return fmt.Errorf("factom.DBlock.Get(): %w", err)
		}
		eb.SetTimestamp(dblock.Timestamp)

		if err := chain.Apply(ctx, c, dblock.KeyMR, eb); err != nil {
			return err
		}
	}
	return nil
}

func (chain *Chain) Apply(ctx context.Context, c *factom.Client,
	dbKeyMR *factom.Bytes32, eb factom.EBlock) error {
	// Get Identity each time in case it wasn't populated before.
	if err := chain.Identity.Get(ctx, c); err != nil {
		// A jsonrpc2.Error indicates that the identity chain doesn't yet
		// exist, which we tolerate.
		if _, ok := err.(jsonrpc2.Error); !ok {
			return err
		}
	}
	// Get all entry data.
	if err := eb.GetEntries(ctx, c); err != nil {
		return err
	}
	if err := chain.Chain.Apply(dbKeyMR, eb); err != nil {
		return err
	}
	// Update ChainStatus
	if !chain.IsIssued() && chain.Issuance.IsPopulated() {
		chain.ChainStatus = ChainStatusIssued
	}
	return nil
}

func (chain *Chain) conflictFn(
	cType sqlite.ConflictType, _ sqlite.ChangesetIter) sqlite.ConflictAction {
	chain.Log.Errorf("ChangesetApply Conflict: %v", cType)
	return sqlite.SQLITE_CHANGESET_ABORT
}
func (chain *Chain) revertPending() error {
	// We must clear the interrupt to prevent from panicking or being
	// interrupted while reverting.
	oldDone := chain.SetInterrupt(nil)
	defer func() {
		// Always clean up our session and snapshots.
		chain.Pending.EndSnapshotRead()
		chain.Pending.Session.Delete()
		chain.Pending.Session = nil
		chain.Pending.OfficialSnapshot.Free()
		chain.Pending.OfficialSnapshot = nil
		chain.SetInterrupt(oldDone)
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
