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
	"time"

	jrpc "github.com/AdamSLevy/jsonrpc2/v11"
	"github.com/Factom-Asset-Tokens/factom"
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
	if chain.Pending.Entries != nil {
		// Load any cached entries that are pending.
		for i := range eb.Entries {
			e := &eb.Entries[i]
			cached, ok := chain.Pending.Entries[*e.Hash]
			if !ok {
				continue
			}
			// Save Timestamp established by EBlock
			cached.Timestamp = e.Timestamp
			*e = cached
		}
		if err := chain.Pending.Sync(chain.Chain); err != nil {
			return err
		}
		session, err := chain.Conn.CreateSession("")
		if err != nil {
			return err
		}
		if err := session.Attach(""); err != nil {
			return err
		}
		chain.Pending.MainSess = session
	}
	prevStatus := chain.ChainStatus
	if err := chain.Apply(c, dbKeyMR, eb); err != nil {
		return err
	}
	Chains.set(chain.ID, chain, prevStatus)
	return nil
}

func ProcessPending(es ...factom.Entry) error {
	e := es[0] // Deliberately panic if we are called with no entries.
	if e.ChainID == nil {
		return nil
	}
	chain := Chains.Get(e.ChainID)
	// We can only apply pending entries to tracked chains.
	if !chain.IsTracked() {
		return nil
	}

	if chain.Pending.Entries == nil {
		if chain.Pending.Chain.Conn == nil {
			if err := chain.Pending.Open(chain.Chain); err != nil {
				return err
			}
		} else {
			if err := chain.Pending.Sync(chain.Chain); err != nil {
				return err
			}
		}

		chain.Pending.Entries = make(map[factom.Bytes32]factom.Entry)

		session, err := chain.Pending.Chain.Conn.CreateSession("")
		if err != nil {
			return err
		}
		if err := session.Attach(""); err != nil {
			return err
		}
		chain.Pending.PendSess = session

		// Small chance Identity is populated now but wasn't before...
		if err := chain.Identity.Get(c); err != nil {
			// A jrpc.Error indicates that the identity chain doesn't yet
			// exist, which we tolerate.
			if _, ok := err.(jrpc.Error); !ok {
				return err
			}
		}
	}

	lenEntries := len(chain.Pending.Entries)
	for _, e := range es {
		if _, ok := chain.Pending.Entries[*e.Hash]; ok {
			// Ignore entries we have seen before.
			continue
		}
		if err := e.Get(c); err != nil {
			return err
		}
		// The timestamp will later be established by the next EBlock so use
		// the current time for now. This is the time we first saw the pending
		// entry.
		e.Timestamp = time.Now()
		if _, err := chain.Pending.Chain.ApplyEntry(e, factom.EBlock{}, -1); err != nil {
			return err
		}
		chain.Pending.Entries[*e.Hash] = e // Cache the entry in memory.
	}
	if lenEntries == len(chain.Pending.Entries) {
		// No new entries
		return nil
	}
	Chains.set(chain.ID, chain, chain.ChainStatus)
	return nil
}

// ApplyPending applies any pending txs to conn.
func (chain *Chain) ApplyPending() {
	if chain.Pending.Chain.Conn == nil || chain.Pending.MainSess != nil {
		return
	}
	chain.Chain = chain.Pending.Chain
}
