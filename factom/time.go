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
	"encoding/json"
	"time"
)

// Time implements the json.Unmarshaler and json.Marshaler interface for
// correctly parsing the timestamps returned by the factomd JSON RPC API.
type Time time.Time

// UnmarshalJSON unmarshals a JSON Number containing the Unix seconds since
// epoch timestamp.
func (t *Time) UnmarshalJSON(data []byte) error {
	var sec uint64
	if err := json.Unmarshal(data, &sec); err != nil {
		return err
	}
	*t = Time(time.Unix(int64(sec), 0))
	return nil
}

// MarshalJSON marshals a JSON Number containing a Unix seconds since epoch
// timestamp.
func (t Time) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Time().Unix())
}

func (t Time) Time() time.Time {
	return (time.Time)(t)
}
