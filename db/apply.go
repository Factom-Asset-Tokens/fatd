// MIT License
//
// Copyright 2018 Canonical Ledgers, LLC
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
// IN THE SOFTWARE.

package db

import (
	"fmt"

	"crawshaw.io/sqlite/sqlitex"
	"github.com/Factom-Asset-Tokens/factom"
	"github.com/Factom-Asset-Tokens/fatd/db/addresses"
	"github.com/Factom-Asset-Tokens/fatd/db/eblocks"
	"github.com/Factom-Asset-Tokens/fatd/db/entries"
	"github.com/Factom-Asset-Tokens/fatd/db/metadata"
	"github.com/Factom-Asset-Tokens/fatd/db/nftokens"
	"github.com/Factom-Asset-Tokens/fatd/db/pegnet"
	"github.com/Factom-Asset-Tokens/fatd/fat"
	"github.com/Factom-Asset-Tokens/fatd/fat/fat0"
	"github.com/Factom-Asset-Tokens/fatd/fat/fat1"
	"github.com/pegnet/pegnet/modules/grader"
)

// pointer to chain, the entry itself, the eblock, and the position of entry in eblock's entrylist
// if the entry was applied on its own, position and total are -1
type applyFunc func(*Chain, int64, factom.Entry, factom.EBlock, int) (txErr, err error)

func (chain *Chain) Apply(dbKeyMR *factom.Bytes32, eb factom.EBlock) (err error) {
	// Ensure entire EBlock is applied atomically.
	defer sqlitex.Save(chain.Conn)(&err)
	defer func(chainCopy Chain) {
		if err != nil {
			// Reset chain on error
			*chain = chainCopy
		}
	}(*chain)

	chain.Head = eb

	// Insert latest EBlock.
	if err = eblocks.Insert(chain.Conn, eb, dbKeyMR); err != nil {
		return
	}

	// Insert each entry and attempt to apply it...
	for i, e := range eb.Entries {
		if _, err = chain.ApplyEntry(e, eb, i); err != nil {
			return
		}
	}
	return
}

func (chain *Chain) ApplyEntry(e factom.Entry, eb factom.EBlock, pos int) (txErr, err error) {
	ei, err := entries.Insert(chain.Conn, e, chain.Head.Sequence)
	if err != nil {
		return
	}
	return chain.apply(chain, ei, e, eb, pos)
}

var alwaysRollbackErr = fmt.Errorf("always rollback")

func (chain *Chain) applyIssuance(ei int64, e factom.Entry) (issueErr, err error) {
	issuance := fat.NewIssuance(e)
	rollback := sqlitex.Save(chain.Conn)
	chainCopy := *chain
	defer func() {
		if err != nil || issueErr != nil {
			rollback(&alwaysRollbackErr)
			// Reset chain on error
			*chain = chainCopy
			if err != nil {
				return
			}
			chain.Log.Debugf("Entry{%v}: invalid Issuance: %v",
				e.Hash, issueErr)
			return
		}
		rollback(&err) // commit
		chain.Log.Debugf("Valid Issuance Entry: %v %+v", e.Hash, issuance)
	}()
	// The Identity must exist prior to issuance.
	if !chain.Identity.IsPopulated() || e.Timestamp.Before(chain.Identity.Timestamp) {
		return
	}
	if issueErr = issuance.Validate(chain.ID1); issueErr != nil {
		return
	}
	if err = metadata.SetInitEntryID(chain.Conn, ei); err != nil {
		return
	}
	chain.Issuance = issuance
	chain.setApplyFunc()
	return
}

