package factom

import (
	"crypto/sha256"
	"fmt"

	"github.com/FactomProject/btcutil/base58"
	"github.com/FactomProject/ed25519"
)

var (
	prefixChars = [...]byte{'F', 'A'}
	prefix      = [...]byte{0x5f, 0xb1}
)

const (
	lenChecksum      = 4
	lenHumanReadable = 52
)

type Address struct {
	PrivateKey *[ed25519.PrivateKeySize]byte
	PublicKey  *[ed25519.PublicKeySize]byte
	rcdHash    *Bytes32
	rcd        []byte
}

func (a *Address) UnmarshalJSON(data []byte) error {
	data = trimQuotes(data)
	if len(data) != 52 {
		return fmt.Errorf("Invalid address length %v", len(data))
	}
	if data[0] != prefixChars[0] || data[1] != prefixChars[1] {
		return fmt.Errorf("Invalid address type")
	}
	b, _, _, err := base58.CheckDecodeWithTwoVersionBytes(string(data))
	if err != nil {
		return fmt.Errorf("base58.CheckDecodeWithTwoVersionBytes(%#v): %v",
			string(data), err)
	}
	a.rcdHash = NewBytes32(b)
	return nil
}

func (a *Address) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%#v", a.String())), nil
}

func (a *Address) String() string {
	a.RCDHash()
	return encodePub(a.rcdHash[:])
}

func (a *Address) RCD() []byte {
	if a.rcd == nil {
		if a.PrivateKey == nil {
			a.PrivateKey = new([ed25519.PrivateKeySize]byte)
		}
		if a.PublicKey == nil {
			a.PublicKey = ed25519.GetPublicKey(a.PrivateKey)
		}
		a.rcd = append([]byte{0x01}, a.PublicKey[:]...)
	}
	return a.rcd
}

func (a *Address) RCDHash() Bytes32 {
	if a.rcdHash == nil {
		rcdHash := Bytes32(sha256d(a.RCD()))
		a.rcdHash = &rcdHash
	}
	return *a.rcdHash
}

func encodePub(data []byte) string {
	return base58.CheckEncodeWithVersionBytes(data, prefix[0], prefix[1])
}

func sha256d(data []byte) [sha256.Size]byte {
	hash := sha256.Sum256(data)
	return sha256.Sum256(hash[:])
}
