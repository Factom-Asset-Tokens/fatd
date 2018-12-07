package fat0

import (
	"encoding/json"
	"fmt"

	"github.com/Factom-Asset-Tokens/fatd/factom"
)

// AddressAmountMap relates the RCDHash of an address to its amount in a
// Transaction.
type AddressAmountMap map[factom.Bytes32]uint64

// AddressAmount is used to marshal and unmarshal the JSON representation of a
// list of inputs or outputs in a Transaction.
type AddressAmount struct {
	Address factom.Address `json:"address"`
	Amount  uint64         `json:"amount"`
}

// UnmarshalJSON unmarshals a list of addresses and amounts used in the inputs
// or outputs of a transaction. Duplicate addresses or addresses with a 0
// amount cause an error.
func (a *AddressAmountMap) UnmarshalJSON(data []byte) error {
	aam := make(AddressAmountMap)
	var aaS []AddressAmount
	if err := json.Unmarshal(data, &aaS); err != nil {
		return err
	}
	for _, aa := range aaS {
		if aa.Amount == 0 {
			return fmt.Errorf("invalid amount (0) for address: %v", aa.Address)
		}
		if _, duplicate := aam[aa.Address.RCDHash()]; duplicate {
			return fmt.Errorf("duplicate address: %v", aa.Address)
		}
		aam[aa.Address.RCDHash()] = aa.Amount
	}
	*a = aam
	return nil
}

// MarshalJSON marshals a list of addresses and amounts used in the inputs or
// outputs of a transaction. Addresses with a 0 amount are omitted.
func (a AddressAmountMap) MarshalJSON() ([]byte, error) {
	as := make([]AddressAmount, 0, len(a))
	for rcdHash, amount := range a {
		rcdHash := rcdHash
		// Omit addresses with 0 amounts.
		if amount == 0 {
			continue
		}

		as = append(as, AddressAmount{
			Address: factom.NewAddress(&rcdHash),
			Amount:  amount,
		})
	}
	return json.Marshal(as)
}
