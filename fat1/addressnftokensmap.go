package fat1

import (
	"encoding/json"
	"fmt"

	"github.com/Factom-Asset-Tokens/fatd/factom"
)

// AddressTokenMap relates the RCDHash of an address to its NFTokenIDs.
type AddressNFTokensMap map[factom.RCDHash]NFTokens

func (m AddressNFTokensMap) MarshalJSON() ([]byte, error) {
	if len(m) == 0 {
		return nil, fmt.Errorf("%T: empty", m)
	}
	if err := m.HasIntersection(); err != nil {
		return nil, fmt.Errorf("%T: %v", m, err)
	}
	strTknsMap := make(map[string]NFTokens, len(m))
	deleteMap := AddressNFTokensMap{} // This will rarely get populated.
	for rcdHash, nfTkns := range m {
		// Omit addresses with 0 amounts.
		if len(nfTkns) == 0 {
			deleteMap[rcdHash] = nil
			continue
		}
		adr := factom.NewAddress(&rcdHash)
		strTknsMap[adr.String()] = nfTkns
	}
	for rcdHash := range deleteMap {
		delete(m, rcdHash)
	}
	if len(strTknsMap) == 0 {
		return nil, fmt.Errorf("%T: empty", m)
	}
	return json.Marshal(strTknsMap)
}

func (m *AddressNFTokensMap) UnmarshalJSON(data []byte) error {
	var adrStrDataMap map[string]json.RawMessage
	if err := json.Unmarshal(data, &adrStrDataMap); err != nil {
		return fmt.Errorf("%T: %v", m, err)
	}
	if len(adrStrDataMap) == 0 {
		return fmt.Errorf("%T: empty", m)
	}
	compactedJSONLen := compactJSONLen(data)
	expectedJSONLen := len(`{}`) - len(`,`) +
		len(adrStrDataMap)*
			len(`"FA2MwhbJFxPckPahsmntwF1ogKjXGz8FSqo2cLWtshdU47GQVZDC":,`)
	*m = make(AddressNFTokensMap, len(adrStrDataMap))
	var rcdHash factom.RCDHash
	var nfTkns NFTokens
	for faAdrStr, data := range adrStrDataMap {
		if err := rcdHash.FromString(faAdrStr); err != nil {
			return fmt.Errorf("%T: %#v: %v", m, faAdrStr, err)
		}
		if err := nfTkns.UnmarshalJSON(data); err != nil {
			return fmt.Errorf("%T: %v: %v", m, rcdHash, err)
		}
		if err := m.Intersects(nfTkns); err != nil {
			return fmt.Errorf("%T: %v and %v", m, rcdHash, err)
		}
		(*m)[rcdHash] = nfTkns
		expectedJSONLen += len(data)
	}
	if expectedJSONLen != compactedJSONLen {
		return fmt.Errorf("%T: unexpected JSON length", m)
	}
	return nil
}

func (m AddressNFTokensMap) Intersects(nfTkns NFTokens) error {
	for rcdHash, existingNFTkns := range m {
		if err := existingNFTkns.Intersects(nfTkns); err != nil {
			return fmt.Errorf("%v: %v", rcdHash, err)
		}
	}
	return nil
}

func (m AddressNFTokensMap) HasIntersection() error {
	checked := make(map[factom.RCDHash]map[factom.RCDHash]bool, len(m))
	for rcdHash := range m {
		checked[rcdHash] = make(map[factom.RCDHash]bool, len(m))
	}
	for rcdHash, nfTkns := range m {
		if len(checked[rcdHash]) == len(m)-1 {
			continue
		}
		for rcdH, nfT := range m {
			if rcdHash == rcdH || checked[rcdHash][rcdH] {
				continue
			}
			if err := nfTkns.Intersects(nfT); err != nil {
				return fmt.Errorf("%v and %v: %v",
					rcdHash, rcdH, err)
			}
			checked[rcdHash][rcdH] = true
			checked[rcdH][rcdHash] = true
			if len(checked[rcdHash]) == len(m)-1 {
				break
			}
		}
	}
	return nil
}
