package db

import (
	"fmt"

	"crawshaw.io/sqlite/sqlitex"
	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat"
	"github.com/Factom-Asset-Tokens/fatd/fat/fat0"
	"github.com/Factom-Asset-Tokens/fatd/fat/fat1"
)

type applyFunc func(int64, factom.Entry) error

func (chain *Chain) Apply(dbKeyMR *factom.Bytes32, eb factom.EBlock) (err error) {
	// Ensure entire EBlock is applied atomically.
	defer sqlitex.Save(chain.Conn)(&err)

	// Save latest EBlock.
	if err = chain.insertEBlock(eb, dbKeyMR); err != nil {
		return
	}
	chain.Head = eb

	// Save and apply each entry.
	for _, e := range eb.Entries {
		var ei int64
		ei, err = chain.insertEntry(e, chain.Head.Sequence)
		if err != nil {
			return
		}
		if err = chain.apply(ei, e); err != nil {
			return
		}
	}
	return
}

func (chain *Chain) applyIssuance(ei int64, e factom.Entry) error {
	// The Identity must exist prior to issuance.
	if !chain.Identity.IsPopulated() ||
		e.Timestamp.Before(chain.Identity.Timestamp) {
		chain.Log.Debugf("Entry{%v}: invalid issuance: %v", e.Hash,
			"created before identity")
		return nil
	}
	issuance := fat.NewIssuance(e)
	if err := issuance.Validate(chain.ID1); err != nil {
		chain.Log.Debugf("Entry{%v}: invalid issuance: %v", e.Hash, err)
		return nil
	}
	// check sig and is valid
	if err := chain.setInitEntryID(ei); err != nil {
		return err
	}
	chain.Issuance = issuance
	chain.setApplyFunc()
	chain.Log.Debugf("Valid Issuance Entry: %v %+v", e.Hash, issuance)
	return nil
}

func applyTxRollback(chain *Chain, e factom.Entry, tx interface {
	IsCoinbase() bool
},
	rollback func(*error), txErr, err *error) {
	if *err != nil {
		rollback(err)
	} else if *txErr != nil {
		rollback(txErr)
		chain.Log.Debugf("Entry{%v}: invalid %v transaction: %v",
			e.Hash, chain.Type, txErr)
	} else {
		var cbStr string
		if tx.IsCoinbase() {
			cbStr = "Coinbase "
		}
		chain.Log.Debugf("Valid %v %vTransaction: %v %+v",
			chain.Type, cbStr, e.Hash, tx)
	}
}

func (chain *Chain) ApplyFAT0Tx(ei int64, e factom.Entry) (tx fat0.Transaction,
	txErr, err error) {
	tx = fat0.NewTransaction(e)
	txErr, err = chain.applyTx(ei, e, &tx)
	if err != nil || txErr != nil {
		return
	}

	rollback := sqlitex.Save(chain.Conn)
	defer applyTxRollback(chain, e, tx, rollback, &txErr, &err)

	if err = chain.setEntryValid(ei); err != nil {
		return
	}

	if tx.IsCoinbase() {
		addIssued := tx.Inputs[fat.Coinbase()]
		if chain.Supply > 0 && int64(chain.NumIssued+addIssued) > chain.Supply {
			txErr = fmt.Errorf("coinbase exceeds max supply")
			return
		}
		if err = chain.numIssuedAdd(addIssued); err != nil {
			return
		}
		if _, err = chain.insertAddressTransaction(1, ei, false); err != nil {
			return
		}
	} else {
		for adr, amount := range tx.Inputs {
			var ai int64
			ai, txErr, err = chain.addressSub(&adr, amount)
			if err != nil || txErr != nil {
				return
			}
			if _, err = chain.insertAddressTransaction(ai, ei,
				false); err != nil {
				return
			}
		}
	}

	for adr, amount := range tx.Outputs {
		var ai int64
		ai, err = chain.addressAdd(&adr, amount)
		if err != nil {
			return
		}
		if _, err = chain.insertAddressTransaction(ai, ei, true); err != nil {
			return
		}
	}

	return
}

