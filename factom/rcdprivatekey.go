package factom

import "golang.org/x/crypto/ed25519"

// RCDPrivateKey is the interface implemented by the four SK Key types and the
// Fs Address type.
type RCDPrivateKey interface {
	// RCD returns the RCD corresponding to the private key.
	RCD() []byte

	// PrivateKey returns the ed25519.PrivateKey which can be used for
	// signing data.
	PrivateKey() ed25519.PrivateKey
	// PublicKey returns the ed25519.PublicKey which can be used for
	// verifying signatures.
	PublicKey() ed25519.PublicKey
}
