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

	"github.com/Factom-Asset-Tokens/factom"
	"github.com/Factom-Asset-Tokens/fatd/internal/db"
)

// Chain is the interface for advancing a Chain's state. This is used by the
// engine to apply EBlocks and Entries.
type Chain interface {
	// UpdateSidechainData updates any data from external Chains that the
	// state depends on. This should be called before ApplyEBlock.
	UpdateSidechainData(context.Context) error

	// Apply applies the next EBlock to the chain state.
	ApplyEBlock(*factom.Bytes32, factom.EBlock) error

	// ApplyEntry applies the next Entry. This is used by the engine for
	// applying pending entries.
	ApplyEntry(context.Context, factom.Entry) (id int64, err error)

	SetSync(uint32, *factom.Bytes32) error

	// Copy returns a copy of the current state. This allows the engine to
	// save and rollback the in-memory state data.
	Copy() Chain

	Save() func(*error)

	// ToFactomChain returns a pointer to the underlying db.FactomChain
	// that all chains embed.
	ToFactomChain() *db.FactomChain

	Close() error
}

type UnknownChain struct{}

var _ Chain = UnknownChain{}

func (chain UnknownChain) UpdateSidechainData(context.Context) error {
	panic("UnknownChain should not be used")
}
func (chain UnknownChain) ApplyEBlock(*factom.Bytes32, factom.EBlock) error {
	panic("UnknownChain should not be used")
}
func (chain UnknownChain) ApplyEntry(context.Context, factom.Entry) (int64, error) {
	panic("UnknownChain should not be used")
}
func (chain UnknownChain) ToFactomChain() *db.FactomChain {
	panic("UnknownChain should not be used")
}
func (chain UnknownChain) Save() func(*error) {
	panic("UnknownChain should not be used")
}
func (chain UnknownChain) SetSync(uint32, *factom.Bytes32) error {
	panic("IgnoreChain should not be used")
}
func (chain UnknownChain) Copy() Chain {
	panic("UnknownChain should not be used")
}
func (chain UnknownChain) Close() error {
	panic("UnknownChain should not be used")
}