func (chain *Chain) applyTx(ei int64, e factom.Entry, tx fat.Validator) (error, error) {
	if err := tx.Validate(chain.ID1); err != nil {
		chain.Log.Debugf("Entry{%v}: invalid %v transaction: %v",
			e.Hash, chain.Type, err)
		return err, nil
	}
	valid, err := checkEntryUniqueValid(chain.Conn, ei, e.Hash)
	if err != nil {
		return nil, err
	}
	if !valid {
		chain.Log.Debugf("Entry{%v}: invalid %v transaction: %v",
			e.Hash, chain.Type, "replay: hash previously marked valid")
		return fmt.Errorf("replay: hash previously marked valid"), nil
	}
	return nil, nil
}

func (chain *Chain) setApplyFunc() {
	if !chain.Issuance.IsPopulated() {
		chain.apply = chain.applyIssuance
		return
	}
	// Adapt to match ApplyFunc.
	switch chain.Type {
	case fat0.Type:
		chain.apply = func(ei int64, e factom.Entry) error {
			_, _, err := chain.ApplyFAT0Tx(ei, e)
			return err
		}
	case fat1.Type:
		chain.apply = func(ei int64, e factom.Entry) error {
			_, _, err := chain.ApplyFAT1Tx(ei, e)
			return err
		}
	default:
		panic("invalid type")
	}
}

func (chain *Chain) ApplyFAT1Tx(ei int64, e factom.Entry) (tx fat1.Transaction,
	txErr, err error) {
	tx = fat1.NewTransaction(e)
	txErr, err = chain.applyTx(ei, e, &tx)
	if err != nil {
		return
	}
	if txErr != nil {
		return
	}

	rollback := sqlitex.Save(chain.Conn)
	defer applyTxRollback(chain, e, tx, rollback, &txErr, &err)

	if err = chain.setEntryValid(ei); err != nil {
		return
	}

	if tx.IsCoinbase() {
		nfTkns := tx.Inputs[fat.Coinbase()]
		addIssued := uint64(len(nfTkns))
		if chain.Supply > 0 && int64(chain.NumIssued+addIssued) > chain.Supply {
			txErr = fmt.Errorf("coinbase exceeds max supply")
			return
		}
		if err = chain.numIssuedAdd(addIssued); err != nil {
			return
		}
		var adrTxID int64
		adrTxID, err = chain.insertAddressTransaction(1, ei, false)
		if err != nil {
			return
		}
		for nfID := range nfTkns {
			// Insert the NFToken with the coinbase address as a
			// placeholder for the owner.
			txErr, err = chain.insertNFToken(nfID, 1, ei)
			if err != nil || txErr != nil {
				return
			}
			if err = chain.insertNFTokenTransaction(nfID, adrTxID); err != nil {
				return
			}
			metadata := tx.TokenMetadata[nfID]
			if len(metadata) == 0 {
				continue
			}
			if err = chain.setNFTokenMetadata(nfID, metadata); err != nil {
				return
			}
		}
	} else {
		for adr, nfTkns := range tx.Inputs {
			var ai int64
			ai, txErr, err = chain.addressSub(&adr, uint64(len(nfTkns)))
			if err != nil || txErr != nil {
				return
			}
			var adrTxID int64
			adrTxID, err = chain.insertAddressTransaction(ai, ei, false)
			if err != nil {
				return
			}
			for nfTkn := range nfTkns {
				var ownerID int64
				ownerID, err = SelectNFTokenOwnerID(chain.Conn, nfTkn)
				if err != nil {
					return
				}
				if ownerID == -1 {
					txErr = fmt.Errorf("no such NFToken{%v}", nfTkn)
					return
				}
				if ownerID != ai {
					txErr = fmt.Errorf("NFToken{%v} not owned by %v",
						nfTkn, adr)
					return
				}
				if err = chain.insertNFTokenTransaction(
					nfTkn, adrTxID); err != nil {
					return
				}
			}
		}
	}

	for adr, nfTkns := range tx.Outputs {
		var ai int64
		ai, err = chain.addressAdd(&adr, uint64(len(nfTkns)))
		if err != nil {
			return
		}
		var adrTxID int64
		adrTxID, err = chain.insertAddressTransaction(ai, ei, true)
		if err != nil {
			return
		}
		for nfID := range nfTkns {
			if err = chain.setNFTokenOwner(nfID, ai); err != nil {
				return
			}
			if err = chain.insertNFTokenTransaction(nfID, adrTxID); err != nil {
				return
			}
		}
	}

	return
}