package factom

import (
	"crypto/sha256"
	"fmt"

	"github.com/FactomProject/ed25519"
)

const (
	// RCDType is the magic number identifying the currenctly accepted RCD.
	RCDType byte = 0x01
	// RCDSize is the size of the RCD.
	RCDSize = ed25519.PublicKeySize + 1
	// SignatureSize is the size of the ed25519 signatures.
	SignatureSize = ed25519.SignatureSize
)

// Address represents a Factoid address.
type Address struct {
	privateKey *PrivateKey
	rcdHash    *RCDHash
	rcd        []byte
}

// NewAddress returns an Address with the given rcdHash.
func NewAddress(rcdHash *RCDHash) Address {
	return Address{rcdHash: rcdHash}
}

// NewAddressFromString returns an Address
func NewAddressFromString(adrStr string) (Address, error) {
	adr := Address{}
	err := adr.UnmarshalJSON([]byte(fmt.Sprintf("%#v", adrStr)))
	return adr, err
}

func (a *Address) Get() error {
	if a.privateKey != nil {
		return nil
	}
	params := struct {
		A *Address `json:"address"`
	}{A: a}
	result := struct {
		A *Address `json:"secret"`
	}{A: a}
	if err := WalletRequest("address", params, &result); err != nil {
		return err
	}
	return nil
}

// UnmarshalJSON unmarshals a string with a human readable Factoid Address.
func (a *Address) UnmarshalJSON(data []byte) error {
	if data[0] != '"' || data[len(data)-1] != '"' {
		return fmt.Errorf("%T: expected JSON string", a)
	}
	if len(data) < 2 {
		return fmt.Errorf("%T: invalid length", a)
	}
	adrStr := string(data[1 : len(data)-1])
	if err := a.FromString(adrStr); err != nil {
		return fmt.Errorf("%T: %v", a, err)
	}
	return nil
}

func (a *Address) FromString(adrStr string) error {
	var address interface{ FromString(string) error }
	switch adrStr[:2] {
	case "FA":
		a.rcdHash = new(RCDHash)
		address = a.rcdHash
	case "Fs":
		a.privateKey = new(PrivateKey)
		address = a.privateKey
	default:
		return fmt.Errorf("invalid prefix")
	}
	return address.FromString(adrStr)
}

// MarshalJSON marshals a string with the human readable Factoid Address.
func (a Address) MarshalJSON() ([]byte, error) {
	return a.RCDHash().MarshalJSON()
}

// String returns the human readable Factoid Address.
func (a Address) String() string {
	return a.RCDHash().String()
}

// RCD returns the RCD of the Address. If the rcd is nil, then it is computed
// and saved for future reuse. If the PrivateKey is nil, then the PrivateKey is
// allocated with all zeroes.
func (a *Address) RCD() []byte {
	if a.rcd == nil {
		a.rcd = make([]byte, RCDSize)
		a.rcd[0] = RCDType
		copy(a.rcd[1:], a.PublicKey()[:])
	}
	return a.rcd
}

// RCDHash returns the RCDHash of the Address. If the rcdHash is nil, then it
// is computed and saved for future reuse.
func (a *Address) RCDHash() *RCDHash {
	if a.rcdHash == nil {
		rcdHash := RCDHash(sha256d(a.RCD()))
		a.rcdHash = &rcdHash
	}
	return a.rcdHash
}
func (a *Address) PrivateKey() *PrivateKey {
	if a.privateKey == nil {
		a.privateKey = new(PrivateKey)
	}
	return a.privateKey
}
func (a *Address) PublicKey() *[ed25519.PublicKeySize]byte {
	return a.PrivateKey().PublicKey()
}

// sha256d computes two rounds of the sha256 hash on data.
func sha256d(data []byte) [sha256.Size]byte {
	hash := sha256.Sum256(data)
	return sha256.Sum256(hash[:])
}
