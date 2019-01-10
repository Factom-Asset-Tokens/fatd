package fat1

import (
	"encoding/json"
	"fmt"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat0"
)

// Transaction represents a fat1 transaction, which can be a normal account
// transaction or a coinbase transaction depending on the Inputs and the
// RCD/signature pair.
type Transaction struct {
	Inputs  AddressNFTokensMap `json:"inputs"`
	Outputs AddressNFTokensMap `json:"outputs"`
	fat0.Entry
}

// NewTransaction returns a Transaction initialized with the given entry.
func NewTransaction(entry factom.Entry) Transaction {
	return Transaction{Entry: fat0.Entry{Entry: entry}}
}

func (t *Transaction) UnmarshalJSON(data []byte) error {
	tRaw := struct {
		Inputs  json.RawMessage `json:"inputs"`
		Outputs json.RawMessage `json:"outputs"`
		fat0.Entry
	}{}
	if err := json.Unmarshal(data, &tRaw); err != nil {
		return fmt.Errorf("%T: %v", t, err)
	}
	if err := json.Unmarshal(tRaw.Inputs, &t.Inputs); err != nil {
		return fmt.Errorf("%T.Inputs: %v", t, err)
	}
	if err := json.Unmarshal(tRaw.Outputs, &t.Outputs); err != nil {
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

func (t Transaction) ValidData() error {
	if err := t.Inputs.NoAddressIntersection(t.Outputs); err != nil {
		return err
	}
	if err := t.Inputs.NFTokenIDsConserved(t.Outputs); err != nil {
		return err
	}
	// Coinbase transactions must only have one input.
	if t.IsCoinbase() && len(t.Inputs) != 1 {
		return fmt.Errorf("invalid coinbase transaction")
	}
	return nil
}

var coinbase factom.Address

// IsCoinbase returns true if the coinbase address is in t.Input. This does not
// necessarily mean that t is a valid coinbase transaction.
func (t Transaction) IsCoinbase() bool {
	tkns := t.Inputs[*coinbase.RCDHash()]
	return len(tkns) != 0
}

// UnmarshalEntry unmarshals the entry content as a Transaction.
func (t *Transaction) UnmarshalEntry() error {
	return t.Entry.UnmarshalEntry(t)
}

// MarshalEntry marshals the Transaction into the Entry content.
func (t *Transaction) MarshalEntry() error {
	return t.Entry.MarshalEntry(t)
}

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

func (t Transaction) ValidExtIDs() error {
	if len(t.ExtIDs) != 2*len(t.Inputs)+1 {
		return fmt.Errorf("incorrect number of ExtIDs")
	}
	return t.Entry.ValidExtIDs()
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
	for inputRCDHash := range t.Inputs {
		if _, ok := rcdHashes[inputRCDHash]; !ok {
			return false
		}
	}
	return true
}
