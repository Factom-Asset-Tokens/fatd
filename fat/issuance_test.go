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
	"testing"

	"github.com/Factom-Asset-Tokens/factom"
	"github.com/stretchr/testify/assert"
)

var coinbaseAddressStr = "FA1zT4aFpEvcnPqPCigB3fvGu4Q4mTXY22iiuV69DqE1pNhdF2MC"

func TestCoinbase(t *testing.T) {
	a := Coinbase()
	assert.Equal(t, coinbaseAddressStr, a.String())
}

var issuanceTests = []struct {
	Name      string
	Error     string
	IssuerKey factom.ID1Key
	Issuance
}{{
	Name:      "valid",
	IssuerKey: issuerKey,
	Issuance:  validIssuance(),
}, {
	Name:      "valid (omit symbol)",
	IssuerKey: issuerKey,
	Issuance:  omitFieldIssuance("symbol"),
}, {
	Name:      "valid (omit name)",
	IssuerKey: issuerKey,
	Issuance:  omitFieldIssuance("name"),
}, {
	Name:      "valid (omit metadata)",
	IssuerKey: issuerKey,
	Issuance:  omitFieldIssuance("metadata"),
}, {
	Name:      "invalid JSON (unknown field)",
	Error:     `*fat.Issuance: unexpected JSON length`,
	IssuerKey: issuerKey,
	Issuance:  setFieldIssuance("invalid", 5),
}, {
	Name:      "invalid JSON (invalid type)",
	Error:     `*fat.Issuance: *fat.Type: expected JSON string`,
	IssuerKey: issuerKey,
	Issuance:  invalidIssuance("type"),
}, {
	Name:      "invalid JSON (invalid supply)",
	Error:     `*fat.Issuance: json: cannot unmarshal array into Go struct field issuance.supply of type int64`,
	IssuerKey: issuerKey,
	Issuance:  invalidIssuance("supply"),
}, {
	Name:      "invalid JSON (invalid symbol)",
	Error:     `*fat.Issuance: json: cannot unmarshal array into Go struct field issuance.symbol of type string`,
	IssuerKey: issuerKey,
	Issuance:  invalidIssuance("symbol"),
}, {
	Name:      "invalid JSON (nil)",
	Error:     `unexpected end of JSON input`,
	IssuerKey: issuerKey,
	Issuance:  issuanceFromRaw(nil),
}, {
	Name:      "invalid data (type)",
	Error:     `*fat.Issuance: *fat.Type: invalid format`,
	IssuerKey: issuerKey,
	Issuance:  setFieldIssuance("type", "invalid"),
}, {
	Name:      "invalid data (type omitted)",
	Error:     `*fat.Issuance: unexpected JSON length`,
	IssuerKey: issuerKey,
	Issuance:  omitFieldIssuance("type"),
}, {
	Name:      "invalid data (supply: 0)",
	Error:     `*fat.Issuance: invalid "supply": must be positive or -1`,
	IssuerKey: issuerKey,
	Issuance:  setFieldIssuance("supply", 0),
}, {
	Name:      "invalid data (supply: -5)",
	Error:     `*fat.Issuance: invalid "supply": must be positive or -1`,
	IssuerKey: issuerKey,
	Issuance:  setFieldIssuance("supply", -5),
}, {
	Name:      "invalid data (supply: omitted)",
	Error:     `*fat.Issuance: invalid "supply": must be positive or -1`,
	IssuerKey: issuerKey,
	Issuance:  omitFieldIssuance("supply"),
}, {
	Name:      "invalid ExtIDs (timestamp)",
	Error:     `timestamp salt expired`,
	IssuerKey: issuerKey,
	Issuance: func() Issuance {
		i := validIssuance()
		i.ExtIDs[0] = factom.Bytes("10")
		return i
	}(),
}, {
	Name:      "invalid ExtIDs (length)",
	Error:     `invalid number of ExtIDs`,
	IssuerKey: issuerKey,
	Issuance: func() Issuance {
		i := validIssuance()
		i.ExtIDs = append(i.ExtIDs, factom.Bytes{})
		return i
	}(),
}, {
	Name:     "invalid RCD hash",
	Error:    `invalid RCD`,
	Issuance: validIssuance(),
}}

func TestIssuance(t *testing.T) {
	for _, test := range issuanceTests {
		t.Run(test.Name, func(t *testing.T) {
			assert := assert.New(t)
			i := test.Issuance
			key := test.IssuerKey
			err := i.Validate((*factom.Bytes32)(&key))
			if len(test.Error) == 0 {
				assert.NoError(err)
			} else {
				assert.EqualError(err, test.Error)
			}
		})
	}
}

func validIssuanceEntryContentMap() map[string]interface{} {
	return map[string]interface{}{
		"type":     "FAT-0",
		"supply":   int64(100000),
		"symbol":   "TEST",
		"metadata": []int{0},
	}
}

func validIssuance() Issuance {
	return issuanceFromRaw(marshal(validIssuanceEntryContentMap()))
}

var issuerSecret = func() factom.SK1Key {
	a, _ := factom.GenerateSK1Key()
	return a
}()
var issuerKey = issuerSecret.ID1Key()

func issuanceFromRaw(content factom.Bytes) Issuance {
	e := Entry{factom.Entry{
		ChainID: new(factom.Bytes32),
		Content: content,
	}}
	e.Sign(issuerSecret)
	id1Key := issuerSecret.ID1Key()
	i, _ := NewIssuance(e.Entry, (*factom.Bytes32)(&id1Key))
	return i
}

func invalidIssuance(field string) Issuance {
	return setFieldIssuance(field, []int{0})
}

func omitFieldIssuance(field string) Issuance {
	m := validIssuanceEntryContentMap()
	delete(m, field)
	return issuanceFromRaw(marshal(m))
}

func setFieldIssuance(field string, value interface{}) Issuance {
	m := validIssuanceEntryContentMap()
	m[field] = value
	return issuanceFromRaw(marshal(m))
}

func marshal(v map[string]interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}

var issuanceMarshalEntryTests = []struct {
	Name  string
	Error string
	Issuance
}{{
	Name:     "valid",
	Issuance: newIssuance(),
}, {
	Name: "valid (metadata)",
	Issuance: func() Issuance {
		i := newIssuance()
		i.Metadata = json.RawMessage(`{"memo":"new token"}`)
		return i
	}(),
}, {
	Name:  "invalid data",
	Error: `json: error calling MarshalJSON for type *fat.Issuance: invalid "type": invalid fat.Type: 1000`,
	Issuance: func() Issuance {
		i := newIssuance()
		i.Type = 1000
		return i
	}(),
}, {
	Name:  "invalid metadata JSON",
	Error: `json: error calling MarshalJSON for type *fat.Issuance: json: error calling MarshalJSON for type json.RawMessage: invalid character 'a' looking for beginning of object key string`,
	Issuance: func() Issuance {
		i := newIssuance()
		i.Metadata = json.RawMessage("{asdf")
		return i
	}(),
}}

func TestIssuanceMarshalEntry(t *testing.T) {
	for _, test := range issuanceMarshalEntryTests {
		t.Run(test.Name, func(t *testing.T) {
			assert := assert.New(t)
			i := test.Issuance
			err := i.PopulateEntry(&issuerSecret)
			if len(test.Error) == 0 {
				assert.NoError(err)
			} else {
				assert.EqualError(err, test.Error)
			}
		})
	}
}

func newIssuance() Issuance {
	return Issuance{
		Type:   TypeFAT0,
		Supply: 1000000,
		Symbol: "TEST",
	}
}
