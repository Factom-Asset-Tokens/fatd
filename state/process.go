package state

import (
	"fmt"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat0"
)

func ProcessEBlock(eb factom.EBlock) error {
	// Get the saved data for this chain.
	chain := chains.Get(eb.ChainID)

	// Skip ignored chains.
	if chain.IsIgnored() {
		return nil
	}

	// Load this Entry Block.
	if err := eb.Get(); err != nil {
		return fmt.Errorf("%#v.Get(): %v", eb, err)
	}
	if !eb.IsPopulated() {
		return fmt.Errorf("%#v.IsPopulated(): false", eb)
	}

	// Check for new chains.
	if eb.IsFirst() {
		// Load first entry of new chain.
		if err := eb.Entries[0].Get(); err != nil {
			return fmt.Errorf("%#v.Get: %v", eb.Entries[0], err)
		}
		if !eb.Entries[0].IsPopulated() {
			return fmt.Errorf("%#v.IsPopulated(): false", eb.Entries[0])
		}
		nameIDs := eb.Entries[0].ExtIDs

		// Ignore chains with NameIDs that don't match the fat0
		// pattern.
		if !fat0.ValidTokenNameIDs(nameIDs) {
			chains.Ignore(eb.ChainID)
			return nil
		}

		// Track this chain going forward.
		var err error
		chain, err = chains.Track(eb.ChainID, nameIDs)
		if err != nil {
			return err
		}
		// The first entry cannot be a valid Issuance entry, so discard
		// it and process the rest.
		eb.Entries = eb.Entries[1:]
	} else if !chain.IsTracked() {
		// Ignore chains that are not already tracked.
		chains.Ignore(eb.ChainID)
		return nil
	}

	if err := chain.ProcessEntries(eb.Entries); err != nil {
		return err
	}
	return nil
}
