package factom

import (
	"database/sql"
	"database/sql/driver"
	"encoding/hex"
	"fmt"
)

// Bytes32 implements json.Marshaler and json.Unmarshaler to encode and decode
// strings with exactly 32 bytes of hex encoded data, such as chain IDs and key
// MRs.
type Bytes32 [32]byte

// NewBytes32 allocates a new Bytes32 object with the first 32 bytes of data
// contained in s32.
func NewBytes32(s32 []byte) *Bytes32 {
	b32 := new(Bytes32)
	copy(b32[:], s32)
	return b32
}

// String returns the hex encoded data of b.
func (b Bytes32) String() string {
	return hex.EncodeToString(b[:])
}

// UnmarshalJSON unmarshals a string with exactly 32 bytes of hex encoded data.
func (b *Bytes32) UnmarshalJSON(data []byte) error {
	if data[0] != '"' || data[len(data)-1] != '"' {
		return fmt.Errorf("invalid type")
	}
	data = data[1 : len(data)-1]
	if len(data) != len(b)*2 {
		return fmt.Errorf("invalid length")
	}
	if _, err := hex.Decode(b[:], data); err != nil {
		return err
	}
	return nil
}

// MarshalJSON marshals b into hex encoded data.
func (b Bytes32) MarshalJSON() ([]byte, error) {
	return bytesMarshalJSON(b[:])
}

func (b *Bytes32) Scan(v interface{}) error {
	data, ok := v.([]byte)
	if !ok {
		return fmt.Errorf("value must be type []byte but is type %T", v)
	}
	if len(data) != 32 {
		return fmt.Errorf("invalid length")
	}
	copy(b[:], data)
	return nil
}
func (b Bytes32) Value() (driver.Value, error) {
	return b[:], nil
}

var _ sql.Scanner = &Bytes32{}
var _ driver.Valuer = Bytes32{}

// Bytes implements json.Marshaler and json.Unmarshaler to encode and decode
// strings with hex encoded data, such as an Entry's external IDs or content.
type Bytes []byte

// String returns the hex encoded data of b.
func (b Bytes) String() string {
	return hex.EncodeToString(b[:])
}

// UnmarshalJSON unmarshals a string of hex encoded data.
func (b *Bytes) UnmarshalJSON(data []byte) error {
	if data[0] != '"' || data[len(data)-1] != '"' {
		return fmt.Errorf("invalid type")
	}
	data = data[1 : len(data)-1]
	*b = make(Bytes, hex.DecodedLen(len(data)))

	_, err := hex.Decode(*b, data)
	if err != nil {
		return err
	}
	return nil
}

// MarshalJSON marshals b into hex encoded data.
func (b Bytes) MarshalJSON() ([]byte, error) {
	return bytesMarshalJSON(b)
}

// bytesMarshalJSON marshals b into hex encoded data.
func bytesMarshalJSON(b []byte) ([]byte, error) {
	l := hex.EncodedLen(len(b)) + 2
	data := make([]byte, l)
	hex.Encode(data[1:], b[:])
	data[0] = '"'
	data[len(data)-1] = '"'
	return data, nil
}

var zeroBytes32 Bytes32

// ZeroBytes32 returns an all zero Byte32.
func ZeroBytes32() Bytes32 {
	return Bytes32{}
}
