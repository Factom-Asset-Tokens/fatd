package state

import (
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
	chain.ChainStatus = ChainStatusIgnored
	chains.Set(chain)
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
	chains.Set(chain)
	return nil
}
func (chain *Chain) Issue(issuance fat0.Issuance) error {
	chain.Issuance = issuance
	if err := chain.saveIssuance(); err != nil {
		return err
	}
	return nil
}

// setupDB a database for a given token chain.
func (chain *Chain) setupDB() error {
	fname := fmt.Sprintf("%v%v", chain.ID, dbFileExtension)
	var err error
	if chain.DB, err = open(fname); err != nil {
		return err
	}
	// Ensure the db gets closed if there are any issues.
	defer func() {
		if err != nil {
			chain.Close()
		}
	}()
	if err := chain.Save(&chain.metadata).Error; err != nil {
		return err
	}
	return nil
}

func (chain *Chain) loadMetadata() error {
	var metadataTableCount int
	if err := chain.DB.Model(&metadata{}).Count(&metadataTableCount).Error; err != nil {
		return err
	}
	if metadataTableCount != 1 {
		return fmt.Errorf(`table "metadata" must have exactly one row`)
	}
	if err := chain.First(&chain.metadata).Error; err != nil {
		return err
	}
	if !fat0.ValidTokenNameIDs(fat0.NameIDs(chain.Token, chain.Issuer)) ||
		*chain.ID != fat0.ChainID(chain.Token, chain.Issuer) {
		return fmt.Errorf("corrupted metadata table for chain %v", chain.ID)
	}
	return nil
}

func (chain *Chain) loadIssuance() error {
	e := entry{}
	if err := chain.First(&e).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		return err
	}
	if !e.IsValid() {
		return fmt.Errorf("corrupted entry hash")
	}
	chain.Issuance = fat0.NewIssuance(e.Entry())
	if err := chain.Issuance.UnmarshalEntry(); err != nil {
		return err
	}
	chain.ChainStatus = ChainStatusIssued
	return nil
}

func (chain *Chain) saveIssuance() error {
	if chain.IsIssued() {
		return fmt.Errorf("already issued")
	}
	var entriesTableCount int
	if err := chain.DB.Model(&entry{}).Count(&entriesTableCount).Error; err != nil {
		return err
	}
	if entriesTableCount != 0 {
		return fmt.Errorf(`table "entries" must be empty prior to issuance`)
	}

	if err := chain.Save(newEntry(chain.Issuance.Entry.Entry)).Error; err != nil {
		return err
	}
	chain.ChainStatus = ChainStatusIssued
	return nil
}

func (chain Chain) ProcessEntries(es []factom.Entry) error {
	if !chain.IsIssued() {
		return chain.processIssuance(es)
	}
	return chain.processTransactions(es)
}

// In general the following checks are ordered from cheapest to most expensive
// in terms of computation and memory.
func (chain *Chain) processIssuance(es []factom.Entry) error {
	if len(es) == 0 {
		return nil
	}
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
			return fmt.Errorf("not populated: %v", e)
		}
		issuance := fat0.NewIssuance(e)
		if issuance.Valid(*chain.Identity.IDKey) != nil {
			// ignore invalid entries
			continue
		}

		if err := chain.Issue(issuance); err != nil {
			return err
		}

		// Process remaining entries as transactions
		return chain.processTransactions(es[i+1:])
	}
	return nil
}

func (chain Chain) processTransactions(es []factom.Entry) error {
	for _, e := range es {
		if err := e.Get(); err != nil {
			return fmt.Errorf("%#v.Get(): %v", e, err)
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
