package factom

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/Factom-Asset-Tokens/base58"
)

// payload implements helper functions used by all Address and IDKey types.
type payload [sha256.Size]byte

// StringPrefix encodes payload as a base58check string with the given prefix.
func (pld payload) StringPrefix(prefix []byte) string {
	return base58.CheckEncode(pld[:], prefix...)
}

// MarshalJSONPrefix encodes payload as a base58check JSON string with the
// given prefix.
func (pld payload) MarshalJSONPrefix(prefix []byte) ([]byte, error) {
	return []byte(fmt.Sprintf("%q", pld.StringPrefix(prefix))), nil
}

// SetPrefix attempts to parse adrStr into adr enforcing that adrStr
// starts with prefix if prefix is not empty.
func (pld *payload) SetPrefix(str, prefix string) error {
	if len(str) != 50+len(prefix) {
		return fmt.Errorf("invalid length")
	}
	if len(prefix) > 0 && str[:len(prefix)] != prefix {
		return fmt.Errorf("invalid prefix")
	}
	b, _, err := base58.CheckDecode(str, len(prefix))
	if err != nil {
		return err
	}
	copy(pld[:], b)
	return nil
}

// UnmarshalJSONPrefix unmarshals a human readable address JSON string with the
// given prefix.
func (pld *payload) UnmarshalJSONPrefix(data []byte, prefix string) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	return pld.SetPrefix(str, prefix)
}