func (chain *Chain) setApplyFunc() {
	if !chain.Identity.IsPopulated() {
		chain.Type = fat.TypeFAT2
	} else if !chain.Issuance.IsPopulated() {
		chain.apply = func(chain *Chain, ei int64, e factom.Entry, eb factom.EBlock, pos int) (
			txErr, err error) {
			txErr, err = chain.applyIssuance(ei, e)
			return
		}
		return
	}
	// Adapt to match ApplyFunc.
	switch chain.Type {
	case fat0.Type:
		chain.apply = func(chain *Chain, ei int64, e factom.Entry, eb factom.EBlock, pos int) (
			txErr, err error) {
			_, txErr, err = chain.ApplyFAT0Tx(ei, e)
			return
		}
	case fat1.Type:
		chain.apply = func(chain *Chain, ei int64, e factom.Entry, eb factom.EBlock, pos int) (
			txErr, err error) {
			_, txErr, err = chain.ApplyFAT1Tx(ei, e)
			return
		}
	case fat.TypeFAT2:
		chain.apply = func(chain *Chain, ei int64, e factom.Entry, eb factom.EBlock, pos int) (
			txErr, err error) {
			err = chain.ApplyFAT2OPR(ei, e, eb, pos)
			return
		}
	default:
		panic("invalid FAT type")
	}
}

func (chain *Chain) Save(tx fat.Transaction) func(txErr, err *error) {
	rollback := sqlitex.Save(chain.Conn)
	chainCopy := *chain
	return func(txErr, err *error) {
		e := tx.FactomEntry()
		if *err != nil || *txErr != nil {
			rollback(&alwaysRollbackErr)
			// Reset chain on error
			*chain = chainCopy
			if *err != nil {
				return
			}
			chain.Log.Debugf("Entry{%v}: invalid %v Transaction: %v",
				e.Hash, chain.Type, *txErr)
			return
		}
		rollback(err)
		var cbStr string
		if tx.IsCoinbase() {
			cbStr = "Coinbase "
		}
		chain.Log.Debugf("Valid %v %vTransaction: %v %+v",
			chain.Type, cbStr, e.Hash, tx)
	}
}

func (chain *Chain) applyTx(ei int64, tx fat.Transaction) (txErr, err error) {
	if txErr = tx.Validate(chain.ID1); txErr != nil {
		return
	}
	e := tx.FactomEntry()
	valid, err := entries.CheckUniquelyValid(chain.Conn, ei, e.Hash)
	if err != nil {
		return
	}
	if !valid {
		txErr = fmt.Errorf("replay: hash previously marked valid")
		return
	}

	if err = entries.SetValid(chain.Conn, ei); err != nil {
		return
	}
	return
}

func (chain *Chain) ApplyFAT0Tx(ei int64, e factom.Entry) (tx *fat0.Transaction,
	txErr, err error) {
	tx = fat0.NewTransaction(e)
	defer chain.Save(tx)(&txErr, &err)

	if txErr, err = chain.applyTx(ei, tx); err != nil || txErr != nil {
		return
	}

	if tx.IsCoinbase() {
		addIssued := tx.Inputs[fat.Coinbase()]
		if chain.Supply > 0 && int64(chain.NumIssued+addIssued) > chain.Supply {
			txErr = fmt.Errorf("coinbase exceeds max supply")
			return
		}
		if err = chain.addNumIssued(addIssued); err != nil {
			return
		}
		if _, err = addresses.InsertTransactionRelation(
			chain.Conn, 1, ei, false); err != nil {
			return
		}
	} else {
		for adr, amount := range tx.Inputs {
			var ai int64
			ai, txErr, err = addresses.Sub(chain.Conn, &adr, amount)
			if err != nil || txErr != nil {
				return
			}
			if _, err = addresses.InsertTransactionRelation(
				chain.Conn, ai, ei, false); err != nil {
				return
			}
		}
	}

	for adr, amount := range tx.Outputs {
		var ai int64
		ai, err = addresses.Add(chain.Conn, &adr, amount)
		if err != nil {
			return
		}
		if _, err = addresses.InsertTransactionRelation(
			chain.Conn, ai, ei, true); err != nil {
			return
		}
	}

	return
}

var tmp grader.BlockGrader

