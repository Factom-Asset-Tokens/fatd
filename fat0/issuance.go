package fat0

import (
	"crypto/sha256"
	"unicode/utf8"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"golang.org/x/crypto/ed25519"
)

func ValidTokenNameIDs(nameIDs []factom.Bytes) bool {
	if len(nameIDs) == 4 &&
		string(nameIDs[0]) == "token" && string(nameIDs[2]) == "issuer" &&
		ValidIdentityChainID(nameIDs[3]) &&
		utf8.Valid(nameIDs[1]) {
		return true
	}
	return false
}

type Issuance struct {
	Type   string `json:"type"`
	Supply int64  `json:"supply"`

	Symbol string `json:"symbol,omitempty"`
	Name   string `json:"name,omitempty"`
	Entry
}

func (i *Issuance) Valid(idKey *factom.Bytes32) bool {
	if !i.ValidExtIDs() {
		return false
	}
	if i.RCDHash() != *idKey {
		return false
	}
	if i.Unmarshal() != nil {
		return false
	}
	if !i.ValidData() {
		return false
	}
	if !i.VerifySignature() {
		return false
	}
	return true
}

func (i *Issuance) ValidData() bool {
	return i.Type == "FAT-0" && i.Supply != 0
}

func (i *Issuance) Unmarshal() error {
	return i.Entry.Unmarshal(i)
}

const (
	RCDType       byte = 0x01
	RCDSize            = ed25519.PublicKeySize + 1
	SignatureSize      = ed25519.SignatureSize
)

func (i *Issuance) ValidExtIDs() bool {
	return len(i.ExtIDs) == 2 &&
		len(i.ExtIDs[0]) == RCDSize && i.ExtIDs[0][0] == RCDType &&
		len(i.ExtIDs[1]) == SignatureSize
}

func (i *Issuance) RCDHash() [sha256.Size]byte {
	return sha256d(i.ExtIDs[0])
}

func (i *Issuance) VerifySignature() bool {
	pubKey := ed25519.PublicKey(i.ExtIDs[0][1:])
	return ed25519.Verify(pubKey, append(i.ChainID[:], i.Content...), i.ExtIDs[1])
}

func sha256d(data []byte) [sha256.Size]byte {
	hash := sha256.Sum256(data)
	return sha256.Sum256(hash[:])
}
