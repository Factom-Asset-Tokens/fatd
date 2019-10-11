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

package fat0

import (
	"encoding/json"
	"fmt"

	"github.com/Factom-Asset-Tokens/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat"
	"github.com/Factom-Asset-Tokens/fatd/fat/jsonlen"
)

const Type = fat.TypeFAT0

// Transaction represents a fat0 transaction, which can be a normal account
// transaction or a coinbase transaction depending on the Inputs and the
// RCD/signature pair.
type Transaction struct {
	Inputs  AddressAmountMap `json:"inputs"`
	Outputs AddressAmountMap `json:"outputs"`
	fat.Entry
}

var _ fat.Transaction = &Transaction{}

// NewTransaction returns a Transaction initialized with the given entry.
func NewTransaction(entry factom.Entry) *Transaction {
	return &Transaction{Entry: fat.Entry{Entry: entry}}
}

func (t *Transaction) UnmarshalJSON(data []byte) error {
	data = jsonlen.Compact(data)
	tRaw := struct {
		Inputs  json.RawMessage `json:"inputs"`
		Outputs json.RawMessage `json:"outputs"`
		fat.Entry
	}{}
	if err := json.Unmarshal(data, &tRaw); err != nil {
		return fmt.Errorf("%T: %w", t, err)
	}
	if err := t.Inputs.UnmarshalJSON(tRaw.Inputs); err != nil {
		return fmt.Errorf("%T.Inputs: %w", t, err)
	}
	if err := t.Outputs.UnmarshalJSON(tRaw.Outputs); err != nil {
		return fmt.Errorf("%T.Outputs: %w", t, err)
	}
	t.Metadata = tRaw.Metadata

	if err := t.ValidData(); err != nil {
		return fmt.Errorf("%T: %w", t, err)
	}

	expectedJSONLen := len(`{"inputs":,"outputs":}`) +
		len(tRaw.Inputs) + len(tRaw.Outputs) +
		tRaw.MetadataJSONLen()
	if expectedJSONLen != len(data) {
		return fmt.Errorf("%T: unexpected JSON length", t)
	}

	return nil
}

type transaction Transaction

func (t Transaction) MarshalJSON() ([]byte, error) {
	if err := t.ValidData(); err != nil {
		return nil, err
	}
	return json.Marshal(transaction(t))
}

func (t Transaction) String() string {
	data, err := t.MarshalJSON()
	if err != nil {
		return err.Error()
	}
	return string(data)
}

// UnmarshalEntry unmarshals the entry content as a Transaction.
func (t *Transaction) UnmarshalEntry() error {
	return t.Entry.UnmarshalEntry(t)
}

// MarshalEntry marshals the Transaction into the Entry content.
func (t *Transaction) MarshalEntry() error {
	return t.Entry.MarshalEntry(t)
}

// IsCoinbase returns true if the coinbase address is in t.Input. This does not
// necessarily mean that t is a valid coinbase transaction.
func (t Transaction) IsCoinbase() bool {
	amount := t.Inputs[fat.Coinbase()]
	return amount != 0
}

// Validate performs all validation checks and returns nil if t is a valid
// Transaction. If t is a coinbase transaction then idKey is used to validate
// the RCD. Otherwise RCDs are checked against the input addresses.
func (t *Transaction) Validate(idKey *factom.ID1Key) error {
	if err := t.UnmarshalEntry(); err != nil {
		return err
	}
	if err := t.ValidExtIDs(); err != nil {
		return err
	}
	if t.IsCoinbase() {
		if t.ID1Key() != *idKey {
			return fmt.Errorf("invalid RCD")
		}
	} else {
		if !t.ValidRCDs() {
			return fmt.Errorf("invalid RCDs")
		}
	}
	return nil
}

// ValidData validates the Transaction data and returns nil if no errors are
// present. ValidData assumes that the entry content has been unmarshaled.
func (t Transaction) ValidData() error {
	if t.Inputs.Sum() != t.Outputs.Sum() {
		return fmt.Errorf("sum(inputs) != sum(outputs)")
	}
	// Coinbase transactions must only have one input.
	if t.IsCoinbase() && len(t.Inputs) != 1 {
		return fmt.Errorf("invalid coinbase transaction")
	}
	return nil
}

// ValidExtIDs validates the structure of the external IDs of the entry to make
// sure that it has the correct number of RCD/signature pairs. ValidExtIDs does
// not validate the content of the RCD or signature. ValidExtIDs assumes that
// the entry content has been unmarshaled and that ValidData returns nil.
func (t Transaction) ValidExtIDs() error {
	return t.Entry.ValidExtIDs(len(t.Inputs))
}

func (t Transaction) ValidRCDs() bool {
	// Create a map of all RCDs that are present in the ExtIDs.
	rcdHashes := make(map[factom.FAAddress]struct{}, len(t.Inputs))
	extIDs := t.ExtIDs[1:]
	for i := 0; i < len(extIDs)/2; i++ {
		rcdHashes[t.FAAddress(i)] = struct{}{}
	}

	// Ensure that for all Inputs there is a corresponding RCD in the
	// ExtIDs.
	for rcdHash := range t.Inputs {
		if _, ok := rcdHashes[rcdHash]; !ok {
			return false
		}
	}
	return true
}
