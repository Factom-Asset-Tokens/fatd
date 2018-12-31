package factom

import (
	"crypto/sha256"
	"fmt"

	"github.com/Factom-Asset-Tokens/base58"
	"github.com/FactomProject/ed25519"
)

var (
	prefixChars = [...]byte{'F', 'A'}
	prefix      = [...]byte{0x5f, 0xb1}
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
	PrivateKey *[ed25519.PrivateKeySize]byte
	PublicKey  *[ed25519.PublicKeySize]byte
	rcdHash    *Bytes32
	rcd        []byte
}

// NewAddress returns an Address with the given rcdHash.
func NewAddress(rcdHash *Bytes32) Address {
	return Address{rcdHash: rcdHash}
}

func (a *Address) Get() error {
	params := struct {
		A *Address `json:"address"`
	}{A: a}
	result := struct {
		Key *privateKey `json:"secret"`
	}{}
	if err := WalletRequest("address", params, &result); err != nil {
		return err
	}
	a.PrivateKey = (*[ed25519.PrivateKeySize]byte)(result.Key)
	a.PublicKey = ed25519.GetPublicKey(a.PrivateKey)
	return nil
}

type privateKey [ed25519.PrivateKeySize]byte

func (pk *privateKey) UnmarshalJSON(data []byte) error {
	if data[0] != '"' || data[len(data)-1] != '"' {
		return fmt.Errorf("invalid type")
	}
	data = data[1 : len(data)-1]
	if len(data) != 52 {
		return fmt.Errorf("invalid length")
	}
	if string(data[0:2]) != "Fs" {
		return fmt.Errorf("invalid prefix")
	}
	b, _, err := base58.CheckDecode(string(data), 2)
	if err != nil {
		return err
	}
	copy(pk[:], b)
	return nil
}

// UnmarshalJSON unmarshals a string with a human readable Factoid Address.
func (a *Address) UnmarshalJSON(data []byte) error {
	if data[0] != '"' || data[len(data)-1] != '"' {
		return fmt.Errorf("invalid type")
	}
	data = data[1 : len(data)-1]
	if len(data) != 52 {
		return fmt.Errorf("invalid length")
	}
	if string(data[0:2]) != "FA" {
		return fmt.Errorf("invalid prefix")
	}
	b, _, err := base58.CheckDecode(string(data), 2)
	if err != nil {
		return err
	}
	a.rcdHash = NewBytes32(b)
	return nil
}

// MarshalJSON marshals a string with the human readable Factoid Address.
func (a Address) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%#v", a.String())), nil
}

// String returns the human readable Factoid Address.
func (a Address) String() string {
	a.RCDHash()
	return encodePub(a.rcdHash[:])
}

// RCD returns the RCD of the Address. If the rcd is nil, then it is computed
// and saved for future reuse. If the PrivateKey is nil, then the PrivateKey is
// allocated with all zeroes.
func (a *Address) RCD() []byte {
	if a.rcd == nil {
		if a.PrivateKey == nil {
			a.PrivateKey = new([ed25519.PrivateKeySize]byte)
		}
		if a.PublicKey == nil {
			a.PublicKey = ed25519.GetPublicKey(a.PrivateKey)
		}
		a.rcd = append([]byte{RCDType}, a.PublicKey[:]...)
	}
	return a.rcd
}

// RCDHash returns the RCDHash of the Address. If the rcdHash is nil, then it
// is computed and saved for future reuse.
func (a *Address) RCDHash() *Bytes32 {
	if a.rcdHash == nil {
		rcdHash := Bytes32(sha256d(a.RCD()))
		a.rcdHash = &rcdHash
	}
	return a.rcdHash
}

// encodePub encodes data using a base58 checksum encoding with the two prefix
// bytes used for Factoid public addresses.
func encodePub(data []byte) string {
	return base58.CheckEncode(data, prefix[0], prefix[1])
}

// sha256d computes two rounds of the sha256 hash on data.
func sha256d(data []byte) [sha256.Size]byte {
	hash := sha256.Sum256(data)
	return sha256.Sum256(hash[:])
}
