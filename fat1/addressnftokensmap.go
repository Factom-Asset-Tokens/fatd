package fat1

import (
	"encoding/json"
	"fmt"

	"github.com/Factom-Asset-Tokens/fatd/factom"
)

// AddressTokenMap relates the RCDHash of an address to its NFTokenIDs.
type AddressNFTokensMap map[factom.RCDHash]NFTokens

func (m AddressNFTokensMap) MarshalJSON() ([]byte, error) {
	if m.NumNFTokenIDs() == 0 {
		return nil, fmt.Errorf("empty")
	}
	if err := m.NoInternalNFTokensIntersection(); err != nil {
		return nil, err
	}
	adrStrTknsMap := make(map[string]NFTokens, len(m))
	for rcdHash, tkns := range m {
		// Omit addresses with empty NFTokens.
		if len(tkns) == 0 {
			continue
		}
		adrStrTknsMap[rcdHash.String()] = tkns
	}
	return json.Marshal(adrStrTknsMap)
}

func (m *AddressNFTokensMap) UnmarshalJSON(data []byte) error {
	var adrStrDataMap map[string]json.RawMessage
	if err := json.Unmarshal(data, &adrStrDataMap); err != nil {
		return fmt.Errorf("%T: %v", m, err)
	}
	if len(adrStrDataMap) == 0 {
		return fmt.Errorf("%T: empty", m)
	}
	expectedJSONLen := len(`{}`) - len(`,`) +
		len(adrStrDataMap)*
			len(`"FA2MwhbJFxPckPahsmntwF1ogKjXGz8FSqo2cLWtshdU47GQVZDC":,`)
	*m = make(AddressNFTokensMap, len(adrStrDataMap))
	var rcdHash factom.RCDHash
	var tkns NFTokens
	for adrStr, data := range adrStrDataMap {
		if err := rcdHash.FromString(adrStr); err != nil {
			return fmt.Errorf("%T: %#v: %v", m, adrStr, err)
		}
		if err := tkns.UnmarshalJSON(data); err != nil {
			return fmt.Errorf("%T: %v: %v", m, rcdHash, err)
		}
		if err := m.NoNFTokensIntersection(tkns); err != nil {
			return fmt.Errorf("%T: %v and %v", m, rcdHash, err)
		}
		(*m)[rcdHash] = tkns
		expectedJSONLen += compactJSONLen(data)
	}
	if expectedJSONLen != compactJSONLen(data) {
		return fmt.Errorf("%T: unexpected JSON length", m)
	}
	return nil
}

func (m AddressNFTokensMap) NoNFTokensIntersection(newTkns NFTokens) error {
	for rcdHash, existingTkns := range m {
		if err := existingTkns.NoIntersection(newTkns); err != nil {
			return fmt.Errorf("%v: %v", rcdHash, err)
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
			return fmt.Errorf("duplicate Address: %v", rcdHash)
		}
	}
	return nil
}

func (m AddressNFTokensMap) NFTokenIDsConserved(n AddressNFTokensMap) error {
	numTknIDs := m.NumNFTokenIDs()
	if numTknIDs != n.NumNFTokenIDs() {
		return fmt.Errorf("number of NFTokenIDs differ")
	}
	allTkns := make(NFTokens, numTknIDs)
	for _, tkns := range m {
		for tknID := range tkns {
			allTkns[tknID] = struct{}{}
		}
	}
	for _, tkns := range n {
		for tknID := range tkns {
			if _, ok := allTkns[tknID]; !ok {
				return fmt.Errorf("missing NFTokenID: %v", tknID)
			}
		}
	}
	return nil
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
			tknID := NFTokenID(err.(errorNFTokenIDIntersection))
			delete(m, rcdHash)
			otherRCDHash := m.Owner(tknID)
			m[rcdHash] = tkns
			return fmt.Errorf("%v and %v: %v", rcdHash, otherRCDHash, err)

		}
	}
	return nil
}

func (m AddressNFTokensMap) Owner(tknID NFTokenID) factom.RCDHash {
	var rcdHash factom.RCDHash
	var tkns NFTokens
	for rcdHash, tkns = range m {
		if _, ok := tkns[tknID]; ok {
			break
		}
	}
	return rcdHash
}
