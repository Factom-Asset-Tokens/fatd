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
	"github.com/Factom-Asset-Tokens/fatd/internal/db"
	"github.com/Factom-Asset-Tokens/fatd/internal/db/eblock"
	"github.com/Factom-Asset-Tokens/fatd/internal/db/entry"
	"github.com/Factom-Asset-Tokens/fatd/internal/db/metadata"
)

type FactomChain db.FactomChain

func (chain *FactomChain) UpdateSidechainData(context.Context, *factom.Client) error {
	return nil
}

func (chain *FactomChain) ApplyEBlock(dbKeyMR *factom.Bytes32, eb factom.EBlock) error {
	// Insert latest EBlock.
	if err := eblock.Insert(chain.Conn, eb, dbKeyMR); err != nil {
		return fmt.Errorf("eblock.Insert(): %w", err)
	}
	if err := chain.SetSync(eb.Height, dbKeyMR); err != nil {
		return fmt.Errorf("state.FactomChain.SetSync(): %w", err)
	}
	chain.Head = eb
	return nil
}

func (chain *FactomChain) ApplyEntry(e factom.Entry) (eID int64, err error) {
	return entry.Insert(chain.Conn, e, chain.Head.Sequence)
}

func (chain *FactomChain) SetSync(height uint32, dbKeyMR *factom.Bytes32) error {
	if height <= chain.SyncHeight {
		return nil
	}
	if err := metadata.SetSync(chain.Conn, height, dbKeyMR); err != nil {
		return err
	}
	chain.SyncHeight = height
	chain.SyncDBKeyMR = dbKeyMR
	return nil
}

func ToFactomChain(chain Chain) (factomChain *FactomChain, ok bool) {
	factomChain, ok = chain.(*FactomChain)
	return
}

func (chain *FactomChain) ToFactomChain() *db.FactomChain {
	return (*db.FactomChain)(chain)
}

func (chain *FactomChain) Close() error {
	return chain.ToFactomChain().Close()
}
