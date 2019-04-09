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
