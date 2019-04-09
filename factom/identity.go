package factom

import "fmt"

// ValidIdentityChainID returns true if the chainID matches the pattern for an
// Identity Chain ID.
//
// The Identity Chain specification can be found here:
// https://github.com/FactomProject/FactomDocs/blob/master/Identity.md#factom-identity-chain-creation
func ValidIdentityChainID(chainID Bytes) bool {
	if len(chainID) == len(Bytes32{}) &&
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
func ValidIdentityNameIDs(nameIDs []Bytes) bool {
	if len(nameIDs) == 7 &&
		len(nameIDs[0]) == 1 && nameIDs[0][0] == 0x00 &&
		string(nameIDs[1]) == "Identity Chain" &&
		len(nameIDs[2]) == len(ID1Key{}) &&
		len(nameIDs[3]) == len(ID2Key{}) &&
		len(nameIDs[4]) == len(ID3Key{}) &&
		len(nameIDs[5]) == len(ID4Key{}) {
		return true
	}
	return false
}

// Identity represents the Token Issuer's Identity Chain and the public ID1Key.
type Identity struct {
	ID1 ID1Key
	Entry
}

// NewIdentity initializes an Identity with the given chainID.
func NewIdentity(chainID *Bytes32) (i Identity) {
	i.ChainID = chainID
	return
}

// IsPopulated returns true if the Identity has been populated with an ID1Key.
func (i Identity) IsPopulated() bool {
	return i.ID1 != ID1Key(zeroBytes32)
}

// Get validates i.ChainID as an Identity Chain and parses out the ID1Key.
func (i *Identity) Get(c *Client) error {
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
	eb := EBlock{ChainID: i.ChainID}
	if err := eb.GetFirst(c); err != nil {
		return err
	}

	// Get first entry of first entry block.
	first := eb.Entries[0]
	if err := first.Get(c); err != nil {
		return err
	}

	if !ValidIdentityNameIDs(first.ExtIDs) {
		return nil
	}

	i.Entry = first
	copy(i.ID1[:], first.ExtIDs[2])

	return nil
}