func (chain *Chain) ApplyFAT2OPR(ei int64, e factom.Entry, eb factom.EBlock, pos int) (err error) {
	if !eb.IsPopulated() || pos < 0 { // 'processing' entry
		return
	}

	// beginning of every EBlock
	if pos == 0 {
		grader.InitLX()
		ver := uint8(1)
		if eb.Height >= 210330 {
			ver = uint8(2)
		}

		prev, err := pegnet.GetGrade(chain.Conn, eb.Sequence-1)
		if err != nil {
			return err
		}

		tmp, err = grader.NewGrader(ver, int32(eb.Height), prev)
		if err != nil {
			return err
		}
	}

	// Every Entry
	var extids [][]byte
	for _, x := range e.ExtIDs {
		extids = append(extids, []byte(x))
	}

	tmp.AddOPR(e.Hash[:], extids, []byte(e.Content))

	//fmt.Println(pos, len(eb.Entries))

	// After every EBlock
	if pos == len(eb.Entries)-1 {
		graded := tmp.Grade()
		winners := graded.WinnersShortHashes()
		//fmt.Println("winners", winners)
		err := pegnet.InsertGrade(chain.Conn, eb, winners)
		if err != nil {
			return err
		}

		oprs := graded.Winners()
		if len(oprs) > 0 {
			for _, t := range oprs[0].OPR.GetOrderedAssetsUint() {
				err = pegnet.InsertRate(chain.Conn, eb, t.Name, t.Value)
				if err != nil {
					return err
				}
			}
		}
		tmp = nil
		graded = nil
	}
	return
}

func (chain *Chain) ApplyFAT1Tx(ei int64, e factom.Entry) (tx *fat1.Transaction,
	txErr, err error) {
	tx = fat1.NewTransaction(e)
	defer chain.Save(tx)(&txErr, &err)

	if txErr, err = chain.applyTx(ei, tx); err != nil || txErr != nil {
		return
	}

	if tx.IsCoinbase() {
		nfTkns := tx.Inputs[fat.Coinbase()]
		addIssued := uint64(len(nfTkns))
		if chain.Supply > 0 && int64(chain.NumIssued+addIssued) > chain.Supply {
			txErr = fmt.Errorf("coinbase exceeds max supply")
			return
		}
		if err = chain.addNumIssued(addIssued); err != nil {
			return
		}
		var adrTxID int64
		adrTxID, err = addresses.InsertTransactionRelation(chain.Conn, 1, ei, false)
		if err != nil {
			return
		}
		for nfID := range nfTkns {
			// Insert the NFToken with the coinbase address as a
			// placeholder for the owner.
			txErr, err = nftokens.Insert(chain.Conn, nfID, 1, ei)
			if err != nil || txErr != nil {
				return
			}
			if err = nftokens.InsertTransactionRelation(
				chain.Conn, nfID, adrTxID); err != nil {
				return
			}
			metadata := tx.TokenMetadata[nfID]
			if len(metadata) == 0 {
				continue
			}
			if err = nftokens.SetMetadata(
				chain.Conn, nfID, metadata); err != nil {
				return
			}
		}
	} else {
		for adr, nfTkns := range tx.Inputs {
			var ai int64
			ai, txErr, err = addresses.Sub(
				chain.Conn, &adr, uint64(len(nfTkns)))
			if err != nil || txErr != nil {
				return
			}
			var adrTxID int64
			adrTxID, err = addresses.InsertTransactionRelation(
				chain.Conn, ai, ei, false)
			if err != nil {
				return
			}
			for nfTkn := range nfTkns {
				var ownerID int64
				ownerID, err = nftokens.SelectOwnerID(chain.Conn, nfTkn)
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
				if err = nftokens.InsertTransactionRelation(
					chain.Conn, nfTkn, adrTxID); err != nil {
					return
				}
			}
		}
	}

	for adr, nfTkns := range tx.Outputs {
		var ai int64
		ai, err = addresses.Add(chain.Conn, &adr, uint64(len(nfTkns)))
		if err != nil {
			return
		}
		var adrTxID int64
		adrTxID, err = addresses.InsertTransactionRelation(
			chain.Conn, ai, ei, true)
		if err != nil {
			return
		}
		for nfID := range nfTkns {
			if err = nftokens.SetOwner(chain.Conn, nfID, ai); err != nil {
				return
			}
			if err = nftokens.InsertTransactionRelation(
				chain.Conn, nfID, adrTxID); err != nil {
				return
			}
		}
	}

	return
}
