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
func (a *AddressAmountMap) UnmarshalJSON(data []byte) error {
	aam := make(AddressAmountMap)
	var aaS map[string]uint64
	if err := json.Unmarshal(data, &aaS); err != nil {
		return err
	}
	for address, amount := range aaS {
		data := []byte(fmt.Sprintf("%#v", address))
		address := factom.Address{}
		if amount == 0 {
			return fmt.Errorf("invalid amount (0) for address: %v", address)
		}
		json.Unmarshal(data, &address)
		aam[address.RCDHash()] = amount
	}
	*a = aam
	return nil
}

// MarshalJSON marshals a list of addresses and amounts used in the inputs or
// outputs of a transaction. Addresses with a 0 amount are omitted.
func (a AddressAmountMap) MarshalJSON() ([]byte, error) {
	as := make(map[string]uint64, len(a))
	for rcdHash, amount := range a {
		// Omit addresses with 0 amounts.
		if amount == 0 {
			continue
		}
		address := factom.NewAddress(&rcdHash)
		as[address.String()] = amount
	}
	return json.Marshal(as)
}
