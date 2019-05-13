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
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/Factom-Asset-Tokens/base58"
)

// payload implements helper functions used by all Address and IDKey types.
type payload [sha256.Size]byte

// StringPrefix encodes payload as a base58check string with the given prefix.
func (pld payload) StringPrefix(prefix []byte) string {
	return base58.CheckEncode(pld[:], prefix...)
}

// MarshalJSONPrefix encodes payload as a base58check JSON string with the
// given prefix.
func (pld payload) MarshalJSONPrefix(prefix []byte) ([]byte, error) {
	return []byte(fmt.Sprintf("%q", pld.StringPrefix(prefix))), nil
}

// SetPrefix attempts to parse adrStr into adr enforcing that adrStr
// starts with prefix if prefix is not empty.
func (pld *payload) SetPrefix(str, prefix string) error {
	if len(str) != 50+len(prefix) {
		return fmt.Errorf("invalid length")
	}
	if len(prefix) > 0 && str[:len(prefix)] != prefix {
		return fmt.Errorf("invalid prefix")
	}
	b, _, err := base58.CheckDecode(str, len(prefix))
	if err != nil {
		return err
	}
	copy(pld[:], b)
	return nil
}

// UnmarshalJSONPrefix unmarshals a human readable address JSON string with the
// given prefix.
func (pld *payload) UnmarshalJSONPrefix(data []byte, prefix string) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	return pld.SetPrefix(str, prefix)
}
