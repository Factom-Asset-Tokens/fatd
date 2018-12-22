package state

import (
	"database/sql"
	"fmt"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat0"
	"github.com/jinzhu/gorm"
)

type Chain struct {
	ID *factom.Bytes32
	ChainStatus
	fat0.Identity
	fat0.Issuance
	metadata
	*gorm.DB
}

func (chain *Chain) Ignore() {
	chain.ID = nil
	chain.ChainStatus = ChainStatusIgnored
}
func (chain *Chain) Track(nameIDs []factom.Bytes) error {
	token := string(nameIDs[1])
	identityChainID := factom.NewBytes32(nameIDs[3])

	chain.ChainStatus = ChainStatusTracked
	chain.metadata = metadata{Token: token, Issuer: identityChainID}
	chain.Identity = fat0.Identity{ChainID: identityChainID}

	if err := chain.setupDB(); err != nil {
		return err
	}

	return nil
}
func (chain *Chain) Issue(issuance fat0.Issuance) error {
	chain.Issuance = issuance
	if err := chain.saveIssuance(); err != nil {
		return err
	}
	return nil
}

func (chain *Chain) ProcessEBlock(eb factom.EBlock) error {
	defer chain.saveHeight(eb.Height)
	if !chain.IsIssued() {
		log.Debug("is not issued")
		return chain.processIssuance(eb.Entries)
	}
	log.Debug("is issued")
	return chain.processTransactions(eb.Entries)
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
			return fmt.Errorf("%#v.Get(): %v", e, err)
		}
		if !e.IsPopulated() {
			return fmt.Errorf("%#v.IsPopulated(): false", e)
		}
		issuance := fat0.NewIssuance(e)
		if issuance.Valid(*chain.Identity.IDKey) != nil {
			// ignore invalid entries
			continue
		}

		log.Debug("issue")
		if err := chain.Issue(issuance); err != nil {
			return err
		}

		// Process remaining entries as transactions
		return chain.processTransactions(es[i+1:])
	}
	return nil
}

func (chain *Chain) processTransactions(es []factom.Entry) error {
	for _, e := range es {
		if err := e.Get(); err != nil {
			return fmt.Errorf("%#v.Get(): %v", e, err)
		}
		if !e.IsPopulated() {
			return fmt.Errorf("%#v.IsPopulated(): false", e)
		}
		transaction := fat0.NewTransaction(e)
		if err := transaction.Valid(*chain.Identity.IDKey); err != nil {
			continue
		}
		if err := chain.Apply(transaction); err != nil {
			return err
		}
	}
	return nil
}

func (chain *Chain) Apply(transaction fat0.Transaction) (err error) {
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
		a := address{}
		if err := chain.Where(
			&address{RCDHash: &rcdHash}).First(&a).Error; err != nil &&
			err != gorm.ErrRecordNotFound {
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
		a := address{}
		if err := chain.Where(
			&address{RCDHash: &rcdHash}).First(&a).Error; err != nil &&
			err != gorm.ErrRecordNotFound {
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
