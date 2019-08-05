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
	"github.com/Factom-Asset-Tokens/fatd/flag"
)

func Process(dbKeyMR *factom.Bytes32, eb factom.EBlock) error {
	// Skip ignored chains or EBlocks for heights earlier than this chain's
	// head height.
	chain := Chains.Get(eb.ChainID)
	if chain.IsIgnored() {
		return nil
	}
	if chain.IsUnknown() {
		if flag.IgnoreNewChains() {
			Chains.ignore(eb.ChainID)
			return nil
		}
		// Load this Entry Block.
		if err := eb.Get(c); err != nil {
			return fmt.Errorf("%#v.Get(c): %v", eb, err)
		}
		if !eb.IsFirst() {
			Chains.ignore(eb.ChainID)
			return nil
		}
		// Load first entry of new chain.
		first := eb.Entries[0]
		if err := first.Get(c); err != nil {
			return fmt.Errorf("%#v.Get(c): %v", first, err)
		}

		// Ignore chains with NameIDs that don't match the fat pattern.
		nameIDs := first.ExtIDs
		if !fat.ValidTokenNameIDs(nameIDs) {
			Chains.ignore(eb.ChainID)
			return nil
		}

		var identity factom.Identity
		_, identity.ChainID = fat.TokenIssuer(nameIDs)
		if err := identity.Get(c); err != nil {
			// A jrpc.Error indicates that the identity chain
			// doesn't yet exist, which we tolerate.
			if _, ok := err.(jrpc.Error); !ok {
				return err
			}
		}

		if err := eb.GetEntries(c); err != nil {
			return fmt.Errorf("%#v.GetEntries(c): %v", eb, err)
		}

		var err error
		chain.Chain, err = db.OpenNew(dbKeyMR, eb, flag.NetworkID,
			identity)
		if err != nil {
			return fmt.Errorf("db.OpenNew(): %v")
		}
		if chain.Issuance.IsPopulated() {
			chain.ChainStatus = ChainStatusIssued
		} else {
			chain.ChainStatus = ChainStatusTracked
		}
		Chains.set(chain.ID, chain)
		return nil
	}
	if eb.Height <= chain.Head.Height {
		return nil
	}
	return chain.Apply(c, dbKeyMR, eb)
}
