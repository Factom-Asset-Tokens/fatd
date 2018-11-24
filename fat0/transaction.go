package fat0

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/Factom-Asset-Tokens/fatd/factom"
)

var (
	// coinbase is the factom.Address with an all zero private key.
	coinbase factom.Address
)

// Transaction represents a fat0 transaction, which can be a normal account
// transaction or a coinbase transaction depending on the Inputs and the
// RCD/signature pair.
type Transaction struct {
	Inputs  AddressAmountMap `json:"inputs"`
	Outputs AddressAmountMap `json:"outputs"`
	Salt    string           `json:"salt,omitempty"`
	Entry
}

// NewTransaction returns a Transaction initialized with the given entry.
func NewTransaction(entry factom.Entry) Transaction {
	return Transaction{Entry: Entry{Entry: entry}}
}

// UnmarshalEntry unmarshals the entry content as a Transaction.
func (t *Transaction) UnmarshalEntry() error {
	return t.unmarshalEntry(t)
}

// Coinbase returns true if the coinbase address is in t.Input. This does not
// necessarily mean that t is a valid coinbase transaction.
func (t Transaction) Coinbase() bool {
	_, ok := t.Inputs[coinbase.RCDHash()]
	return ok
}

// Valid performs all validation checks and returns nil if t is a valid
// Transaction. If t is a coinbase transaction then idKey is used to validate
// the RCD. Otherwise RCDs are checked against the input addresses.
func (t *Transaction) Valid(idKey factom.Bytes32) error {
	if err := t.UnmarshalEntry(); err != nil {
		return err
	}
	if err := t.ValidData(); err != nil {
		return err
	}
	if err := t.ValidExtIDs(); err != nil {
		return err
	}
	if t.Coinbase() {
		if t.RCDHash() != idKey {
			return fmt.Errorf("invalid RCD")
		}
	} else {
		if !t.ValidRCDs() {
			return fmt.Errorf("invalid RCDs")
		}
	}
	if !t.ValidSignatures() {
		return fmt.Errorf("invalid signatures")
	}
	return nil
}

// ValidData validates the Transaction data and returns nil if no errors are
// present. ValidData assumes that the entry content has been unmarshaled.
func (t Transaction) ValidData() error {
	if len(t.Inputs) == 0 {
		return fmt.Errorf("no inputs")
	}
	if len(t.Outputs) == 0 {
		return fmt.Errorf("no outputs")
	}
	if sum(t.Inputs) != sum(t.Outputs) {
		return fmt.Errorf("sum(inputs) != sum(outputs)")
	}
	// Coinbase transactions must only have one input.
	if t.Coinbase() && len(t.Inputs) != 1 {
		return fmt.Errorf("invalid coinbase transaction")
	}
	// Ensure that no address exists in both the Inputs and Outputs.
	if !emptyIntersection(t.Inputs, t.Outputs) {
		return fmt.Errorf("an address appears in both the inputs and the outputs")
	}
	return nil
}

// sum the amounts in aam.
func sum(aam AddressAmountMap) uint64 {
	var sum uint64
	for _, amount := range aam {
		sum += amount
	}
	return sum
}

// emptyIntersection returns true if a and b have no keys with non-zero values
// in common.
func emptyIntersection(a, b AddressAmountMap) bool {
	// Select the shortest map to range through.
	short := a
	long := b
	if len(b) < len(a) {
		short = b
		long = a
	}
	for rcdHash, amount := range short {
		// Omit addresses with 0 amounts.
		if amount == 0 {
			continue
		}
		if amount := long[rcdHash]; amount != 0 {
			return false
		}
	}
	return true
}

// ValidExtIDs validates the structure of the external IDs of the entry to make
// sure that it has the correct number of RCD/signature pairs. ValidExtIDs does
// not validate the content of the RCD or signature. ValidExtIDs assumes that
// the entry content has been unmarshaled and that ValidData returns nil.
func (t Transaction) ValidExtIDs() error {
	if len(t.ExtIDs) != 2*len(t.Inputs) {
		return fmt.Errorf("invalid number of ExtIDs")
	}
	for i := 0; i < len(t.Inputs); i++ {
		rcd := t.ExtIDs[i*2]
		if len(rcd) != factom.RCDSize {
			return fmt.Errorf("invalid RCD size")
		}
		if rcd[0] != factom.RCDType {
			return fmt.Errorf("invalid RCD type")
		}
		sig := t.ExtIDs[i*2+1]
		if len(sig) != factom.SignatureSize {
			return fmt.Errorf("invalid signature size")
		}
	}
	return nil
}

// ValidSignatures returns true if the RCD/signature pairs are valid.
// ValidSignatures assumes that ValidExtIDs returns nil.
func (t Transaction) ValidSignatures() bool {
	return t.validSignatures(len(t.Inputs))
}

// ValidRCDs returns true if for each input there is an external ID containing
// an RCD corresponding to the input. ValidRCDs assumes that UnmarshalEntry has
// been called and returned nil, and that ValidExtIDs returns nil.
func (t Transaction) ValidRCDs() bool {
	// Create a map of all RCDs that are present in the ExtIDs.
	rcdHashes := make(AddressAmountMap)
	for i := 0; i < len(t.ExtIDs)/2; i++ {
		rcdHashes[sha256d(t.ExtIDs[i*2])] = 0
	}

	// Ensure that for all Inputs there is a corresponding RCD in the
	// ExtIDs.
	for inputRCDHash := range t.Inputs {
		if _, ok := rcdHashes[inputRCDHash]; !ok {
			return false
		}
	}
	return true
}

// RCDHash returns the SHA256d hash of the first external ID of the entry,
// which should be the RCD of the IDKey of the issuing Identity, if t is a
// coinbase transaction.
func (t Transaction) RCDHash() [sha256.Size]byte {
	return sha256d(t.ExtIDs[0])
}

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
		if _, duplicate := aam[address.RCDHash()]; duplicate {
			return fmt.Errorf("duplicate address: %v", address)
		}
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
