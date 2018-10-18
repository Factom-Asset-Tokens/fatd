package factom

import (
	"crypto/sha256"
	"fmt"

	"github.com/FactomProject/btcutil/base58"
	"golang.org/x/crypto/ed25519"
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
	ed25519.PrivateKey
	ed25519.PublicKey
	rcdHash *Bytes32
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
	return encode(a.rcdHash[:])
}

func (a *Address) RCDHash() Bytes32 {
	if a.rcdHash == nil {
		if a.PublicKey == nil {
			a.PublicKey = make(ed25519.PublicKey, ed25519.PublicKeySize)
		}
		rcdHash := Bytes32(sha256d([]byte(a.PublicKey)))
		a.rcdHash = &rcdHash
	}
	return *a.rcdHash
}

func encode(data []byte) string {
	return base58.CheckEncodeWithVersionBytes(data, prefix[0], prefix[1])
}

func sha256d(data []byte) [sha256.Size]byte {
	hash := sha256.Sum256(data)
	return sha256.Sum256(hash[:])
}
