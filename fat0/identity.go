package fat0

import (
	"fmt"
	"time"

	"github.com/Factom-Asset-Tokens/fatd/factom"
)

// ValidIdentityChainID returns true if the chainID matches the pattern for an
// Identity Chain ID.
//
// The Identity Chain specification can be found here:
// https://github.com/FactomProject/FactomDocs/blob/master/Identity.md#factom-identity-chain-creation
func ValidIdentityChainID(chainID factom.Bytes) bool {
	if len(chainID) == len(factom.Bytes32{}) &&
		chainID[0] == 0x88 &&
		chainID[1] == 0x88 &&
		chainID[2] == 0x88 {
		return true
	}
	return false
}

// ValidIdentityNameIDs returns true if the nameIDs match the pattern for a
// valid Identity Chain. The nameIDs for a chain are the ExtIDs of the first
// entry in the chain.
//
// The Identity Chain specification can be found here:
// https://github.com/FactomProject/FactomDocs/blob/master/Identity.md#factom-identity-chain-creation
func ValidIdentityNameIDs(nameIDs []factom.Bytes) bool {
	if len(nameIDs) == 7 &&
		len(nameIDs[0]) == 1 && nameIDs[0][0] == 0x00 &&
		string(nameIDs[1]) == "Identity Chain" &&
		len(nameIDs[2]) == len(factom.Bytes32{}) &&
		len(nameIDs[3]) == len(factom.Bytes32{}) &&
		len(nameIDs[4]) == len(factom.Bytes32{}) &&
		len(nameIDs[5]) == len(factom.Bytes32{}) {
		return true
	}
	return false
}

// Identity represents the Token Issuer's Identity Chain and the public IDKey
// used to sign Issuance and coinbase Transaction Entries.
type Identity struct {
	ChainID   *factom.Bytes32
	IDKey     *factom.Bytes32
	Height    uint64
	Timestamp time.Time
}

// IsPopulated returns true if the Identity has been populated with an IDKey.
func (i Identity) IsPopulated() bool {
	return i.IDKey != nil
}

// Get validates i.ChainID as an Identity Chain and parses out the IDKey.
//
// Get returns any networking or marshaling errors, but not JSON RPC or chain
// parsing errors. To check if the Identity has been successfully populated,
// call IsPopulated().
func (i *Identity) Get() error {
	if i.ChainID == nil {
		return fmt.Errorf("ChainID is nil")
	}
	if i.IsPopulated() {
		return nil
	}
	if !ValidIdentityChainID(i.ChainID[:]) {
		return nil
	}

	// Get first entry block of Identity Chain.
	eb := factom.EBlock{ChainID: i.ChainID}
	if err := eb.GetFirst(); err != nil {
		return err
	}
	if !eb.IsFirst() {
		return nil
	}

	// Get first entry of first entry block.
	first := eb.Entries[0]
	if err := first.Get(); err != nil {
		return err
	}

	if !ValidIdentityNameIDs(first.ExtIDs) {
		return nil
	}

	i.IDKey = factom.NewBytes32(first.ExtIDs[2])
	i.Height = first.Height
	i.Timestamp = first.Timestamp.Time

	return nil
}
