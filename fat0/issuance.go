package fat0

import (
	"encoding/json"
	"fmt"

	"github.com/Factom-Asset-Tokens/fatd/factom"
)

// Issuance represents the Issuance of a token.
type Issuance struct {
	Type   string `json:"type"`
	Supply int64  `json:"supply"`

	Symbol string `json:"symbol,omitempty"`
	Name   string `json:"name,omitempty"`
	Entry
}

type issuance Issuance

func (i *Issuance) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, (*issuance)(i)); err != nil {
		return fmt.Errorf("%T: %v", i, err)
	}
	if err := i.ValidData(); err != nil {
		return fmt.Errorf("%T: %v", i, err)
	}
	if i.expectedJSONLength() != compactJSONLen(data) {
		return fmt.Errorf("%T: unexpected JSON length", i)
	}
	return nil
}

// ExpectedJSONLength returns the expected JSON length for i.
func (i Issuance) expectedJSONLength() int {
	l := len(`{}`)
	l += len(`"type":""`) + len(i.Type)
	l += len(`,"supply":`) + int64StrLen(i.Supply)
	l += jsonStrLen("symbol", i.Symbol)
	l += jsonStrLen("name", i.Name)
	l += i.MetadataJSONLen()
	return l
}
func jsonStrLen(name, value string) int {
	if len(value) == 0 {
		return 0
	}
	return len(`,"":""`) + len(name) + len(value)
}

func (i Issuance) MarshalJSON() ([]byte, error) {
	if err := i.ValidData(); err != nil {
		return nil, err
	}
	return json.Marshal(issuance(i))
}

// NewIssuance returns an Issuance initialized with the given entry.
func NewIssuance(entry factom.Entry) Issuance {
	return Issuance{Entry: Entry{Entry: entry}}
}

// UnmarshalEntry unmarshals the entry content as an Issuance.
func (i *Issuance) UnmarshalEntry() error {
	return i.Entry.UnmarshalEntry(i)
}

// MarshalEntry marshals the entry content as an Issuance.
func (i *Issuance) MarshalEntry() error {
	return i.Entry.MarshalEntry(i)
}

// Valid performs all validation checks and returns nil if i is a valid
// Issuance.
func (i *Issuance) Valid(idKey *factom.RCDHash) error {
	if err := i.UnmarshalEntry(); err != nil {
		return err
	}
	if err := i.ValidExtIDs(); err != nil {
		return err
	}
	if i.RCDHash(0) != *idKey {
		return fmt.Errorf("invalid RCD")
	}
	return nil
}

// ValidData validates the Issuance data and returns nil if no errors are
// present. ValidData assumes that the entry content has been unmarshaled.
func (i Issuance) ValidData() error {
	if i.Type != "FAT-0" {
		return fmt.Errorf(`invalid "type": %#v`, i.Type)
	}
	if i.Supply == 0 || i.Supply < -1 {
		return fmt.Errorf(`invalid "supply": must be positive or -1`)
	}
	return nil
}

// ValidExtIDs validates the structure of the external IDs of the entry to make
// sure that it has an RCD and signature. It does not validate the content of
// the RCD or signature.
func (i Issuance) ValidExtIDs() error {
	if len(i.ExtIDs) != 3 {
		return fmt.Errorf("incorrect number of ExtIDs")
	}
	return i.Entry.ValidExtIDs()
}
