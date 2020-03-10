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

package state

import (
	"context"
	"fmt"
	"os"

	"github.com/Factom-Asset-Tokens/fatd/internal/log"

	jsonrpc2 "github.com/AdamSLevy/jsonrpc2/v14"
	"github.com/Factom-Asset-Tokens/factom"
	"github.com/Factom-Asset-Tokens/factom/fat"
	"github.com/Factom-Asset-Tokens/factom/fat0"
	"github.com/Factom-Asset-Tokens/factom/fat1"
	"github.com/Factom-Asset-Tokens/fatd/internal/db"
	"github.com/Factom-Asset-Tokens/fatd/internal/db/address"
	"github.com/Factom-Asset-Tokens/fatd/internal/db/entry"
	"github.com/Factom-Asset-Tokens/fatd/internal/db/metadata"
	"github.com/Factom-Asset-Tokens/fatd/internal/db/nftoken"
)

type FATChain db.FATChain

var _ Chain = &FATChain{}

func (chain *FATChain) UpdateSidechainData(ctx context.Context, c *factom.Client) error {
	// Get Identity each time in case it wasn't populated before.
	if err := chain.Identity.Get(ctx, c); err != nil {
		// A jsonrpc2.Error indicates that the identity chain doesn't yet
		// exist, which we tolerate.
		if _, ok := err.(jsonrpc2.Error); !ok {
			return fmt.Errorf("factom.Identity.Get(): %w", err)
		}
		return nil
	}
	return metadata.UpdateIdentity(chain.Conn, chain.Identity)
}

func (chain *FATChain) ApplyEBlock(dbKeyMR *factom.Bytes32, eb factom.EBlock) error {
	return (*FactomChain)(&chain.FactomChain).ApplyEBlock(dbKeyMR, eb)
}

func (chain *FATChain) SetSync(height uint32, dbKeyMR *factom.Bytes32) error {
	return chain.ToFactomChain().SetSync(height, dbKeyMR)
}

func ToFATChain(chain Chain) (fatChain *FATChain, ok bool) {
	fatChain, ok = chain.(*FATChain)
	return
}
func (chain *FATChain) ToDBFATChain() *db.FATChain {
	return (*db.FATChain)(chain)
}
func (chain *FATChain) ToFactomChain() *db.FactomChain {
	return (*db.FactomChain)(&chain.FactomChain)
}

func (chain *FATChain) ApplyEntry(e factom.Entry) (eID int64, err error) {

	eID, err = (*FactomChain)(&chain.FactomChain).ApplyEntry(e)
	if err != nil {
		return
	}

	var txErr error
	entryType := "Tx"
	defer chain.save()(&txErr, &err)
	if !chain.IsIssued() {
		txErr, err = chain.ApplyIssuance(eID, e)
		entryType = "Issuance"
	} else {
		_, txErr, err = chain.ApplyTx(eID, e)
	}

	//if txErr != nil {
	//	chain.Log.Debugf("Invalid %v: %v %v", entryType, txErr, e.Hash)
	//} else {
	//	chain.Log.Debugf("Valid %v: %v", entryType, e.Hash)
	//}
	_ = entryType

	return
}
func (chain *FATChain) IsIssued() bool {
	return chain.Issuance.Entry.IsPopulated()
}

func (chain *FATChain) save() func(_, _ *error) {
	rollback := chain.Save()
	return func(txErr, err *error) {
		if *txErr != nil {
			rollback(txErr)
			return
		}
		rollback(err)
	}
}

func (chain *FATChain) ApplyIssuance(ei int64, e factom.Entry) (txErr, err error) {
	// The Identity must exist prior to issuance.
	if !chain.Identity.IsPopulated() ||
		e.Timestamp.Before(chain.Identity.Timestamp) {
		txErr = fmt.Errorf("Identity not set up prior to this Entry")
		return
	}

	issuance, txErr := fat.NewIssuance(e, (*factom.Bytes32)(chain.Identity.ID1Key))
	if txErr != nil {
		return
	}

	if err = metadata.SetInitEntryID(chain.Conn, ei); err != nil {
		return
	}
	chain.Issuance = issuance
	return
}

