package state

import (
	"fmt"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat0"
)

func (chain Chain) Process(eb factom.EBlock) error {
	defer Chains.set(eb.ChainID, &chain)

	// Load this Entry Block.
	if err := eb.Get(); err != nil {
		return fmt.Errorf("%#v.Get(): %v", eb, err)
	}
	if !eb.IsPopulated() {
		return fmt.Errorf("%#v.IsPopulated(): false", eb)
	}

	// Check if the EBlock represents a new chain.
	if eb.IsFirst() {
		// Load first entry of new chain.
		first := eb.Entries[0]
		if err := first.Get(); err != nil {
			return fmt.Errorf("%#v.Get: %v", first, err)
		}
		if !first.IsPopulated() {
			return fmt.Errorf("%#v.IsPopulated(): false", first)
		}

		// Ignore chains with NameIDs that don't match the fat0
		// pattern.
		if !fat0.ValidTokenNameIDs(first.ExtIDs) {
			chain.ignore()
			return nil
		}

		// Track this chain going forward.
		if err := chain.track(first); err != nil {
			return err
		}
		if len(eb.Entries) == 1 {
			return nil
		}
		// The first entry cannot be a valid Issuance entry, so discard
		// it and process the rest.
		eb.Entries = eb.Entries[1:]
	} else if !chain.IsTracked() {
		// Ignore chains that are not already tracked.
		chain.ignore()
		return nil
	}

	if err := chain.process(eb); err != nil {
		return err
	}
	return nil
}

func (chain *Chain) process(eb factom.EBlock) (err error) {
	defer func() {
		if err != nil {
			return
		}
		chain.saveHeight(eb.Height)
	}()
	es := eb.Entries
	if !chain.IsIssued() {
		return chain.processIssuance(es)
	}
	return chain.processTransactions(es)
}

// In general the following checks are ordered from cheapest to most expensive
// in terms of computation and memory.
func (chain *Chain) processIssuance(es []factom.Entry) error {
	if !chain.Identity.IsPopulated() {
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
			log.Debugf("Invalid Issuance Entry: %v, %v", e.Hash,
				"created before identity")
			continue
		}
		// Get the data for the entry.
		if err := e.Get(); err != nil {
			return fmt.Errorf("Entry%+v.Get(): %v", e, err)
		}
		if !e.IsPopulated() {
			return fmt.Errorf("Entry%+v.IsPopulated(): false", e)
		}
		issuance := fat0.NewIssuance(e)
		if err := issuance.Valid(chain.Identity.IDKey); err != nil {
			log.Debugf("Invalid Issuance Entry: %v, %v", e.Hash, err)
			continue
		}

		if err := chain.issue(issuance); err != nil {
			return err
		}

		// Process remaining entries as transactions
		return chain.processTransactions(es[i+1:])
	}
	return nil
}

func (chain *Chain) processTransactions(es []factom.Entry) error {
	if !chain.Identity.IsPopulated() {
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
	}
	for _, e := range es {
		if err := e.Get(); err != nil {
			return fmt.Errorf("Entry%v.Get(): %v", e, err)
		}
		if !e.IsPopulated() {
			return fmt.Errorf("%#v.IsPopulated(): false", e)
		}
		transaction := fat0.NewTransaction(e)
		if err := transaction.Valid(chain.Identity.IDKey); err != nil {
			log.Debugf("Invalid Transaction Entry: %v, %v", e.Hash, err)
			continue
		}
		if err := chain.apply(transaction); err != nil {
			return err
		}
	}
	return nil
}

func (chain *Chain) apply(transaction fat0.Transaction) (err error) {
	db := chain.Begin()
	defer chain.rollbackUnlessCommitted(*chain, &err)
	chain.DB = db

	entry, err := chain.createEntry(transaction.Entry.Entry)
	if entry == nil {
		// replayed transaction
		if err == nil {
			log.Debugf("Invalid Transaction Entry: %v, "+
				"replayed transaction",
				transaction.Hash)
		}
		return err
	}

	for rcdHash, amount := range transaction.Inputs {
		adr, err := chain.getAddress(&rcdHash)
		if err != nil {
			return err
		}
		if err := chain.DB.Model(&adr).
			Association("From").Append(entry).Error; err != nil {
			return err
		}
		if transaction.IsCoinbase() {
			if chain.Supply > 0 &&
				uint64(chain.Supply)-chain.Issued < amount {
				// insufficient coinbase supply
				log.Debugf("Invalid Transaction Entry: %v, "+
					"insufficient coinbase supply",
					entry.Hash)
				return nil
			}
			chain.Issued += amount
			if err := chain.saveMetadata(); err != nil {
				return err
			}
			break
		}
		if adr.Balance < amount {
			// insufficient balance
			log.Debugf("Invalid Transaction Entry: %v, "+
				"insufficient balance: %v",
				entry.Hash, adr.Address())
			return nil
		}
		adr.Balance -= amount
		if err := chain.Save(&adr).Error; err != nil {
			return err
		}
	}

	for rcdHash, amount := range transaction.Outputs {
		a, err := chain.getAddress(&rcdHash)
		if err != nil {
			return err
		}
		a.Balance += amount
		if err := chain.Save(&a).Error; err != nil {
			return err
		}
		if err := chain.DB.Model(&a).
			Association("To").Append(entry).Error; err != nil {
			return err
		}
	}
	log.Debugf("Valid Transaction Entry: %+v", transaction)

	return chain.Commit().Error
}
