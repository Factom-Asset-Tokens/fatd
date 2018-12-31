package state

import (
	"database/sql"
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
			//log.Debugln("ignoring", first.ChainID)
			chain.ignore()
			return nil
		}

		log.Debugf("tracking Entry%+v", first)
		if chain.IsTracked() {
			log.Debugf("already tracked! EBlock%+v Chain%+v", eb, chain)
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
		//log.Debug("ignoring")
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
			continue
		}
		if err := chain.apply(transaction); err != nil {
			return err
		}
	}
	return nil
}

func (chain *Chain) apply(transaction fat0.Transaction) (err error) {
	dbEntry := newEntry(transaction.Entry.Entry)
	if !dbEntry.IsValid() {
		return fmt.Errorf("invalid hash: %#v", dbEntry)
	}
	savedDB := chain.DB
	savedIssued := chain.Issued
	chain.DB = chain.Begin()
	defer func() {
		// This rollback will silently fail if the db tx has already
		// been committed.
		rberr := chain.Rollback().Error
		chain.DB = savedDB
		if rberr == sql.ErrTxDone {
			// already committed
			return
		}
		if rberr != nil && err != nil {
			// Report other Rollback errors if there wasn't already
			// a returned error.
			err = rberr
			return
		}
		// complete rollback
		chain.Issued = savedIssued
	}()
	if chain.DB.Error != nil {
		return chain.DB.Error
	}
	if err := chain.Create(&dbEntry).Error; err != nil {
		return err
	}
	for rcdHash, amount := range transaction.Inputs {
		if transaction.IsCoinbase() {
			if chain.Supply > 0 &&
				uint64(chain.Supply)-chain.Issued < amount {
				// Insufficient supply for this coinbase tx.
				return nil
			}
			chain.Issued += amount
			if err := chain.saveMetadata(); err != nil {
				return err
			}
			continue
		}
		a, err := chain.getAddress(&rcdHash)
		if err != nil {
			return err
		}
		if a.Balance < amount {
			// insufficient funds
			return nil
		}
		a.Balance -= amount
		if err := chain.Save(&a).Error; err != nil {
			return err
		}
		if err := chain.DB.Model(&a).
			Association("From").Append(dbEntry).Error; err != nil {
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
			Association("To").Append(dbEntry).Error; err != nil {
			return err
		}
	}
	chain.Commit()
	return nil
}
