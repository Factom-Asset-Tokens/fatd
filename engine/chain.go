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

	jrpc "github.com/AdamSLevy/jsonrpc2/v11"
	"github.com/Factom-Asset-Tokens/fatd/db"
	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat"
)

type Chain struct {
	ChainStatus
	*db.Chain
}

func (chain Chain) String() string {
	return fmt.Sprintf("{ChainStatus:%v, ID:%v, "+
		"fat.Identity:%+v, fat.Issuance:%+v}",
		chain.ChainStatus, chain.ID,
		chain.Identity, chain.Issuance)
}

func (chain *Chain) ignore() {
	chain.ID = nil // Allow to be GC'd.
	chain.ChainStatus = ChainStatusIgnored
}
func (chain *Chain) track(first factom.EBlock, dbKeyMR *factom.Bytes32) error {
	return nil
}
func (chain *Chain) issue(issuance fat.Issuance) error {
	return nil
}

func (chain *Chain) Sync(c *factom.Client) error {
	eblocks, err := factom.EBlock{}.GetPrevUpTo(c, *chain.Head.KeyMR)
	if err != nil {
		return fmt.Errorf("factom.EBlock{}.GetPrevUpTo(): %v", err)
	}
	for i := range eblocks {
		eb := eblocks[len(eblocks)-1-i] // Earliest EBlock first.

		// Get DBlock Timestamp and KeyMR
		var dblock factom.DBlock
		dblock.Header.Height = eb.Height
		if err := dblock.Get(c); err != nil {
			return fmt.Errorf("factom.DBlock{}.Get(): %v", err)
		}
		eb.SetTimestamp(dblock.Header.Timestamp)

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
