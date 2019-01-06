package fat0

import (
	"encoding/json"
	"fmt"

	"github.com/Factom-Asset-Tokens/fatd/factom"
)

// AddressAmountMap relates the RCDHash of an address to its amount in a
// Transaction.
type AddressAmountMap map[factom.Bytes32]uint64

// UnmarshalJSON unmarshals a list of addresses and amounts used in the inputs
// or outputs of a transaction. Duplicate addresses or addresses with a 0
// amount cause an error.
func (mPtr *AddressAmountMap) UnmarshalJSON(data []byte) error {
	var mS map[string]uint64
	if err := json.Unmarshal(data, &mS); err != nil {
		return err
	}
	m := make(AddressAmountMap, len(mS))
	for faAdrStr, amount := range mS {
		adr, err := factom.NewAddressFromString(faAdrStr)
		if err != nil {
			return err
		}
		if amount == 0 {
			return fmt.Errorf("%T: invalid amount (0) for address: %v",
				mPtr, adr)
		}
		m[*adr.RCDHash()] = amount
	}
	*mPtr = m
	return nil
}

// MarshalJSON marshals a list of addresses and amounts used in the inputs or
// outputs of a transaction. Addresses with a 0 amount are omitted and pruned
// from a.
func (m AddressAmountMap) MarshalJSON() ([]byte, error) {
	mS := make(map[string]uint64, len(m))
	for rcdHash, amount := range m {
		// Omit addresses with 0 amounts.
		if amount == 0 {
			delete(m, rcdHash)
			continue
		}
		adr := factom.NewAddress(&rcdHash)
		mS[adr.String()] = amount
	}
	return json.Marshal(mS)
}

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
