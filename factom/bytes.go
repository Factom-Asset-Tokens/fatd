package factom

import (
	"bytes"
	"encoding/hex"
	"fmt"
)

type Bytes32 [32]byte

func NewBytes32(s32 []byte) *Bytes32 {
	if len(s32) != len(Bytes32{}) {
		return nil
	}
	b32 := new(Bytes32)
	copy(b32[:], s32)
	return b32
}

func (b Bytes32) String() string {
	return hex.EncodeToString(b[:])
}

func (b *Bytes32) UnmarshalJSON(data []byte) error {
	n, err := hex.Decode(b[:], bytes.Trim(data, "\""))
	if err != nil {
		return err
	}
	if n != len(b) {
		return fmt.Errorf("only read %v bytes out of %v", n, len(b))
	}
	return nil
}
func (b Bytes32) MarshalJSON() ([]byte, error) {
	return bytesMarshalJSON(b[:])
}

type Bytes []byte

func (b Bytes) String() string {
	return hex.EncodeToString(b[:])
}

func (b *Bytes) UnmarshalJSON(data []byte) error {
	data = bytes.Trim(data, "\"")
	l := hex.DecodedLen(len(data))
	if len(*b) == 0 {
		*b = make(Bytes, l)
	}
	_, err := hex.Decode(*b, data)
	if err != nil {
		return err
	}
	return nil
}

func (b Bytes) MarshalJSON() ([]byte, error) {
	return bytesMarshalJSON(b)
}

func bytesMarshalJSON(b []byte) ([]byte, error) {
	l := hex.EncodedLen(len(b)) + 2
	data := make([]byte, l)
	hex.Encode(data[1:], b[:])
	data[0] = '"'
	data[len(data)-1] = '"'
	return data, nil
}
