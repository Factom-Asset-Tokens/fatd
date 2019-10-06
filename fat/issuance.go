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

package fat

import (
	"encoding/json"
	"fmt"

	"github.com/Factom-Asset-Tokens/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat/jsonlen"
)

var (
	coinbase = func() factom.FAAddress {
		priv := factom.FsAddress{}
		return priv.FAAddress()
	}()
)

func Coinbase() factom.FAAddress {
	return coinbase
}

// Issuance represents the Issuance of a token.
type Issuance struct {
	Type   Type  `json:"type"`
	Supply int64 `json:"supply"`

	Symbol string `json:"symbol,omitempty"`
	Entry
}

type issuance Issuance

func (i *Issuance) UnmarshalJSON(data []byte) error {
	data = jsonlen.Compact(data)
	if err := json.Unmarshal(data, (*issuance)(i)); err != nil {
		return fmt.Errorf("%T: %v", i, err)
	}
	if err := i.ValidData(); err != nil {
		return fmt.Errorf("%T: %v", i, err)
	}
	if i.expectedJSONLength() != len(data) {
		return fmt.Errorf("%T: unexpected JSON length", i)
	}
	return nil
}
func (i Issuance) expectedJSONLength() int {
	l := len(`{}`)
	l += len(`"type":""`) + len(i.Type.String())
	l += len(`,"supply":`) + jsonlen.Int64(i.Supply)
	l += jsonStrLen("symbol", i.Symbol)
	l += i.MetadataJSONLen()
	return l
}
func jsonStrLen(name, value string) int {
	if len(value) == 0 {
		return 0
	}
	return len(`,"":""`) + len(name) + len(value)
}

func (i Issuance) MarshalJSON() ([]byte, error) {
	if err := i.ValidData(); err != nil {
		return nil, err
	}
	return json.Marshal(issuance(i))
}

// NewIssuance returns an Issuance initialized with the given entry.
func NewIssuance(entry factom.Entry) Issuance {
	return Issuance{Entry: Entry{Entry: entry}}
}

// UnmarshalEntry unmarshals the entry content as an Issuance.
func (i *Issuance) UnmarshalEntry() error {
	return i.Entry.UnmarshalEntry(i)
}

// MarshalEntry marshals the entry content as an Issuance.
func (i *Issuance) MarshalEntry() error {
	return i.Entry.MarshalEntry(i)
}

// Validate performs all validation checks and returns nil if i is a valid
// Issuance.
func (i *Issuance) Validate(idKey *factom.ID1Key) error {
	if err := i.UnmarshalEntry(); err != nil {
		return err
	}
	if err := i.ValidExtIDs(); err != nil {
		return err
	}
	if i.ID1Key() != *idKey {
		return fmt.Errorf("invalid RCD")
	}
	return nil
}

// ValidData validates the Issuance data and returns nil if no errors are
// present. ValidData assumes that the entry content has been unmarshaled.
func (i Issuance) ValidData() error {
	if !i.Type.IsValid() {
		return fmt.Errorf(`invalid "type": %v`, i.Type)
	}
	if i.Supply == 0 || i.Supply < -1 {
		return fmt.Errorf(`invalid "supply": must be positive or -1`)
	}
	return nil
}

// ValidExtIDs validates the structure of the external IDs of the entry to make
// sure that it has an RCD and signature. It does not validate the content of
// the RCD or signature.
func (i Issuance) ValidExtIDs() error {
	return i.Entry.ValidExtIDs(1)
}
