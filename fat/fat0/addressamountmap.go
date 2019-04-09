package fat0

import (
	"encoding/json"
	"fmt"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat/jsonlen"
)

// AddressAmountMap relates a factom.FAAddress to its amount for the Inputs and
// Outputs of a Transaction.
type AddressAmountMap map[factom.FAAddress]uint64

// MarshalJSON marshals a list of addresses and amounts used in the inputs or
// outputs of a transaction. Addresses with a 0 amount are omitted.
func (m AddressAmountMap) MarshalJSON() ([]byte, error) {
	if m.Sum() == 0 {
		return nil, fmt.Errorf("empty")
	}
	adrStrAmountMap := make(map[string]uint64, len(m))
	for adr, amount := range m {
		// Omit addresses with 0 amounts.
		if amount == 0 {
			continue
		}
		adrStrAmountMap[adr.String()] = amount
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
	adrJSONLen := len(`"":,`) + len(factom.FAAddress{}.String())
	expectedJSONLen := len(`{}`) - len(`,`) + len(adrStrAmountMap)*adrJSONLen
	*m = make(AddressAmountMap, len(adrStrAmountMap))
	var adr factom.FAAddress
	for adrStr, amount := range adrStrAmountMap {
		if err := adr.Set(adrStr); err != nil {
			return fmt.Errorf("%T: %v", m, err)
		}
		if amount == 0 {
			return fmt.Errorf("%T: invalid amount (0): %v", m, adr)
		}
		(*m)[adr] = amount
		expectedJSONLen += jsonlen.Uint64(amount)
	}
	if expectedJSONLen != len(data) {
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

func (m AddressAmountMap) NoAddressIntersection(n AddressAmountMap) error {
	short, long := m, n
	if len(short) > len(long) {
		short, long = long, short
	}
	for adr, amount := range short {
		if amount == 0 {
			continue
		}
		if amount := long[adr]; amount != 0 {
			return fmt.Errorf("duplicate address: %v", adr)
		}
	}
	return nil
}
