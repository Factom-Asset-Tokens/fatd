package state

import (
	"fmt"
	"sync"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat0"
)

// Assumption: eb is not nil and has valid ChainID and KeyMR.
func processEBlock(wg *sync.WaitGroup, eb factom.EBlock) {
	defer wg.Done()

	// Get the saved data for this chain.
	chain := chains.Get(eb.ChainID)

	// Skip ignored chains.
	if chain.Ignored() {
		return
	}

	// Load this Entry Block.
	if err := eb.Get(); err != nil {
		errorStop(fmt.Errorf("factom.GetEntryBlock(%#v): %v", eb, err))
		return
	}

	// Check whether this is a new chain.
	if !eb.IsFirst() {
		// Check whether this chain is tracked.
		if chain.Tracked() {
			if err := processEntries(chain, eb.Entries); err != nil {
				errorStop(err)
			}
			return
		}
		// Since the chian is not new and isn't already tracked, then
		// ignore this chain going forward.
		chains.Ignore(eb.ChainID)
		return
	}

	// Load first entry of new chain.
	if err := eb.Entries[0].Get(); err != nil {
		errorStop(fmt.Errorf("Entry%#v.Get: %v", eb.Entries[0], err))
		return
	}
	nameIDs := eb.Entries[0].ExtIDs

	// Filter out chains with NameIDs that don't match the fat0 pattern.
	if !fat0.ValidTokenNameIDs(nameIDs) {
		chains.Ignore(eb.ChainID)
		return
	}

	// Track this chain going forward.
	chains.Track(eb.ChainID, fat0.Identity{ChainID: factom.NewBytes32(nameIDs[3])})

	// The first entry cannot be a valid Issuance entry, so discard it and
	// process the rest.
	if err := processEntries(chain, eb.Entries[1:]); err != nil {
		errorStop(err)
		return
	}
}

func processEntries(chain Chain, es []factom.Entry) error {
	if !chain.Issued() {
		return processIssuance(chain, es)
	}
	return processTransactions(chain, es)
}

// In general the following checks are ordered from cheapest to most expensive
// in terms of computation and memory.
func processIssuance(chain Chain, es []factom.Entry) error {
	if len(es) == 0 {
		return nil
	}
	// The Identity may not have existed when this chain was first tracked.
	// Attempt to retrieve it.
	if err := chain.Identity.Get(); err != nil {
		return err
	}
	// If the Identity isn't yet populated then Issuance entries can't be
	// validated.
	if !chain.Identity.IsPopulated() {
		return nil
	}
	// If these entries were created in a lower block height than the
	// Identity entry, then none of them can be a valid Issuance entry.
	if es[0].Height < chain.Identity.Height {
		return nil
	}

	for i, e := range es {
		// If this entry was created before the Identity entry then it
		// can't be valid.
		if e.Timestamp.Before(chain.Identity.Timestamp) {
			continue
		}
		// Get the data for the entry.
		if err := e.Get(); err != nil {
			return fmt.Errorf("factom.Entry%#v.Get(): %v", e, err)
		}
		issuance := fat0.NewIssuance(e)
		if issuance.Valid(*chain.Identity.IDKey) != nil {
			continue
		}
		chains.Issue(issuance.ChainID, issuance)

		// Process remaining entries as transactions
		return processTransactions(chain, es[i+1:])
	}
	return nil
}

func processTransactions(chain Chain, es []factom.Entry) error {
	for _, e := range es {
		if err := e.Get(); err != nil {
			return fmt.Errorf("factom.Entry%#v.Get(): %v", e, err)
		}
		transaction := fat0.NewTransaction(e)
		if err := transaction.Valid(*chain.Identity.IDKey); err != nil {
			continue
		}
		// db.Apply?
		//if !chain.Apply(transaction) {
		//	continue
		//}
	}
	return nil
}
