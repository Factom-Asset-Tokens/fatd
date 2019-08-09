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
	"github.com/Factom-Asset-Tokens/fatd/factom"
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
		var err error
		chain, err = OpenNew(c, dbKeyMR, eb)
		if err != nil {
			return err
		}
		if chain.IsUnknown() {
			Chains.ignore(eb.ChainID)
			return nil
		}
		if err := chain.Sync(c); err != nil {
			return err
		}
		Chains.set(chain.ID, chain, ChainStatusUnknown)
		return nil
	}
	if eb.Height <= chain.Head.Height {
		return nil
	}
	prevStatus := chain.ChainStatus
	if err := chain.Apply(c, dbKeyMR, eb); err != nil {
		return err
	}
	Chains.set(chain.ID, chain, prevStatus)
	return nil
}
