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
	"fmt"

	"crawshaw.io/sqlite"

	jrpc "github.com/AdamSLevy/jsonrpc2/v11"

	"github.com/Factom-Asset-Tokens/factom"
	"github.com/Factom-Asset-Tokens/fatd/db"
	"github.com/Factom-Asset-Tokens/fatd/fat"
	"github.com/Factom-Asset-Tokens/fatd/flag"
)

type Chain struct {
	ChainStatus
	db.Chain

	Pending Pending
}

type Pending struct {
	Session          *sqlite.Session
	OfficialSnapshot *sqlite.Snapshot
	OfficialChain    db.Chain
	Entries          map[factom.Bytes32]factom.Entry
}

func (chain Chain) String() string {
	return fmt.Sprintf("{ChainStatus:%v, ID:%v, "+
		"fat.Identity:%+v, fat.Issuance:%+v}",
		chain.ChainStatus, chain.ID,
		chain.Identity, chain.Issuance)
}

func OpenNew(c *factom.Client,
	dbKeyMR *factom.Bytes32, eb factom.EBlock) (chain Chain, err error) {
	if err := eb.Get(c); err != nil {
		return chain, fmt.Errorf("%#v.Get(c): %v", eb, err)
	}
	// Load first entry of new chain.
	first := &eb.Entries[0]
	if err := first.Get(c); err != nil {
		return chain, fmt.Errorf("%#v.Get(c): %v", first, err)
	}
	if !eb.IsFirst() {
		return
	}

	// Ignore chains with NameIDs that don't match the fat pattern.
	nameIDs := first.ExtIDs
	if !fat.ValidTokenNameIDs(nameIDs) {
		return
	}

	var identity factom.Identity
	identity.ChainID = new(factom.Bytes32)
	_, *identity.ChainID = fat.TokenIssuer(nameIDs)
	if err = identity.Get(c); err != nil {
		// A jrpc.Error indicates that the identity chain
		// doesn't yet exist, which we tolerate.
		if _, ok := err.(jrpc.Error); !ok {
			return
		}
	}

	if err := eb.GetEntries(c); err != nil {
		return chain, fmt.Errorf("%#v.GetEntries(c): %v", eb, err)
	}

	chain.Chain, err = db.OpenNew(flag.DBPath, dbKeyMR, eb, flag.NetworkID, identity)
	if err != nil {
		return chain, fmt.Errorf("db.OpenNew(): %v", err)
	}
	if chain.Issuance.IsPopulated() {
		chain.ChainStatus = ChainStatusIssued
	} else {
		chain.ChainStatus = ChainStatusTracked
	}
	return
}

func (chain *Chain) OpenNewByChainID(c *factom.Client, chainID *factom.Bytes32) error {
	eblocks, err := factom.EBlock{ChainID: chainID}.GetPrevAll(c)
	if err != nil {
		return fmt.Errorf("factom.EBlock{}.GetPrevAll(): %v", err)
	}

	first := eblocks[len(eblocks)-1]
	// Get DBlock Timestamp and KeyMR
	var dblock factom.DBlock
	dblock.Height = first.Height
	if err := dblock.Get(c); err != nil {
		return fmt.Errorf("factom.DBlock{}.Get(): %v", err)
	}
	first.SetTimestamp(dblock.Timestamp)

	*chain, err = OpenNew(c, dblock.KeyMR, first)
	if err != nil {
		return err
	}
	if chain.IsUnknown() {
		return fmt.Errorf("not a valid FAT chain: %v", chainID)
	}

	// We already applied the first EBlock. Sync the remaining.
	return chain.SyncEBlocks(c, eblocks[:len(eblocks)-1])
}

func (chain *Chain) Sync(c *factom.Client) error {
	eblocks, err := factom.EBlock{ChainID: chain.ID}.GetPrevUpTo(c, *chain.Head.KeyMR)
	if err != nil {
		return fmt.Errorf("factom.EBlock{}.GetPrevUpTo(): %v", err)
	}
	return chain.SyncEBlocks(c, eblocks)
}

func (chain *Chain) SyncEBlocks(c *factom.Client, ebs []factom.EBlock) error {
	for i := range ebs {
		eb := ebs[len(ebs)-1-i] // Earliest EBlock first.

		// Get DBlock Timestamp and KeyMR
		var dblock factom.DBlock
		dblock.Height = eb.Height
		if err := dblock.Get(c); err != nil {
			return fmt.Errorf("factom.DBlock{}.Get(): %v", err)
		}
		eb.SetTimestamp(dblock.Timestamp)

		if err := chain.Apply(c, dblock.KeyMR, eb); err != nil {
			return err
		}
	}
	return nil
}

func (chain *Chain) Apply(c *factom.Client,
	dbKeyMR *factom.Bytes32, eb factom.EBlock) error {
	// Get Identity each time in case it wasn't populated before.
	if err := chain.Identity.Get(c); err != nil {
		// A jrpc.Error indicates that the identity chain doesn't yet
		// exist, which we tolerate.
		if _, ok := err.(jrpc.Error); !ok {
			return err
		}
	}
	// Get all entry data.
	if err := eb.GetEntries(c); err != nil {
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
