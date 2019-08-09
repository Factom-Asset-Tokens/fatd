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

func (chain *Chain) applyFAT0Tx(
	ei int64, e factom.Entry) (tx fat0.Transaction, err error) {
	tx = fat0.NewTransaction(e)
	valid, err := chain.applyTx(ei, e, &tx)
	if err != nil {
		return
	}
	if !valid {
		return
	}

	// Do not return, but log, any errors past this point as they are
	// related to being unable to apply a transaction.
	defer func() {
		if err != nil {
			chain.Log.Debugf("Entry{%v}: invalid %v transaction: %v",
				e.Hash, chain.Type, err)
			err = nil
		} else {
			var cbStr string
			if tx.IsCoinbase() {
				cbStr = "Coinbase "
			}
			chain.Log.Debugf("Valid %v %vTransaction: %v %+v",
				chain.Type, cbStr, tx.Hash, tx)
		}
	}()
	// But first rollback on any error.
	defer sqlitex.Save(chain.Conn)(&err)

	if err = chain.setEntryValid(ei); err != nil {
		return
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

	if tx.IsCoinbase() {
		addIssued := tx.Inputs[fat.Coinbase()]
		if chain.Supply > 0 &&
			int64(chain.NumIssued+addIssued) > chain.Supply {
			err = fmt.Errorf("coinbase exceeds max supply")
			return
		}
		if _, err = chain.insertAddressTransaction(1, ei, false); err != nil {
			return
		}
		err = chain.numIssuedAdd(addIssued)
		return
	}

	for adr, amount := range tx.Inputs {
		var ai int64
		ai, err = chain.addressSub(&adr, amount)
		if err != nil {
			return
		}
		if _, err = chain.insertAddressTransaction(ai, ei, false); err != nil {
			return
		}
	}
	return
}

func (chain *Chain) applyTx(ei int64, e factom.Entry, tx fat.Validator) (bool, error) {
	if err := tx.Validate(chain.ID1); err != nil {
		chain.Log.Debugf("Entry{%v}: invalid %v transaction: %v",
			e.Hash, chain.Type, err)
		return false, nil
	}
	valid, err := checkEntryUniqueValid(chain.Conn, ei, e.Hash)
	if err != nil {
		return false, err
	}
	if !valid {
		chain.Log.Debugf("Entry{%v}: invalid %v transaction: %v",
			e.Hash, chain.Type, "replay")
	}
	return valid, nil
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
			_, err := chain.applyFAT0Tx(ei, e)
			return err
		}
	case fat1.Type:
		chain.apply = func(ei int64, e factom.Entry) error {
			_, err := chain.applyFAT1Tx(ei, e)
			return err
		}
	default:
		panic("invalid type")
	}
}

func (chain *Chain) applyFAT1Tx(
	ei int64, e factom.Entry) (tx fat1.Transaction, err error) {
	tx = fat1.NewTransaction(e)
	valid, err := chain.applyTx(ei, e, &tx)
	if err != nil {
		return
	}
	if !valid {
		return
	}

	// Do not return, but log, any errors past this point as they are
	// related to being unable to apply a transaction.
	defer func() {
		if err != nil {
			chain.Log.Debugf("Entry{%v}: invalid %v transaction: %v",
				e.Hash, chain.Type, err)
			err = nil
		} else {
			var cbStr string
			if tx.IsCoinbase() {
				cbStr = "Coinbase "
			}
			chain.Log.Debugf("Valid %v %vTransaction: %v %+v",
				chain.Type, cbStr, tx.Hash, tx)
		}
	}()
	// But first rollback on any error.
	defer sqlitex.Save(chain.Conn)(&err)

	if err = chain.setEntryValid(ei); err != nil {
		return
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
			if err = chain.setNFTokenOwner(nfID, ai, ei); err != nil {
				return
			}
			if err = chain.insertNFTokenTransaction(
				nfID, adrTxID); err != nil {
				return
			}
		}
	}

	if tx.IsCoinbase() {
		nfTkns := tx.Inputs[fat.Coinbase()]
		addIssued := uint64(len(nfTkns))
		if chain.Supply > 0 &&
			int64(chain.NumIssued+addIssued) > chain.Supply {
			err = fmt.Errorf("coinbase exceeds max supply")
			return
		}
		var adrTxID int64
		adrTxID, err = chain.insertAddressTransaction(1, ei, false)
		if err != nil {
			return
		}
		for nfID := range nfTkns {
			if err = chain.insertNFTokenTransaction(
				nfID, adrTxID); err != nil {
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
		err = chain.numIssuedAdd(addIssued)
		return
	}

	for adr, nfTkns := range tx.Inputs {
		var ai int64
		ai, err = chain.addressSub(&adr, uint64(len(nfTkns)))
		if err != nil {
			return
		}
		var adrTxID int64
		adrTxID, err = chain.insertAddressTransaction(ai, ei, false)
		if err != nil {
			return
		}
		for nfID := range nfTkns {
			if err = chain.insertNFTokenTransaction(
				nfID, adrTxID); err != nil {
				return
			}
		}
	}
	return
}
