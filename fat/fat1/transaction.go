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

package fat1

import (
	"encoding/json"
	"fmt"

	"github.com/Factom-Asset-Tokens/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat"
	"github.com/Factom-Asset-Tokens/fatd/fat/jsonlen"
)

const Type = fat.TypeFAT1

// Transaction represents a fat1 transaction, which can be a normal account
// transaction or a coinbase transaction depending on the Inputs and the
// RCD/signature pair.
type Transaction struct {
	Inputs        AddressNFTokensMap   `json:"inputs"`
	Outputs       AddressNFTokensMap   `json:"outputs"`
	TokenMetadata NFTokenIDMetadataMap `json:"tokenmetadata,omitempty"`
	fat.Entry
}

var _ fat.Transaction = &Transaction{}

// NewTransaction returns a Transaction initialized with the given entry.
func NewTransaction(entry factom.Entry) *Transaction {
	return &Transaction{Entry: fat.Entry{Entry: entry}}
}

func (t *Transaction) UnmarshalJSON(data []byte) error {
	tRaw := struct {
		Inputs        json.RawMessage `json:"inputs"`
		Outputs       json.RawMessage `json:"outputs"`
		TokenMetadata json.RawMessage `json:"tokenmetadata"`
		fat.Entry
	}{}
	if err := json.Unmarshal(data, &tRaw); err != nil {
		return fmt.Errorf("%T: %v", t, err)
	}
	if err := t.Inputs.UnmarshalJSON(tRaw.Inputs); err != nil {
		return fmt.Errorf("%T.Inputs: %v", t, err)
	}
	var expectedJSONLen int
	if len(tRaw.TokenMetadata) > 0 {
		if !t.IsCoinbase() {
			return fmt.Errorf(`%T: %v`, t,
				`invalid field for non-coinbase transaction: "tokenmetadata"`)
		}
		if err := t.TokenMetadata.UnmarshalJSON(tRaw.TokenMetadata); err != nil {
			return fmt.Errorf("%T.TokenMetadata: %v", t, err)

		}
		if err := t.TokenMetadata.IsSubsetOf(t.Inputs[fat.Coinbase()]); err != nil {
			return fmt.Errorf("%T.TokenMetadata: %v", t, err)
		}

		expectedJSONLen = len(`,"tokenmetadata":`) +
			len(jsonlen.Compact(tRaw.TokenMetadata))
	} else {
		if t.IsCoinbase() {
			// Avoid a nil map.
			t.TokenMetadata = make(NFTokenIDMetadataMap, 0)
		}
	}

	if err := t.Outputs.UnmarshalJSON(tRaw.Outputs); err != nil {
		return fmt.Errorf("%T.Outputs: %v", t, err)
	}
	t.Metadata = tRaw.Metadata

	if err := t.ValidData(); err != nil {
		return fmt.Errorf("%T: %v", t, err)
	}

	expectedJSONLen += len(`{"inputs":,"outputs":}`) +
		len(jsonlen.Compact(tRaw.Inputs)) + len(jsonlen.Compact(tRaw.Outputs)) +
		tRaw.MetadataJSONLen()
	if expectedJSONLen != len(jsonlen.Compact(data)) {
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

func (t Transaction) ValidData() error {
	if err := t.Inputs.NoAddressIntersection(t.Outputs); err != nil {
		return fmt.Errorf("Inputs and Outputs intersect: %v", err)
	}
	if err := t.Inputs.NFTokenIDsConserved(t.Outputs); err != nil {
		return fmt.Errorf("Inputs and Outputs mismatch: %v", err)
	}
	// Coinbase transactions must only have one input.
	if t.IsCoinbase() && len(t.Inputs) != 1 {
		return fmt.Errorf("invalid coinbase transaction")
	}
	return nil
}

// IsCoinbase returns true if the coinbase address is in t.Input. This does not
// necessarily mean that t is a valid coinbase transaction.
func (t Transaction) IsCoinbase() bool {
	tkns := t.Inputs[fat.Coinbase()]
	return len(tkns) != 0
}

// UnmarshalEntry unmarshals the entry content as a Transaction.
func (t *Transaction) UnmarshalEntry() error {
	return t.Entry.UnmarshalEntry(t)
}

// MarshalEntry marshals the Transaction into the Entry content.
func (t *Transaction) MarshalEntry() error {
	return t.Entry.MarshalEntry(t)
}

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

func (t Transaction) ValidExtIDs() error {
	return t.Entry.ValidExtIDs(len(t.Inputs))
}

func (t Transaction) ValidRCDs() bool {
	// Create a map of all RCDs that are present in the ExtIDs.
	adrs := make(map[factom.FAAddress]struct{}, len(t.Inputs))
	extIDs := t.ExtIDs[1:]
	for i := 0; i < len(extIDs)/2; i++ {
		adrs[t.FAAddress(i)] = struct{}{}
	}

	// Ensure that for all Inputs there is a corresponding RCD in the
	// ExtIDs.
	for inputAdr := range t.Inputs {
		if _, ok := adrs[inputAdr]; !ok {
			return false
		}
	}
	return true
}
