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
	"fmt"
	"strconv"
)

type Type uint64

const (
	TypeFAT0 Type = iota
	TypeFAT1
)

func (t *Type) Set(s string) error {
	format := s[0:len(`FAT-`)]
	if format != `FAT-` {
		return fmt.Errorf("%T: invalid format", t)
	}
	num := s[len(format):]
	var err error
	if *(*uint64)(t), err = strconv.ParseUint(num, 10, 64); err != nil {
		return fmt.Errorf("%T: %w", t, err)
	}
	return nil
}

func (t *Type) UnmarshalJSON(data []byte) error {
	if data[0] != '"' || data[len(data)-1] != '"' {
		return fmt.Errorf("%T: expected JSON string", t)
	}
	data = data[1 : len(data)-1]
	return t.Set(string(data))
}

func (t Type) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%#v", t.String())), nil
}

func (t Type) String() string {
	fmtStr := "FAT-%v"
	if !t.IsValid() {
		fmtStr = "invalid fat.Type: %v"
	}
	return fmt.Sprintf(fmtStr, uint64(t))
}

func (t Type) IsValid() bool {
	switch t {
	case TypeFAT0:
		fallthrough
	case TypeFAT1:
		return true
	}
	return false
}
