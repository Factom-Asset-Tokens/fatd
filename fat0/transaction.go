package fat0

import (
	"crypto/sha256"
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

func (t Transaction) ExpectedJSONLength() int {
	l := len(`{`)
	l += len(`"inputs":`) + addressAmountMapJSONLen(t.Inputs)
	l += len(`,`)
	l += len(`"outputs":`) + addressAmountMapJSONLen(t.Outputs)
	l += t.metadataLen()
	l += len(`}`)
	return l
}

func addressAmountMapJSONLen(m AddressAmountMap) int {
	l := len(`{}`)
	if len(m) > 0 {
		l += len(m) * len(`"FA3p291ptJvHAFjf22naELozdFEKfbAPt8zLKaGiSVXfM6AUDVM5":,`)
		l -= len(`,`)
		for _, a := range m {
			l += digitLen(int64(a))
		}
	}
	return l
}

func digitLen(d int64) int {
	l := 1
	if d < 0 {
		l += 1
		d *= -1
	}
	for pow := int64(10); d/pow != 0; pow *= 10 {
		l++
	}
	return l
}

// MarshalEntry marshals the entry content as a Transaction.
func (t *Transaction) MarshalEntry() error {
	return t.marshalEntry(t)
}

// IsCoinbase returns true if the coinbase address is in t.Input. This does not
// necessarily mean that t is a valid coinbase transaction.
func (t Transaction) IsCoinbase() bool {
	amount := t.Inputs[coinbase.RCDHash()]
	return amount != 0
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
	if t.IsCoinbase() {
		if t.RCDHash() != idKey {
			return fmt.Errorf("invalid RCD")
		}
	} else {
		if !t.ValidRCDs() {
			return fmt.Errorf("invalid RCDs")
		}
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
	if t.IsCoinbase() && len(t.Inputs) != 1 {
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
	if len(t.ExtIDs) != 2*len(t.Inputs)+1 {
		return fmt.Errorf("incorrect number of ExtIDs")
	}
	return t.Entry.ValidExtIDs()
}

//// ValidSignatures returns true if the RCD/signature pairs are valid.
//// ValidSignatures assumes that ValidExtIDs returns nil.
//func (t Transaction) ValidSignatures() bool {
//	return t.validSignatures(len(t.Inputs))
//}

// ValidRCDs returns true if for each input there is an external ID containing
// an RCD corresponding to the input. ValidRCDs assumes that UnmarshalEntry has
// been called and returned nil, and that ValidExtIDs returns nil.
func (t Transaction) ValidRCDs() bool {
	// Create a map of all RCDs that are present in the ExtIDs.
	rcdHashes := make(AddressAmountMap)
	extIDs := t.ExtIDs[1:]
	for i := 0; i < len(extIDs)/2; i++ {
		rcdHashes[sha256d(extIDs[i*2])] = 0
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
	return sha256d(t.ExtIDs[1])
}