func (chain *FATChain) ApplyTx(eID int64, e factom.Entry) (tx interface{},
	txErr, err error) {

	valid, err := entry.CheckUniquelyValid(chain.Conn, eID, e.Hash)
	if err != nil {
		return
	}
	if !valid {
		txErr = fmt.Errorf("replay: hash previously marked valid")
		return
	}

	switch chain.Issuance.Type {
	case fat.TypeFAT0:
		_, txErr, err = chain.applyFAT0Tx(eID, e)
	case fat.TypeFAT1:
		_, txErr, err = chain.applyFAT1Tx(eID, e)
	default:
		panic(fmt.Errorf("invalid FAT Type %v", chain.Issuance.Type))
	}

	if txErr != nil || err != nil {
		return
	}

	if err = entry.SetValid(chain.Conn, eID); err != nil {
		return
	}

	return
}

func (chain *FATChain) ApplyFAT0Tx(eID int64, e factom.Entry) (tx fat0.Transaction,
	txErr, err error) {
	var txI interface{}
	txI, txErr, err = chain.ApplyTx(eID, e)
	tx = txI.(fat0.Transaction)
	return
}
func (chain *FATChain) applyFAT0Tx(eID int64, e factom.Entry) (tx fat0.Transaction,
	txErr, err error) {

	tx, txErr = fat0.NewTransaction(e, (*factom.Bytes32)(chain.Identity.ID1Key))
	if txErr != nil {
		return
	}

	if tx.IsCoinbase() {
		addIssued := tx.Inputs[fat.Coinbase()]
		if chain.Issuance.Supply > 0 &&
			int64(chain.NumIssued+addIssued) > chain.Issuance.Supply {
			txErr = fmt.Errorf("coinbase exceeds max supply")
			return
		}
		if err = chain.ToDBFATChain().AddNumIssued(addIssued); err != nil {
			return
		}
		if _, err = address.InsertTxRelation(
			chain.Conn, 1, eID, false); err != nil {
			return
		}
	} else {
		for adr, amount := range tx.Inputs {
			var ai int64
			ai, txErr, err = address.Sub(chain.Conn, &adr, amount)
			if err != nil || txErr != nil {
				return
			}
			if _, err = address.InsertTxRelation(
				chain.Conn, ai, eID, false); err != nil {
				return
			}
		}
	}

	for adr, amount := range tx.Outputs {
		var ai int64
		ai, err = address.Add(chain.Conn, &adr, amount)
		if err != nil {
			return
		}
		if _, err = address.InsertTxRelation(
			chain.Conn, ai, eID, true); err != nil {
			return
		}
	}

	return
}

func (chain *FATChain) ApplyFAT1Tx(eID int64, e factom.Entry) (tx fat1.Transaction,
	txErr, err error) {
	var txI interface{}
	txI, txErr, err = chain.ApplyTx(eID, e)
	tx = txI.(fat1.Transaction)
	return
}
func (chain *FATChain) applyFAT1Tx(eID int64, e factom.Entry) (tx fat1.Transaction,
	txErr, err error) {

	tx, txErr = fat1.NewTransaction(e, (*factom.Bytes32)(chain.Identity.ID1Key))
	if txErr != nil {
		return
	}

	if tx.IsCoinbase() {
		nfTkns := tx.Inputs[fat.Coinbase()]
		addIssued := uint64(len(nfTkns))
		if chain.Issuance.Supply > 0 &&
			int64(chain.NumIssued+addIssued) > chain.Issuance.Supply {
			txErr = fmt.Errorf("coinbase exceeds max supply")
			return
		}
		if err = (*db.FATChain)(chain).AddNumIssued(addIssued); err != nil {
			return
		}
		var adrTxID int64
		adrTxID, err = address.InsertTxRelation(chain.Conn,
			1, eID, false)
		if err != nil {
			return
		}
		for nfID := range nfTkns {
			// Insert the NFToken with the coinbase address as a
			// placeholder for the owner.
			txErr, err = nftoken.Insert(chain.Conn, nfID, 1, eID)
			if err != nil || txErr != nil {
				return
			}
			if err = nftoken.InsertTxRelation(
				chain.Conn, nfID, adrTxID); err != nil {
				return
			}
			metadata := tx.TokenMetadata[nfID]
			if len(metadata) == 0 {
				continue
			}
			if err = nftoken.SetMetadata(
				chain.Conn, nfID, metadata); err != nil {
				return
			}
		}
	} else {
		for adr, nfTkns := range tx.Inputs {
			var ai int64
			ai, txErr, err = address.Sub(
				chain.Conn, &adr, uint64(len(nfTkns)))
			if err != nil || txErr != nil {
				return
			}
			var adrTxID int64
			adrTxID, err = address.InsertTxRelation(
				chain.Conn, ai, eID, false)
			if err != nil {
				return
			}
			for nfTkn := range nfTkns {
				var ownerID int64
				ownerID, err = nftoken.SelectOwnerID(chain.Conn, nfTkn)
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
				if err = nftoken.InsertTxRelation(
					chain.Conn, nfTkn, adrTxID); err != nil {
					return
				}
			}
		}
	}

	for adr, nfTkns := range tx.Outputs {
		var ai int64
		ai, err = address.Add(chain.Conn, &adr, uint64(len(nfTkns)))
		if err != nil {
			return
		}
		var adrTxID int64
		adrTxID, err = address.InsertTxRelation(
			chain.Conn, ai, eID, true)
		if err != nil {
			return
		}
		for nfID := range nfTkns {
			if err = nftoken.SetOwner(chain.Conn, nfID, ai); err != nil {
				return
			}
			if err = nftoken.InsertTxRelation(
				chain.Conn, nfID, adrTxID); err != nil {
				return
			}
		}
	}

	return
}

