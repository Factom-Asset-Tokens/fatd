package fat0

import (
	"crypto/sha256"
	"fmt"
	"unicode/utf8"

	"github.com/Factom-Asset-Tokens/fatd/factom"
)

// ValidTokenNameIDs returns true if the nameIDs match the pattern for a valid
// token chain.
func ValidTokenNameIDs(nameIDs []factom.Bytes) bool {
	if len(nameIDs) == 4 && len(nameIDs[1]) > 0 &&
		string(nameIDs[0]) == "token" && string(nameIDs[2]) == "issuer" &&
		ValidIdentityChainID(nameIDs[3]) &&
		utf8.Valid(nameIDs[1]) {
		return true
	}
	return false
}

// ChainID returns the chain ID for a given token ID and issuer Chain ID.
func ChainID(tokenID string, issuerChainID *factom.Bytes32) *factom.Bytes32 {
	hash := sha256.New()
	extIDs := [][]byte{
		[]byte("token"), []byte(tokenID),
		[]byte("issuer"), issuerChainID[:],
	}
	for _, id := range extIDs {
		idSum := sha256.Sum256(id)
		hash.Write(idSum[:])
	}
	chainID := hash.Sum(nil)
	return factom.NewBytes32(chainID)
}

// Issuance represents the Issuance of a token.
type Issuance struct {
	Type   string `json:"type"`
	Supply int64  `json:"supply"`

	Symbol string `json:"symbol,omitempty"`
	Name   string `json:"name,omitempty"`
	Entry
}

// NewIssuance returns an Issuance initialized with the given entry.
func NewIssuance(entry factom.Entry) Issuance {
	return Issuance{Entry: Entry{Entry: entry}}
}

// UnmarshalEntry unmarshals the entry content as an Issuance.
func (i *Issuance) UnmarshalEntry() error {
	return i.unmarshalEntry(i)
}

// Valid performs all validation checks and returns nil if i is a valid
// Issuance.
func (i Issuance) Valid(idKey factom.Bytes32) error {
	if err := i.ValidExtIDs(); err != nil {
		return err
	}
	if i.RCDHash() != idKey {
		return fmt.Errorf("invalid RCD")
	}
	if err := i.UnmarshalEntry(); err != nil {
		return err
	}
	if err := i.ValidData(); err != nil {
		return err
	}
	if !i.ValidSignature() {
		return fmt.Errorf("invalid signature")
	}
	return nil
}

// ValidData validates the Issuance data and returns nil if no errors are
// present. ValidData assumes that the entry content has been unmarshaled.
func (i Issuance) ValidData() error {
	if i.Type != "FAT-0" {
		return fmt.Errorf(`invalid "type": %#v`, i.Type)
	}
	if i.Supply == 0 {
		return fmt.Errorf(`invalid "supply": must be positive or -1`)
	}
	return nil
}

// ValidExtIDs validates the structure of the external IDs of the entry to make
// sure that it has an RCD and signature. It does not validate the content of
// the RCD or signature.
func (i Issuance) ValidExtIDs() error {
	if len(i.ExtIDs) < 2 {
		return fmt.Errorf("insufficient number of ExtIDs")
	}
	if len(i.ExtIDs[0]) != factom.RCDSize {
		return fmt.Errorf("invalid RCD size")
	}
	if i.ExtIDs[0][0] != factom.RCDType {
		return fmt.Errorf("invalid RCD type")
	}
	if len(i.ExtIDs[1]) != factom.SignatureSize {
		return fmt.Errorf("invalid signature size")
	}
	return nil
}

// RCDHash returns the SHA256d hash of the first external ID of the entry,
// which should be the RCD of the IDKey of the issuing Identity.
func (i Issuance) RCDHash() [sha256.Size]byte {
	return sha256d(i.ExtIDs[0])
}

// ValidSignature returns true if the RCD/signature pair is valid.
// ValidSignature assumes that ValidExtIDs returns nil.
func (i Issuance) ValidSignature() bool {
	return i.validSignatures(1)
}

// sha256d computes two rounds of the sha256 hash.
func sha256d(data []byte) [sha256.Size]byte {
	hash := sha256.Sum256(data)
	return sha256.Sum256(hash[:])
}
