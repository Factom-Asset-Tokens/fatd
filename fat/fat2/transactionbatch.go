package fat2

import (
	"encoding/json"
	"fmt"
	"github.com/Factom-Asset-Tokens/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat"
	"github.com/Factom-Asset-Tokens/fatd/fat/jsonlen"
)

const Type = fat.TypeFAT2

// TransactionBatch represents a fat2 entry, which can be a list of one or more
// transactions to be executed in order
type TransactionBatch struct {
	Version      uint          `json:"version"`
	Transactions []Transaction `json:"transactions"`
	fat.Entry
}

// NewTransactionBatch returns a TransactionBatch initialized with the given
// entry.
func NewTransactionBatch(entry factom.Entry) *TransactionBatch {
	return &TransactionBatch{Entry: fat.Entry{Entry: entry}}
}

type transactionBatch TransactionBatch

// UnmarshalJSON unmarshals the bytes of JSON into a TransactionBatch
// ensuring that there are no duplicate JSON keys.
func (t *TransactionBatch) UnmarshalJSON(data []byte) error {
	data = jsonlen.Compact(data)
	tRaw := struct {
		Version      json.RawMessage `json:"version"`
		Transactions json.RawMessage `json:"transactions"`
		fat.Entry
	}{}
	if err := json.Unmarshal(data, &tRaw); err != nil {
		return fmt.Errorf("%T: %v", t, err)
	}
	if err := json.Unmarshal(tRaw.Version, &t.Version); err != nil {
		return fmt.Errorf("%T.Version: %v", t, err)
	}
	if err := json.Unmarshal(tRaw.Transactions, &t.Transactions); err != nil {
		return fmt.Errorf("%T.Transactions: %v", t, err)
	}

	expectedJSONLen := len(`{"version":,"transactions":}`) +
		len(tRaw.Version) + len(tRaw.Transactions)
	if expectedJSONLen != len(data) {
		return fmt.Errorf("%T: unexpected JSON length", t)
	}
	return nil
}

// MarshalJSON marshals the TransactionBatch content field as JSON, but will
// raise an error if the batch fails the checks in ValidData()
func (t TransactionBatch) MarshalJSON() ([]byte, error) {
	if err := t.ValidData(); err != nil {
		return nil, err
	}
	return json.Marshal(transactionBatch(t))
}

func (t TransactionBatch) String() string {
	data, err := t.MarshalJSON()
	if err != nil {
		return err.Error()
	}
	return string(data)
}

// UnmarshalEntry unmarshals the Entry content as a TransactionBatch
func (t *TransactionBatch) UnmarshalEntry() error {
	return t.Entry.UnmarshalEntry(t)
}

// MarshalEntry marshals the TransactionBatch into the entry content
func (t *TransactionBatch) MarshalEntry() error {
	return t.Entry.MarshalEntry(t)
}

// Validate performs all validation checks and returns nil if it is a valid
// batch. This function assumes the struct's entry field is populated.
func (t *TransactionBatch) Validate() error {
	err := t.ValidData()
	if err != nil {
		return err
	}
	if err = t.ValidExtIDs(); err != nil {
		return err
	}
	return nil
}

// ValidData validates all Transaction data included in the batch and returns
// nil if it is valid. This function assumes that the entry content (or an
// independent JSON object) has been unmarshaled.
func (t *TransactionBatch) ValidData() error {
	if t.Version != 1 {
		return fmt.Errorf("invalid version")
	}
	if len(t.Transactions) == 0 {
		return fmt.Errorf("at least one output required")
	}
	for i, tx := range t.Transactions {
		if err := tx.Validate(); err != nil {
			return fmt.Errorf("invalid transaction at index %d: %v", i, err)
		}
	}
	return nil
}

// ValidExtIDs validates the structure of the external IDs of the entry to make
// sure that it has the correct number of RCD/signature pairs. If no errors are
// found, it will then validate the content of the RCD/signature pair. This
// function assumes that the entry content has been unmarshaled and that
// ValidData returns nil.
func (t TransactionBatch) ValidExtIDs() error {
	// Count unique inputs to know how many signatures are needed on the entry
	uniqueInputs := make(map[factom.FAAddress]struct{})
	for _, tx := range t.Transactions {
		uniqueInputs[tx.Input.Address] = struct{}{}
	}
	if err := t.Entry.ValidExtIDs(len(uniqueInputs)); err != nil {
		return err
	}
	// Create a map of all RCDs that are present in the ExtIDs
	includedRCDHashes := make(map[factom.FAAddress]struct{})
	extIDs := t.ExtIDs[1:]
	for i := 0; i < len(extIDs)/2; i++ {
		includedRCDHashes[t.FAAddress(i)] = struct{}{}
	}
	// Ensure that for all unique inputs there is a corresponding RCD in the ExtIDs
	for address, _ := range uniqueInputs {
		if _, ok := includedRCDHashes[address]; !ok {
			return fmt.Errorf("invalid RCDs")
		}
	}
	return nil
}

// HasConversions returns true if this batch contains at least one transaction
// with a conversion input/output pair. This function assumes that
// TransactionBatch.Valid() returns nil
func (t *TransactionBatch) HasConversions() bool {
	for _, tx := range t.Transactions {
		if tx.IsConversion() {
			return true
		}
	}
	return false
}
