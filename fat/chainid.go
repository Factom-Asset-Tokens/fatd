package fat

import (
	"unicode/utf8"

	"github.com/Factom-Asset-Tokens/fatd/factom"
)

// ValidTokenNameIDs returns true if the nameIDs match the pattern for a valid
// token chain.
func ValidTokenNameIDs(nameIDs []factom.Bytes) bool {
	if len(nameIDs) == 4 && len(nameIDs[1]) > 0 &&
		string(nameIDs[0]) == "token" && string(nameIDs[2]) == "issuer" &&
		factom.ValidIdentityChainID(nameIDs[3]) &&
		utf8.Valid(nameIDs[1]) {
		return true
	}
	return false
}

// NameIDs returns valid NameIDs
func NameIDs(tokenID string, issuerChainID factom.Bytes32) []factom.Bytes {
	return []factom.Bytes{
		[]byte("token"), []byte(tokenID),
		[]byte("issuer"), issuerChainID[:],
	}
}

// ChainID returns the chain ID for a given token ID and issuer Chain ID.
func ChainID(tokenID string, issuerChainID factom.Bytes32) factom.Bytes32 {
	return factom.ChainID(NameIDs(tokenID, issuerChainID))
}
