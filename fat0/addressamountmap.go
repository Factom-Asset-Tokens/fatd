package fat0

import (
	"encoding/json"
	"fmt"

	"github.com/Factom-Asset-Tokens/fatd/factom"
)

// AddressAmountMap relates the RCDHash of an address to its amount in a
// Transaction.
type AddressAmountMap map[factom.RCDHash]uint64

// MarshalJSON marshals a list of addresses and amounts used in the inputs or
// outputs of a transaction. Addresses with a 0 amount are omitted.
func (m AddressAmountMap) MarshalJSON() ([]byte, error) {
	if m.Sum() == 0 {
		return nil, fmt.Errorf("empty")
	}
	adrStrAmountMap := make(map[string]uint64, len(m))
	for rcdHash, amount := range m {
		// Omit addresses with 0 amounts.
		if amount == 0 {
			continue
		}
		adrStrAmountMap[rcdHash.String()] = amount
	}
	return json.Marshal(adrStrAmountMap)
}

// UnmarshalJSON unmarshals a list of addresses and amounts used in the inputs
// or outputs of a transaction. Duplicate addresses or addresses with a 0
// amount cause an error.
func (m *AddressAmountMap) UnmarshalJSON(data []byte) error {
	var adrStrAmountMap map[string]uint64
	if err := json.Unmarshal(data, &adrStrAmountMap); err != nil {
		return err
	}
	if len(adrStrAmountMap) == 0 {
		return fmt.Errorf("%T: empty", m)
	}
	expectedJSONLen := len(`{}`) - len(`,`) +
		len(adrStrAmountMap)*
			len(`"FA2MwhbJFxPckPahsmntwF1ogKjXGz8FSqo2cLWtshdU47GQVZDC":,`)
	*m = make(AddressAmountMap, len(adrStrAmountMap))
	var rcdHash factom.RCDHash
	for adrStr, amount := range adrStrAmountMap {
		if err := rcdHash.FromString(adrStr); err != nil {
			return fmt.Errorf("%T: %v", m, err)
		}
		if amount == 0 {
			return fmt.Errorf("%T: %v: invalid amount (0)",
				m, rcdHash)
		}
		(*m)[rcdHash] = amount
		expectedJSONLen += uint64StrLen(amount)
	}
	if expectedJSONLen != compactJSONLen(data) {
		return fmt.Errorf("%T: unexpected JSON length", m)
	}
	return nil
}

// Sum returns the sum of all amount values.
func (m AddressAmountMap) Sum() uint64 {
	var sum uint64
	for _, amount := range m {
		sum += amount
	}
	return sum
}

func int64StrLen(d int64) int {
	sign := 0
	if d < 0 {
		sign++
		d *= -1
	}
	return sign + uint64StrLen(uint64(d))
}

func uint64StrLen(d uint64) int {
	l := 1
	for pow := uint64(10); d/pow != 0; pow *= 10 {
		l++
	}
	return l
}

func (m AddressAmountMap) NoAddressIntersection(n AddressAmountMap) error {
	short, long := m, n
	if len(short) > len(long) {
		short, long = long, short
	}
	for rcdHash, amount := range short {
		if amount == 0 {
			continue
		}
		if amount := long[rcdHash]; amount != 0 {
			return fmt.Errorf("duplicate Address: %v", rcdHash)
		}
	}
	return nil
}
