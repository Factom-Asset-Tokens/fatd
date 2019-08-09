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

package factom

import (
	"database/sql"
	"database/sql/driver"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// Bytes32 implements json.Marshaler and json.Unmarshaler to encode and decode
// strings with exactly 32 bytes of hex encoded data, such as Chain IDs and
// KeyMRs.
type Bytes32 [32]byte

// Bytes implements json.Marshaler and json.Unmarshaler to encode and decode
// strings with hex encoded data, such as an Entry's External IDs or content.
type Bytes []byte

// NewBytes32 allocates a new Bytes32 object with the first 32 bytes of data
// contained in s32.
func NewBytes32(s32 []byte) *Bytes32 {
	b32 := new(Bytes32)
	copy(b32[:], s32)
	return b32
}

// NewBytes32FromString allocates a new Bytes32 object with the hex encoded
// string data contained in s32.
func NewBytes32FromString(s32 string) *Bytes32 {
	b32 := new(Bytes32)
	b32.Set(s32)
	return b32
}

// Set decodes a string with exactly 32 bytes of hex encoded data.
func (b *Bytes32) Set(hexStr string) error {
	if len(hexStr) == 0 {
		return nil
	}
	if len(hexStr) != hex.EncodedLen(len(b)) {
		return fmt.Errorf("invalid length")
	}
	if _, err := hex.Decode(b[:], []byte(hexStr)); err != nil {
		return err
	}
	return nil
}

// NewBytesFromString makes a new Bytes object with the hex encoded string data
// contained in s.
func NewBytesFromString(s string) Bytes {
	var b Bytes
	b.Set(s)
	return b
}

// Set decodes a string with hex encoded data.
func (b *Bytes) Set(hexStr string) error {
	*b = make(Bytes, hex.DecodedLen(len(hexStr)))
	if _, err := hex.Decode(*b, []byte(hexStr)); err != nil {
		return err
	}
	return nil
}

// UnmarshalJSON decodes a JSON string with exactly 32 bytes of hex encoded
// data.
func (b *Bytes32) UnmarshalJSON(data []byte) error {
	var hexStr string
	if err := json.Unmarshal(data, &hexStr); err != nil {
		return err
	}
	return b.Set(hexStr)
}

// UnmarshalJSON decodes a JSON string with hex encoded data.
func (b *Bytes) UnmarshalJSON(data []byte) error {
	var hexStr string
	if err := json.Unmarshal(data, &hexStr); err != nil {
		return err
	}
	return b.Set(hexStr)
}

// String encodes b as a hex string.
func (b Bytes32) String() string {
	return hex.EncodeToString(b[:])
}

// String encodes b as a hex string.
func (b Bytes) String() string {
	return hex.EncodeToString(b[:])
}

// Type returns "Bytes32". Satisfies pflag.Value interface.
func (b Bytes32) Type() string {
	return "Bytes32"
}

// Type returns "Bytes". Satisfies pflag.Value interface.
func (b Bytes) Type() string {
	return "Bytes"
}

// MarshalJSON encodes b as a hex JSON string.
func (b Bytes32) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%#v", b.String())), nil
}

// MarshalJSON encodes b as a hex JSON string.
func (b Bytes) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%#v", b.String())), nil
}

// Scan expects v to be a byte slice with exactly 32 bytes of data.
func (b *Bytes32) Scan(v interface{}) error {
	data, ok := v.([]byte)
	if !ok {
		return fmt.Errorf("invalid type")
	}
	if len(data) != 32 {
		return fmt.Errorf("invalid length")
	}
	copy(b[:], data)
	return nil
}

// Value expects b to be a byte slice with exactly 32 bytes of data.
func (b Bytes32) Value() (driver.Value, error) {
	return b[:], nil
}

var _ sql.Scanner = &Bytes32{}
var _ driver.Valuer = Bytes32{}

func (b Bytes32) IsZero() bool {
	return b == Bytes32{}
}
