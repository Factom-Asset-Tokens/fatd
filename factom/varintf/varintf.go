// Package varintf implements Factom's varInt_F specification.
//
// The varInt_F specifications uses the top bit (0x80) in each byte as the
// continuation bit. If this bit is set, continue to read the next byte. If
// this bit is not set, then this is the last byte. The remaining 7 bits are
// the actual data of the number. The bytes are ordered big endian, unlike the
// varInt used by protobuf or provided by package encoding/binary.
//
// https://github.com/FactomProject/FactomDocs/blob/master/factomDataStructureDetails.md#variable-integers-varint_f
package varintf

import (
	"math/bits"
)

const continuationBitMask = 0x80

// Encode x into varInt_F bytes.
func Encode(x uint64) []byte {
	bitlen := bits.Len64(x)
	buflen := bitlen / 7
	if bitlen == 0 || bitlen%7 > 0 {
		buflen++
	}
	buf := make([]byte, buflen)
	for i := range buf {
		buf[i] = continuationBitMask | uint8(x>>uint((buflen-i-1)*7))
	}
	// Unset continuation bit in last byte.
	buf[buflen-1] &^= continuationBitMask
	return buf
}

// Decode varInt_F bytes into a uint64 and return the number of bytes used. If
// buf encodes a number larger than 64 bits, 0 and -1 is returned.
func Decode(buf []byte) (uint64, int) {
	buflen := 1
	for b := buf[0]; b&continuationBitMask > 0; b = buf[buflen-1] {
		buflen++
	}
	if buflen > 10 || (buflen == 10 && buf[0] > 0x81) {
		return 0, -1
	}
	var x uint64
	for i := 0; i < buflen; i++ {
		x |= uint64(buf[i]&^continuationBitMask) << uint((buflen-i-1)*7)
	}
	return x, buflen
}