func NewFATChain(ctx context.Context, c *factom.Client,
	dbPath, tokenID string,
	identityChainID, chainID *factom.Bytes32,
	networkID factom.NetworkID) (_ FATChain, err error) {

	chn, err := db.NewFATChain(ctx, dbPath, tokenID,
		chainID, identityChainID,
		networkID)
	if err != nil {
		err = fmt.Errorf("db.NewFATChain(): %w", err)
		return
	}

	chain := FATChain(chn)

	return chain, nil
}

func NewFATChainByEBlock(ctx context.Context, c *factom.Client,
	dbPath string, head factom.EBlock) (chain FATChain, err error) {

	log := log.New("chain", head.ChainID)
	log.Infof("Syncing new chain...")

	log.Info("Downloading all EBlocks...")
	eblocks, err := head.GetPrevAll(ctx, c)
	if err != nil {
		err = fmt.Errorf("factom.EBlock.GetPrevAll(): %w", err)
		return
	}

	firstEB := &eblocks[len(eblocks)-1]
	// Get DBlock Timestamp and KeyMR
	var dblock factom.DBlock
	dblock.Height = firstEB.Height
	if err = dblock.Get(ctx, c); err != nil {
		err = fmt.Errorf("factom.DBlock.Get(): %w", err)
		return
	}

	firstEB.SetTimestamp(dblock.Timestamp)

	if err = firstEB.Get(ctx, c); err != nil {
		err = fmt.Errorf("%#v.Get(): %w", firstEB, err)
		return
	}

	// Load first entry of new chain.
	first := &firstEB.Entries[0]
	if err = first.Get(ctx, c); err != nil {
		err = fmt.Errorf("%#v.Get(): %w", first, err)
		return
	}

	nameIDs := first.ExtIDs
	if !fat.ValidNameIDs(nameIDs) {
		err = fmt.Errorf("not a valid FAT chain: %v", head.ChainID)
		return
	}

	tokenID, identityChainID := fat.ParseTokenIssuer(nameIDs)

	var hasFirstEBlock bool
	chain, err = NewFATChain(ctx, c, dbPath,
		tokenID, &identityChainID,
		head.ChainID, dblock.NetworkID)
	if err != nil {
		err = fmt.Errorf("state.NewFATChain(): %w", err)
		return
	}
	defer func() {
		if err != nil {
			chain.Close()
		}
		if hasFirstEBlock {
			return
		}
		if err := os.Remove(dbPath + chain.DBFile); err != nil &&
			!os.IsNotExist(err) {
			chain.Log.Errorf("os.Remove(): %w", err)
		}
	}()

	if err = chain.UpdateSidechainData(ctx, c); err != nil {
		err = fmt.Errorf("state.Chain.UpdateSidechainData(): %w", err)
		return
	}

	chain.Log.Info("Syncing entries...")

	if err = firstEB.GetEntries(ctx, c); err != nil {
		err = fmt.Errorf("factom.EBlock.GetEntries(): %w", err)
		return
	}
	if err = Apply(&chain, dblock.KeyMR, *firstEB); err != nil {
		err = fmt.Errorf("state.Apply(): %w", err)
		return
	}
	hasFirstEBlock = true

	err = SyncEBlocks(ctx, c, &chain, eblocks[:len(eblocks)-1])
	return
}
