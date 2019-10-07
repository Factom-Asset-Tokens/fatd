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
	"github.com/Factom-Asset-Tokens/fatd/fat/jsonlen"
)

// AddressTokenMap relates the RCDHash of an address to its NFTokenIDs.
type AddressNFTokensMap map[factom.FAAddress]NFTokens

func (m AddressNFTokensMap) MarshalJSON() ([]byte, error) {
	if m.NumNFTokenIDs() == 0 {
		return nil, fmt.Errorf("empty")
	}
	if err := m.NoInternalNFTokensIntersection(); err != nil {
		return nil, err
	}
	adrStrTknsMap := make(map[string]NFTokens, len(m))
	for adr, tkns := range m {
		// Omit addresses with empty NFTokens.
		if len(tkns) == 0 {
			continue
		}
		adrStrTknsMap[adr.String()] = tkns
	}
	return json.Marshal(adrStrTknsMap)
}

func (m *AddressNFTokensMap) UnmarshalJSON(data []byte) error {
	var adrStrDataMap map[string]json.RawMessage
	if err := json.Unmarshal(data, &adrStrDataMap); err != nil {
		return fmt.Errorf("%T: %w", m, err)
	}
	if len(adrStrDataMap) == 0 {
		return fmt.Errorf("%T: empty", m)
	}
	adrJSONLen := len(`"":,`) + len(factom.FAAddress{}.String())
	expectedJSONLen := len(`{}`) - len(`,`) + len(adrStrDataMap)*adrJSONLen
	*m = make(AddressNFTokensMap, len(adrStrDataMap))
	var adr factom.FAAddress
	var tkns NFTokens
	var numTkns int
	for adrStr, data := range adrStrDataMap {
		if err := adr.Set(adrStr); err != nil {
			return fmt.Errorf("%T: %#v: %w", m, adrStr, err)
		}
		if err := tkns.UnmarshalJSON(data); err != nil {
			return fmt.Errorf("%T: %v: %w", m, err, adr)
		}
		numTkns += len(tkns)
		if numTkns > maxCapacity {
			return fmt.Errorf("%T(len:%v): %T(len:%v): %v",
				m, numTkns-len(tkns), tkns, len(tkns), ErrorCapacity)
		}
		if err := m.NoNFTokensIntersection(tkns); err != nil {
			return fmt.Errorf("%T: %w and %v", m, err, adr)
		}
		(*m)[adr] = tkns
		expectedJSONLen += len(jsonlen.Compact(data))
	}
	if expectedJSONLen != len(jsonlen.Compact(data)) {
		return fmt.Errorf("%T: unexpected JSON length", m)
	}
	return nil
}

func (m AddressNFTokensMap) NoNFTokensIntersection(newTkns NFTokens) error {
	for adr, existingTkns := range m {
		if err := existingTkns.NoIntersection(newTkns); err != nil {
			return fmt.Errorf("%w: %v", err, adr)
		}
	}
	return nil
}

func (m AddressNFTokensMap) NoAddressIntersection(n AddressNFTokensMap) error {
	short, long := m, n
	if len(short) > len(long) {
		short, long = long, short
	}
	for rcdHash, tkns := range short {
		if len(tkns) == 0 {
			continue
		}
		if tkns := long[rcdHash]; len(tkns) != 0 {
			return fmt.Errorf("duplicate address: %v", rcdHash)
		}
	}
	return nil
}

func (m AddressNFTokensMap) NFTokenIDsConserved(n AddressNFTokensMap) error {
	numTknIDs := m.NumNFTokenIDs()
	if numTknIDs != n.NumNFTokenIDs() {
		return fmt.Errorf("number of NFTokenIDs differ")
	}
	allTkns := m.AllNFTokens()
	for _, tkns := range n {
		for tknID := range tkns {
			if _, ok := allTkns[tknID]; !ok {
				return fmt.Errorf("missing NFTokenID: %v", tknID)
			}
		}
	}
	return nil
}

func (m AddressNFTokensMap) AllNFTokens() NFTokens {
	allTkns := make(NFTokens, len(m))
	for _, tkns := range m {
		for tknID := range tkns {
			allTkns[tknID] = struct{}{}
		}
	}
	return allTkns
}

func (m AddressNFTokensMap) NumNFTokenIDs() int {
	var numTknIDs int
	for _, tkns := range m {
		numTknIDs += len(tkns)
	}
	return numTknIDs
}

func (m AddressNFTokensMap) NoInternalNFTokensIntersection() error {
	allTkns := make(NFTokens, m.NumNFTokenIDs())
	for rcdHash, tkns := range m {
		if err := allTkns.Append(tkns); err != nil {
			// We found an intersection. To identify the other
			// RCDHash that owns tknID, we temporarily remove
			// rcdHash from m and restore it after we return.
			tknID := NFTokenID(err.(ErrorNFTokenIDIntersection))
			delete(m, rcdHash)
			otherRCDHash := m.Owner(tknID)
			m[rcdHash] = tkns
			return fmt.Errorf("%w: %v and %v", err, rcdHash, otherRCDHash)

		}
	}
	return nil
}

func (m AddressNFTokensMap) Owner(tknID NFTokenID) factom.FAAddress {
	var adr factom.FAAddress
	var tkns NFTokens
	for adr, tkns = range m {
		if _, ok := tkns[tknID]; ok {
			break
		}
	}
	return adr
}
