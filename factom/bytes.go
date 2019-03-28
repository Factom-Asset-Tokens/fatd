package factom

import (
	"database/sql"
	"database/sql/driver"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// Bytes32 implements json.Marshaler and json.Unmarshaler to encode and decode
// strings with exactly 32 bytes of hex encoded data, such as Chain IDs and
// KeyMRs.
type Bytes32 [32]byte

// Bytes implements json.Marshaler and json.Unmarshaler to encode and decode
// strings with hex encoded data, such as an Entry's External IDs or content.
type Bytes []byte

// NewBytes32 allocates a new Bytes32 object with the first 32 bytes of data
// contained in s32.
func NewBytes32(s32 []byte) *Bytes32 {
	b32 := new(Bytes32)
	copy(b32[:], s32)
	return b32
}

// Set decodes a string with exactly 32 bytes of hex encoded data.
func (b *Bytes32) Set(hexStr string) error {
	if len(hexStr) != hex.EncodedLen(len(b)) {
		return fmt.Errorf("invalid length")
	}
	if _, err := hex.Decode(b[:], []byte(hexStr)); err != nil {
		return err
	}
	return nil
}

// Set decodes a string with hex encoded data.
func (b *Bytes) Set(hexStr string) error {
	*b = make(Bytes, hex.DecodedLen(len(hexStr)))
	if _, err := hex.Decode(*b, []byte(hexStr)); err != nil {
		return err
	}
	return nil
}

// UnmarshalJSON decodes a JSON string with exactly 32 bytes of hex encoded
// data.
func (b *Bytes32) UnmarshalJSON(data []byte) error {
	var hexStr string
	if err := json.Unmarshal(data, &hexStr); err != nil {
		return err
	}
	return b.Set(hexStr)
}

// UnmarshalJSON decodes a JSON string with hex encoded data.
func (b *Bytes) UnmarshalJSON(data []byte) error {
	var hexStr string
	if err := json.Unmarshal(data, &hexStr); err != nil {
		return err
	}
	return b.Set(hexStr)
}

// String encodes b as a hex string.
func (b Bytes32) String() string {
	return hex.EncodeToString(b[:])
}

// String encodes b as a hex string.
func (b Bytes) String() string {
	return hex.EncodeToString(b[:])
}

// MarshalJSON encodes b as a hex JSON string.
func (b Bytes32) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%#v", b.String())), nil
}

// MarshalJSON encodes b as a hex JSON string.
func (b Bytes) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%#v", b.String())), nil
}

// Scan expects v to be a byte slice with exactly 32 bytes of data.
func (b *Bytes32) Scan(v interface{}) error {
	data, ok := v.([]byte)
	if !ok {
		return fmt.Errorf("invalid type")
	}
	if len(data) != 32 {
		return fmt.Errorf("invalid length")
	}
	copy(b[:], data)
	return nil
}

// Value expects b to be a byte slice with exactly 32 bytes of data.
func (b Bytes32) Value() (driver.Value, error) {
	return b[:], nil
}

var _ sql.Scanner = &Bytes32{}
var _ driver.Valuer = Bytes32{}

var zeroBytes32 Bytes32

// ZeroBytes32 returns an all zero Byte32.
func ZeroBytes32() Bytes32 {
	return Bytes32{}
}
