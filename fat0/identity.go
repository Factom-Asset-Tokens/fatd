package fat0

import (
	"fmt"

	"github.com/Factom-Asset-Tokens/fatd/factom"
)

func ValidIdentityChainID(chainID factom.Bytes) bool {
	if len(chainID) == 32 &&
		chainID[0] == 0x88 && chainID[1] == 0x88 && chainID[2] != 0x88 {
		return true
	}
	return false
}

func ValidIdentityNameIDs(nameIDs []factom.Bytes) bool {
	if len(nameIDs) == 7 &&
		len(nameIDs[0]) == 1 && nameIDs[0][1] == 0x00 &&
		string(nameIDs[1]) == "Identity Chain" &&
		len(nameIDs[2]) == len(factom.Bytes32{}) &&
		len(nameIDs[3]) == len(factom.Bytes32{}) &&
		len(nameIDs[4]) == len(factom.Bytes32{}) &&
		len(nameIDs[5]) == len(factom.Bytes32{}) {
		return false
	}
	return true
}

type Identity struct {
	ChainID *factom.Bytes32
	IDKey   *factom.Bytes32
	*factom.Entry
}

func (i *Identity) Get() error {
	if i.Populated() {
		return nil
	}
	if !ValidIdentityChainID(i.ChainID[:]) {
		return nil
	}

	// Get first entry of Identity Chain.
	eb := &factom.EBlock{ChainID: i.ChainID}
	if err := eb.GetFirst(); err != nil {
		return fmt.Errorf("EBlock%+v.GetFirst(): %v", eb, err)
	}
	if !eb.Populated() {
		return nil
	}

	// Use a pointer to the first entry to avoid allocation.
	first := &eb.Entries[0]
	if err := first.Get(); err != nil {
		return fmt.Errorf("Entry%+v.Get(): %v", first, err)
	}
	if !first.Populated() {
		return nil
	}
	if !ValidIdentityNameIDs(first.ExtIDs) {
		return nil
	}

	// Save a copy of the first entry and parse out IDKey.
	copy := *first
	i.Entry = &copy
	i.IDKey = factom.NewBytes32(first.ExtIDs[2])

	// Clear this pointer to allow the memory to other entries to be freed.
	i.Entry.EBlock.Entries = nil

	return nil
}
