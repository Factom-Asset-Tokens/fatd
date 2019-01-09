package fat0

import (
	"encoding/json"
	"fmt"

	"github.com/Factom-Asset-Tokens/fatd/factom"
)

// AddressAmountMap relates the RCDHash of an address to its amount in a
// Transaction.
type AddressAmountMap map[factom.RCDHash]uint64

// UnmarshalJSON unmarshals a list of addresses and amounts used in the inputs
// or outputs of a transaction. Duplicate addresses or addresses with a 0
// amount cause an error.
func (m *AddressAmountMap) UnmarshalJSON(data []byte) error {
	var mS map[string]uint64
	if err := json.Unmarshal(data, &mS); err != nil {
		return err
	}
	*m = make(AddressAmountMap, len(mS))
	var rcdHash factom.RCDHash
	for faAdrStr, amount := range mS {
		if err := rcdHash.FromString(faAdrStr); err != nil {
			return fmt.Errorf("%T: %v", m, err)
		}
		if amount == 0 {
			return fmt.Errorf("%T: invalid amount (0) for address: %v",
				m, rcdHash)
		}
		(*m)[rcdHash] = amount
	}
	return nil
}

// MarshalJSON marshals a list of addresses and amounts used in the inputs or
// outputs of a transaction. Addresses with a 0 amount are omitted and pruned
// from a.
func (m AddressAmountMap) MarshalJSON() ([]byte, error) {
	mS := make(map[string]uint64, len(m))
	deleteMap := AddressAmountMap{}
	for rcdHash, amount := range m {
		// Omit addresses with 0 amounts.
		if amount == 0 {
			deleteMap[rcdHash] = 0
			continue
		}
		mS[rcdHash.String()] = amount
	}
	for rcdHash := range deleteMap {
		delete(m, rcdHash)
	}
	return json.Marshal(mS)
}

// Sum returns the sum of all amount values.
func (m AddressAmountMap) Sum() uint64 {
	var sum uint64
	for _, amount := range m {
		sum += amount
	}
	return sum
}

func (m AddressAmountMap) jsonLen() int {
	l := len(`{}`)
	if len(m) > 0 {
		l += len(m) *
			len(`"FA3p291ptJvHAFjf22naELozdFEKfbAPt8zLKaGiSVXfM6AUDVM5":,`)
		l -= len(`,`)
		for _, a := range m {
			l += digitStrLen(int64(a))
		}
	}
	return l
}

func digitStrLen(d int64) int {
	l := 1
	if d < 0 {
		l++
		d *= -1
	}
	for pow := int64(10); d/pow != 0; pow *= 10 {
		l++
	}
	return l
}
