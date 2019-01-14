package fat0

import (
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
	Entry
}

// NewTransaction returns a Transaction initialized with the given entry.
func NewTransaction(entry factom.Entry) Transaction {
	return Transaction{Entry: Entry{Entry: entry}}
}

func (t *Transaction) UnmarshalJSON(data []byte) error {
	tRaw := struct {
		Inputs  json.RawMessage `json:"inputs"`
		Outputs json.RawMessage `json:"outputs"`
		Entry
	}{}
	if err := json.Unmarshal(data, &tRaw); err != nil {
		return fmt.Errorf("%T: %v", t, err)
	}
	if err := t.Inputs.UnmarshalJSON(tRaw.Inputs); err != nil {
		return fmt.Errorf("%T.Inputs: %v", t, err)
	}
	if err := t.Outputs.UnmarshalJSON(tRaw.Outputs); err != nil {
		return fmt.Errorf("%T.Outputs: %v", t, err)
	}
	t.Metadata = tRaw.Metadata

	if err := t.ValidData(); err != nil {
		return fmt.Errorf("%T: %v", t, err)
	}

	expectedJSONLen := len(`{"inputs":,"outputs":}`) +
		compactJSONLen(tRaw.Inputs) + compactJSONLen(tRaw.Outputs) +
		tRaw.MetadataJSONLen()
	if expectedJSONLen != compactJSONLen(data) {
		return fmt.Errorf("%T: unexpected JSON length", t)
	}

	return nil
}

type transaction Transaction

func (t Transaction) MarshalJSON() ([]byte, error) {
	if err := t.ValidData(); err != nil {
		return nil, err
	}
	return json.Marshal(transaction(t))
}

// UnmarshalEntry unmarshals the entry content as a Transaction.
func (t *Transaction) UnmarshalEntry() error {
	return t.Entry.UnmarshalEntry(t)
}

// MarshalEntry marshals the Transaction into the Entry content.
func (t *Transaction) MarshalEntry() error {
	return t.Entry.MarshalEntry(t)
}

// IsCoinbase returns true if the coinbase address is in t.Input. This does not
// necessarily mean that t is a valid coinbase transaction.
func (t Transaction) IsCoinbase() bool {
	amount := t.Inputs[*coinbase.RCDHash()]
	return amount != 0
}

// Valid performs all validation checks and returns nil if t is a valid
// Transaction. If t is a coinbase transaction then idKey is used to validate
// the RCD. Otherwise RCDs are checked against the input addresses.
func (t *Transaction) Valid(idKey *factom.RCDHash) error {
	if err := t.UnmarshalEntry(); err != nil {
		return err
	}
	if err := t.ValidExtIDs(); err != nil {
		return err
	}
	if t.IsCoinbase() {
		if t.RCDHash(0) != *idKey {
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
	if t.Inputs.Sum() != t.Outputs.Sum() {
		return fmt.Errorf("sum(inputs) != sum(outputs)")
	}
	// Coinbase transactions must only have one input.
	if t.IsCoinbase() && len(t.Inputs) != 1 {
		return fmt.Errorf("invalid coinbase transaction")
	}
	// Ensure that no address exists in both the Inputs and Outputs.
	if err := t.Inputs.NoAddressIntersection(t.Outputs); err != nil {
		return err
	}
	return nil
}

// ValidExtIDs validates the structure of the external IDs of the entry to make
// sure that it has the correct number of RCD/signature pairs. ValidExtIDs does
// not validate the content of the RCD or signature. ValidExtIDs assumes that
// the entry content has been unmarshaled and that ValidData returns nil.
func (t Transaction) ValidExtIDs() error {
	return t.Entry.ValidExtIDs(len(t.Inputs))
}

func (t Transaction) ValidRCDs() bool {
	// Create a map of all RCDs that are present in the ExtIDs.
	rcdHashes := make(map[factom.RCDHash]struct{}, len(t.Inputs))
	extIDs := t.ExtIDs[1:]
	for i := 0; i < len(extIDs)/2; i++ {
		rcdHashes[t.RCDHash(i)] = struct{}{}
	}

	// Ensure that for all Inputs there is a corresponding RCD in the
	// ExtIDs.
	for rcdHash := range t.Inputs {
		if _, ok := rcdHashes[rcdHash]; !ok {
			return false
		}
	}
	return true
}
