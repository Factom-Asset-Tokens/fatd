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

	"github.com/Factom-Asset-Tokens/fatd/fat/jsonlen"
)

// NFTokenIDRange represents a contiguous range of NFTokenIDs.
type NFTokenIDRange struct {
	Min NFTokenID `json:"min"`
	Max NFTokenID `json:"max"`
}

func NewNFTokenIDRange(minMax ...NFTokenID) NFTokenIDRange {
	var min, max NFTokenID
	if len(minMax) >= 2 {
		min, max = minMax[0], minMax[1]
		if min > max {
			min, max = max, min
		}
	} else if len(minMax) == 1 {
		min, max = minMax[0], minMax[0]
	}
	return NFTokenIDRange{Min: min, Max: max}
}

func (idRange NFTokenIDRange) IsJSONEfficient() bool {
	var expandedLen int
	for id := idRange.Min; id <= idRange.Max; id++ {
		expandedLen += id.jsonLen() + len(`,`)
	}
	return idRange.jsonLen() <= expandedLen
}

func (idRange NFTokenIDRange) IsStringEfficient() bool {
	var expandedLen int
	for id := idRange.Min; id <= idRange.Max; id++ {
		expandedLen += id.jsonLen() + len(`,`)
	}
	return idRange.strLen() <= expandedLen
}

func (idRange NFTokenIDRange) Len() int {
	return int(idRange.Max - idRange.Min + 1)
}

func (idRange NFTokenIDRange) Set(tkns NFTokens) error {
	if len(tkns)+idRange.Len() > maxCapacity {
		return fmt.Errorf("%T(len:%v): %T(%v): %v",
			tkns, len(tkns), idRange, idRange, ErrorCapacity)
	}
	for id := idRange.Min; id <= idRange.Max; id++ {
		if err := id.Set(tkns); err != nil {
			return err
		}
	}
	return nil
}

func (idRange NFTokenIDRange) Valid() error {
	if idRange.Len() > maxCapacity {
		return ErrorCapacity
	}
	if idRange.Min > idRange.Max {
		return fmt.Errorf("Min is greater than Max")
	}
	return nil
}

type nfTokenIDRange NFTokenIDRange

func (idRange NFTokenIDRange) String() string {
	if !idRange.IsStringEfficient() {
		ids := idRange.Slice()
		return fmt.Sprintf("%v", ids)
	}
	return fmt.Sprintf("%v-%v", idRange.Min, idRange.Max)
}

func (idRange NFTokenIDRange) MarshalJSON() ([]byte, error) {
	if err := idRange.Valid(); err != nil {
		return nil, err
	}
	if !idRange.IsJSONEfficient() {
		ids := idRange.Slice()
		return json.Marshal(ids)
	}
	return json.Marshal(nfTokenIDRange(idRange))
}

// Slice returns a sorted slice of tkns' NFTokenIDs.
func (idRange NFTokenIDRange) Slice() []NFTokenID {
	ids := make([]NFTokenID, idRange.Len())
	for i := range ids {
		ids[i] = NFTokenID(i) + idRange.Min
	}
	return ids
}

func (idRange *NFTokenIDRange) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, (*nfTokenIDRange)(idRange)); err != nil {
		return fmt.Errorf("%T: %w", idRange, err)
	}
	if err := idRange.Valid(); err != nil {
		return fmt.Errorf("%T: %w", idRange, err)
	}
	if len(jsonlen.Compact(data)) != idRange.jsonLen() {
		return fmt.Errorf("%T: unexpected JSON length", idRange)
	}
	return nil
}
func (idRange NFTokenIDRange) jsonLen() int {
	return len(`{"min":`) +
		idRange.Min.jsonLen() +
		len(`,"max":`) +
		idRange.Max.jsonLen() +
		len(`}`)
}

func (idRange NFTokenIDRange) strLen() int {
	return idRange.Min.jsonLen() + len(`-`) + idRange.Max.jsonLen()
}
