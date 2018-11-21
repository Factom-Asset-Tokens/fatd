package fat0

import (
	"bytes"
	"encoding/json"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/FactomProject/ed25519"
)

// Entry has variables and methods common to all fat0 entries.
type Entry struct {
	Metadata json.RawMessage `json:"metadata,omitempty"`

	factom.Entry `json:"-"`
}

// unmarshalEntry unmarshals the content of the factom.Entry into the provided
// variable v, disallowing all unknown fields.
func (e Entry) unmarshalEntry(v interface{}) error {
	d := json.NewDecoder(bytes.NewReader(e.Content))
	d.DisallowUnknownFields()
	return d.Decode(v)
}

// validSignatures returns true if the first num RCD/signature pairs in the
// ExtIDs are valid.
func (e Entry) validSignatures(num int) bool {
	if num <= 0 || num*2 > len(e.ExtIDs) {
		return false
	}
	msg := append(e.ChainID[:], e.Content...)
	pubKey := new([ed25519.PublicKeySize]byte)
	sig := new([ed25519.SignatureSize]byte)
	for i := 0; i < num; i++ {
		copy(pubKey[:], e.ExtIDs[i*2][1:])
		copy(sig[:], e.ExtIDs[i*2+1])
		if !ed25519.VerifyCanonical(pubKey, msg, sig) {
			return false
		}
	}
	return true
}

// Sign the chain ID + content of the factom.Entry and add the RCD + signature
// pairs for the given addresses to the ExtIDs. This clears any existing
// ExtIDs.
func (e *Entry) Sign(as ...factom.Address) {
	msg := append(e.ChainID[:], e.Content...)
	e.ExtIDs = nil
	for _, a := range as {
		e.ExtIDs = append(e.ExtIDs, a.RCD(), ed25519.Sign(a.PrivateKey, msg)[:])
	}
}
