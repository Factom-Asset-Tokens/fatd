package fat0

import (
	"crypto/sha256"
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

func (i Issuance) ExpectedJSONLength() int {
	l := len(`{`)
	l += len(`"type":`) + len(`"`) + len(i.Type) + len(`"`)
	l += len(`,"supply":`) + digitLen(i.Supply)
	l += jsonStringLen("symbol", i.Symbol)
	l += jsonStringLen("name", i.Name)
	l += i.metadataLen()
	l += len(`}`)
	return l
}

func jsonStringLen(name, value string) int {
	if len(value) != 0 {
		return len(`,"`) + len(name) + len(`":`) + len(`"`) + len(value) + len(`"`)
	}
	return 0
}

// NewIssuance returns an Issuance initialized with the given entry.
func NewIssuance(entry factom.Entry) Issuance {
	return Issuance{Entry: Entry{Entry: entry}}
}

// UnmarshalEntry unmarshals the entry content as an Issuance.
func (i *Issuance) UnmarshalEntry() error {
	return i.unmarshalEntry(i)
}

// MarshalEntry marshals the entry content as an Issuance.
func (i *Issuance) MarshalEntry() error {
	return i.marshalEntry(i)
}

// Valid performs all validation checks and returns nil if i is a valid
// Issuance.
func (i *Issuance) Valid(idKey factom.Bytes32) error {
	if err := i.UnmarshalEntry(); err != nil {
		return err
	}
	if err := i.ValidData(); err != nil {
		return err
	}
	if err := i.ValidExtIDs(); err != nil {
		return err
	}
	if i.RCDHash() != idKey {
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

// RCDHash returns the SHA256d hash of the first external ID of the entry,
// which should be the RCD of the IDKey of the issuing Identity.
func (i Issuance) RCDHash() [sha256.Size]byte {
	return sha256d(i.ExtIDs[1])
}

// sha256d computes two rounds of the sha256 hash.
func sha256d(data []byte) [sha256.Size]byte {
	hash := sha256.Sum256(data)
	return sha256.Sum256(hash[:])
}
