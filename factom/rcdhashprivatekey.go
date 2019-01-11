package factom

import (
	"database/sql"
	"database/sql/driver"
	"fmt"

	"github.com/Factom-Asset-Tokens/base58"
	"github.com/FactomProject/ed25519"
)

type RCDHash [32]byte

type PrivateKey [ed25519.PrivateKeySize]byte

func NewRCDHash(s32 []byte) *RCDHash {
	b32 := new(RCDHash)
	copy(b32[:], s32)
	return b32
}
func NewPrivateKey(s64 []byte) *PrivateKey {
	b64 := new(PrivateKey)
	copy(b64[:], s64)
	return b64
}

func (pk *PrivateKey) Bytes() *[ed25519.PrivateKeySize]byte {
	return (*[ed25519.PrivateKeySize]byte)(pk)
}

func (pk *PrivateKey) PublicKey() *[ed25519.PublicKeySize]byte {
	return ed25519.GetPublicKey(pk.Bytes())
}

const (
	faPrefixStr = "FA"
	fsPrefixStr = "Fs"
)

func (rcdHash *RCDHash) UnmarshalJSON(data []byte) error {
	if err := unmarshalBase58JSON(rcdHash[:], data, 52, faPrefixStr); err != nil {
		return fmt.Errorf("%T: %v", rcdHash, err)
	}
	return nil
}

func (pk *PrivateKey) UnmarshalJSON(data []byte) error {
	if err := unmarshalBase58JSON(pk[:], data, 52, fsPrefixStr); err != nil {
		return fmt.Errorf("%T: %v", pk, err)
	}
	pk.PublicKey()
	return nil
}

func unmarshalBase58JSON(dst, data []byte, expectedLen int, prefix string) error {
	if data[0] != '"' || data[len(data)-1] != '"' {
		return fmt.Errorf("expected JSON string")
	}
	adrStr := string(data[1 : len(data)-1])
	return unmarshalBase58String(dst, adrStr, expectedLen, prefix)
}
func unmarshalBase58String(dst []byte, data string,
	expectedLen int, prefix string) error {
	if len(data) != expectedLen {
		return fmt.Errorf("invalid length")
	}
	if string(data[0:2]) != prefix {
		return fmt.Errorf("invalid prefix")
	}
	b, _, err := base58.CheckDecode(string(data), len(prefix))
	if err != nil {
		return err
	}
	copy(dst, b)
	return nil
}

func (rcdHash *RCDHash) FromString(faAdrStr string) error {
	return unmarshalBase58String(rcdHash[:], faAdrStr, 52, faPrefixStr)
}

func (pk *PrivateKey) FromString(fsAdrStr string) error {
	return unmarshalBase58String(pk[:], fsAdrStr, 52, fsPrefixStr)
}

func (rcdHash RCDHash) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%#v", rcdHash.String())), nil
}

func (pk PrivateKey) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%#v", pk.String())), nil
}

var (
	faPrefix = [...]byte{0x5f, 0xb1}
	fsPrefix = [...]byte{0x64, 0x78}
)

func (rcdHash RCDHash) String() string {
	return base58.CheckEncode(rcdHash[:], faPrefix[0], faPrefix[1])
}

func (pk PrivateKey) String() string {
	return base58.CheckEncode(pk[:], fsPrefix[0], fsPrefix[1])
}

func (b *RCDHash) Scan(v interface{}) error {
	return (*Bytes32)(b).Scan(v)
}
func (b RCDHash) Value() (driver.Value, error) {
	return (Bytes32)(b).Value()
}

var _ sql.Scanner = &RCDHash{}
var _ driver.Valuer = RCDHash{}
