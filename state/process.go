package state

import (
	"fmt"

	jrpc "github.com/AdamSLevy/jsonrpc2/v11"
	"github.com/jinzhu/gorm"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat"
	"github.com/Factom-Asset-Tokens/fatd/fat/fat0"
	"github.com/Factom-Asset-Tokens/fatd/fat/fat1"
)

func (chain *Chain) Process(eb factom.EBlock) error {
	// Ensure changes to chain are saved in Chains.
	defer Chains.set(eb.ChainID, chain)

	// Load this Entry Block.
	if err := eb.Get(c); err != nil {
		return fmt.Errorf("%#v.Get(c): %v", eb, err)
	}

	// Check if the EBlock represents a new chain.
	if eb.IsFirst() {
		// Load first entry of new chain.
		first := eb.Entries[0]
		if err := first.Get(c); err != nil {
			return fmt.Errorf("%#v.Get(c): %v", first, err)
		}

		// Ignore chains with NameIDs that don't match the fat pattern.
		// FAT1 will also match here.
		if !fat.ValidTokenNameIDs(first.ExtIDs) {
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

	return chain.process(eb)
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
		if err := chain.Identity.Get(c); err != nil {
			if _, ok := err.(jrpc.Error); ok {
				return nil
			}
			return err
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
		if e.Timestamp.Time().Before(chain.Identity.Timestamp.Time()) {
			log.Debugf("Invalid Issuance Entry: %v, %v", e.Hash,
				"created before identity")
			continue
		}
		// Get the data for the entry.
		if err := e.Get(c); err != nil {
			return fmt.Errorf("Entry%+v.Get(c): %v", e, err)
		}
		issuance := fat.NewIssuance(e)
		if err := issuance.Valid(&chain.Identity.ID1); err != nil {
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
	for _, e := range es {
		if err := e.Get(c); err != nil {
			return fmt.Errorf("Entry%v.Get(c): %v", e, err)
		}
		switch chain.Type {
		case fat0.Type:
			transaction := fat0.NewTransaction(e)
			if err := transaction.Valid(chain.Identity.ID1); err != nil {
				log.Debugf("Invalid Transaction Entry: %v, %v", e.Hash, err)
				continue
			}
			if err := chain.applyFAT0(transaction); err != nil {
				return err
			}
		case fat1.Type:
			transaction := fat1.NewTransaction(e)
			if err := transaction.Valid(chain.Identity.ID1); err != nil {
				log.Debugf("Invalid Transaction Entry: %v, %v", e.Hash, err)
				continue
			}
			if err := chain.applyFAT1(transaction); err != nil {
				return err
			}
		}
	}
	return nil
}

func (chain *Chain) applyFAT0(transaction fat0.Transaction) (err error) {
	db := chain.Begin()
	defer chain.rollbackUnlessCommitted(*chain, &err)
	chain.DB = db

	entry, err := chain.createEntry(transaction.Entry.Entry)
	if err != nil {
		return err
	}
	if entry == nil {
		// replayed transaction
		log.Debugf("Invalid Transaction Entry: %v, replayed transaction",
			transaction.Hash)
		return nil
	}

	for rcdHash, amount := range transaction.Inputs {
		adr, err := chain.GetAddress(&rcdHash)
		if err != nil {
			return err
		}
		if err := chain.DB.Model(&adr).Association("From").
			Append(entry).Error; err != nil {
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
		a, err := chain.GetAddress(&rcdHash)
		if err != nil {
			return err
		}
		a.Balance += amount
		if err := chain.Save(&a).Error; err != nil {
			return err
		}
		if err := chain.DB.Model(&a).Association("To").
			Append(entry).Error; err != nil {
			return err
		}
	}
	log.Debugf("Valid Transaction Entry: %+v", transaction)

	return chain.Commit().Error
}

func (chain *Chain) applyFAT1(transaction fat1.Transaction) (err error) {
	db := chain.Begin()
	defer chain.rollbackUnlessCommitted(*chain, &err)
	chain.DB = db

	entry, err := chain.createEntry(transaction.Entry.Entry)
	if err != nil {
		return err
	}
	if entry == nil {
		// replayed transaction
		log.Debugf("Invalid Transaction Entry: %v, replayed transaction",
			transaction.Hash)
		return nil
	}

	allTkns := make(map[fat1.NFTokenID]NFToken, transaction.Inputs.NumNFTokenIDs())
	for rcdHash, tkns := range transaction.Inputs {
		adr, err := chain.GetAddress(&rcdHash)
		if err != nil {
			return err
		}
		if err := chain.DB.Model(&adr).Association("From").
			Append(entry).Error; err != nil {
			return err
		}
		if transaction.IsCoinbase() {
			if chain.Supply > 0 &&
				uint64(chain.Supply)-chain.Issued < uint64(len(tkns)) {
				// insufficient coinbase supply
				log.Debugf("Invalid Transaction Entry: %v, "+
					"insufficient coinbase supply",
					entry.Hash)
				return nil
			}
			chain.Issued += uint64(len(tkns))
			if err := chain.saveMetadata(); err != nil {
				return err
			}
			for tknID := range tkns {
				tkn, err := chain.createNFToken(tknID,
					transaction.TokenMetadata[tknID])
				if err != nil {
					return err
				}
				if tkn == nil {
					log.Debugf("Invalid Transaction Entry: %v, "+
						"NFTokenID(%v) already exists",
						entry.Hash, tknID)
					return nil
				}
				allTkns[tknID] = *tkn
			}
			break
		}
		if adr.Balance < uint64(len(tkns)) {
			// insufficient balance
			log.Debugf("Invalid Transaction Entry: %v, "+
				"insufficient balance: %v",
				entry.Hash, adr.Address())
			return nil
		}
		adr.Balance -= uint64(len(tkns))
		if err := chain.Save(&adr).Error; err != nil {
			return err
		}
		for tknID := range tkns {
			tkn := NFToken{NFTokenID: tknID, OwnerID: adr.ID}
			err := chain.GetNFToken(&tkn)
			if err == gorm.ErrRecordNotFound {
				log.Debugf("Invalid Transaction Entry: %v, "+
					"NFTokenID(%v) is not owned by %v",
					entry.Hash, tknID, rcdHash)
				return nil
			}
			if err != nil {
				return err
			}
			if err := chain.DB.Model(&tkn).Association("PreviousOwners").
				Append(&adr).Error; err != nil {
				return err
			}
			allTkns[tknID] = tkn
		}
	}

	for rcdHash, tkns := range transaction.Outputs {
		a, err := chain.GetAddress(&rcdHash)
		if err != nil {
			return err
		}
		a.Balance += uint64(len(tkns))
		if err := chain.Save(&a).Error; err != nil {
			return err
		}
		if err := chain.DB.Model(&a).Association("To").
			Append(entry).Error; err != nil {
			return err
		}
		for tknID := range tkns {
			tkn := allTkns[tknID]
			tkn.Owner = a
			tkn.OwnerID = a.ID
			if err := chain.Save(&tkn).Error; err != nil {
				return err
			}
			if err := chain.DB.Model(&tkn).Association("Transactions").
				Append(entry).Error; err != nil {
				return err
			}
		}
	}
	log.Debugf("Valid Transaction Entry: %T%+v", transaction, transaction)

	return chain.Commit().Error
}
