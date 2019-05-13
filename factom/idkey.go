// MIT License
//
// Copyright 2018 Canonical Ledgers, LLC
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
// IN THE SOFTWARE.

package factom

import (
	"crypto/sha256"

	"golang.org/x/crypto/ed25519"
)

// IDKey is the interface implemented by the four ID and SK Key types.
type IDKey interface {
	// PrefixBytes returns the prefix bytes for the key.
	PrefixBytes() []byte
	// PrefixString returns the encoded prefix string for the key.
	PrefixString() string

	// String encodes the key to a base58check string with the appropriate
	// prefix.
	String() string
	// Payload returns the key as a byte array.
	Payload() [sha256.Size]byte
	// RCDHash returns the RCDHash as a byte array. For IDxKeys, this is
	// identical to Payload. For SKxKeys the RCDHash is computed.
	RCDHash() [sha256.Size]byte

	// IDKey returns the corresponding IDxKey in an IDKey interface.
	// IDxKeys return themselves.  Private SKxKeys compute the
	// corresponding IDxKey.
	IDKey() IDKey
}

// SKKey is the interface implemented by the four SK Key types.
type SKKey interface {
	IDKey

	// SKKey returns the SKKey interface. IDxKeys return themselves.
	// Private SKxKeys compute the corresponding IDxKey.
	SKKey() SKKey

	// RCD returns the RCD corresponding to the private key.
	RCD() []byte

	// PrivateKey returns the ed25519.PrivateKey which can be used for
	// signing data.
	PrivateKey() ed25519.PrivateKey
	// PublicKey returns the ed25519.PublicKey which can be used for
	// verifying signatures.
	PublicKey() ed25519.PublicKey
}
