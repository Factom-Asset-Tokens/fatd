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

package factom_test

import (
	"testing"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/stretchr/testify/assert"
)

var (
	JSONTimeInvalids = []struct {
		Data string
		Err  string
	}{
		{Data: `{}`,
			Err: "json: cannot unmarshal object into Go value of type uint64"},
		{Data: `"hello"`,
			Err: "json: cannot unmarshal string into Go value of type uint64"},
		{Data: `["hello"]`,
			Err: "json: cannot unmarshal array into Go value of type uint64"},
		{Data: `123445x`,
			Err: "invalid character 'x' after top-level value"}}
	JSONTimeValid = `1484858820`
)

func TestTimeUnmarshalJSON(t *testing.T) {
	for _, json := range JSONTimeInvalids {
		testTimeUnmarshalJSON(t, "Invalid", json.Data, json.Err)
	}
	json := JSONTimeValid
	t.Run("Valid", func(t *testing.T) {
		var time factom.Time
		assert.NoErrorf(t, time.UnmarshalJSON([]byte(json)), "json: %v", json)
	})
}

func testTimeUnmarshalJSON(t *testing.T, name string, json string, errStr string) {
	t.Run(name, func(t *testing.T) {
		var time factom.Time
		assert.EqualErrorf(t, time.UnmarshalJSON([]byte(json)),
			errStr, "json: %v", json)
	})
}
