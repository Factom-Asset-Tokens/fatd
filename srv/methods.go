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

package srv

import (
	"bytes"
	"encoding/json"
	"fmt"

	jrpc "github.com/AdamSLevy/jsonrpc2/v11"
	"github.com/gocraft/dbr"
	"github.com/jinzhu/gorm"

	"github.com/Factom-Asset-Tokens/fatd/engine"
	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat"
	"github.com/Factom-Asset-Tokens/fatd/fat/fat0"
	"github.com/Factom-Asset-Tokens/fatd/fat/fat1"
	"github.com/Factom-Asset-Tokens/fatd/flag"
	"github.com/Factom-Asset-Tokens/fatd/state"
)

var c = flag.FactomClient

var jrpcMethods = jrpc.MethodMap{
	"get-issuance":           getIssuance(false),
	"get-issuance-entry":     getIssuance(true),
	"get-transaction":        getTransaction(false),
	"get-transaction-entry":  getTransaction(true),
	"get-transactions":       getTransactions(false),
	"get-transactions-entry": getTransactions(true),
	"get-balance":            getBalance,
	"get-nf-balance":         getNFBalance,
	"get-stats":              getStats,
	"get-nf-token":           getNFToken,
	"get-nf-tokens":          getNFTokens,

	"send-transaction": sendTransaction,

	"get-daemon-tokens":     getDaemonTokens,
	"get-daemon-properties": getDaemonProperties,
	"get-sync-status":       getSyncStatus,
}

type ResultGetIssuance struct {
	ParamsToken
	Hash      *factom.Bytes32 `json:"entryhash"`
	Timestamp int64           `json:"timestamp"`
	Issuance  fat.Issuance    `json:"issuance"`
}

func getIssuance(entry bool) jrpc.MethodFunc {
	return func(data json.RawMessage) interface{} {
		params := ParamsToken{}
		chain, err := validate(data, &params)
		if err != nil {
			return err
		}

		if entry {
			return chain.Issuance.Entry.Entry
		}
		return ResultGetIssuance{
			ParamsToken: ParamsToken{
				ChainID:       chain.ID,
				TokenID:       chain.Token,
				IssuerChainID: chain.Identity.ChainID,
			},
			Hash:      chain.Issuance.Hash,
			Timestamp: chain.Issuance.Timestamp.Unix(),
			Issuance:  chain.Issuance,
		}
	}
}

type ResultGetTransaction struct {
	Hash      *factom.Bytes32 `json:"entryhash"`
	Timestamp int64           `json:"timestamp"`
	Tx        interface{}     `json:"data"`
}

func getTransaction(getEntry bool) jrpc.MethodFunc {
	return func(data json.RawMessage) interface{} {
		params := ParamsGetTransaction{}
		chain, err := validate(data, &params)
		if err != nil {
			return err
		}

		entry, err := chain.GetEntry(params.Hash)
		if err == gorm.ErrRecordNotFound {
			return ErrorTransactionNotFound
		}
		if err != nil {
			panic(err)
		}

		if getEntry {
			return entry
		}

		switch chain.Type {
		case fat0.Type:
			tx := fat0.NewTransaction(entry)
			if err := tx.UnmarshalEntry(); err != nil {
				panic(err)
			}
			return ResultGetTransaction{
				Hash:      tx.Hash,
				Timestamp: tx.Timestamp.Unix(),
				Tx:        tx,
			}
		case fat1.Type:
			tx := fat1.NewTransaction(entry)
			if err := tx.UnmarshalEntry(); err != nil {
				panic(err)
			}
			return ResultGetTransaction{
				Hash:      tx.Hash,
				Timestamp: tx.Timestamp.Unix(),
				Tx:        tx,
			}
		default:
			panic(fmt.Sprintf("unknown FAT type: %v", chain.Type))
		}
	}
}

func getTransactions(getEntry bool) jrpc.MethodFunc {
	return func(data json.RawMessage) interface{} {
		params := ParamsGetTransactions{}
		chain, err := validate(data, &params)
		if err != nil {
			return err
		}
		if params.NFTokenID != nil && chain.Type != fat1.Type {
			err := ErrorTokenNotFound
			err.Data = "Token Chain is not FAT-1"
			return err
		}

		// Lookup Txs
		entries, err := chain.GetEntries(params.StartHash,
			params.Addresses, params.NFTokenID,
			params.ToFrom, params.Order,
			*params.Page, *params.Limit)
		if err == dbr.ErrNotFound {
			return ErrorTransactionNotFound
		}
		if err != nil {
			panic(err)
		}
		if len(entries) == 0 {
			return ErrorTransactionNotFound
		}
		if getEntry {
			// Omit the ChainID from the response since the client
			// already knows it.
			for i := range entries {
				entries[i].ChainID = nil
			}
			return entries
		}

		switch chain.Type {
		case fat0.Type:
			txs := make([]ResultGetTransaction, len(entries))
			for i := range txs {
				tx := fat0.NewTransaction(entries[i])
				if err := tx.UnmarshalEntry(); err != nil {
					panic(err)
				}
				txs[i].Hash = entries[i].Hash
				txs[i].Timestamp = entries[i].Timestamp.Unix()
				txs[i].Tx = tx
			}
			return txs
		case fat1.Type:
			txs := make([]ResultGetTransaction, len(entries))
			for i := range txs {
				tx := fat1.NewTransaction(entries[i])
				if err := tx.UnmarshalEntry(); err != nil {
					panic(err)
				}
				txs[i].Hash = entries[i].Hash
				txs[i].Timestamp = entries[i].Timestamp.Unix()
				txs[i].Tx = tx
			}
			return txs
		default:
			panic(fmt.Sprintf("unknown FAT type: %v", chain.Type))
		}

	}
}

func getBalance(data json.RawMessage) interface{} {
	params := ParamsGetBalance{}
	chain, err := validate(data, &params)
	if err != nil {
		return err
	}

	adr, err := chain.GetAddress(params.Address)
	if err != nil {
		panic(err)
	}
	return adr.Balance
}

func getNFBalance(data json.RawMessage) interface{} {
	params := ParamsGetNFBalance{}
	chain, err := validate(data, &params)
	if err != nil {
		return err
	}

	if chain.Type != fat1.Type {
		err := ErrorTokenNotFound
		err.Data = "Token Chain is not FAT-1"
		return err
	}

	tkns, err := chain.GetNFTokensForOwner(params.Address,
		*params.Page, *params.Limit, params.Order)
	if err != nil {
		panic(err)
	}

	// Empty fat1.NFTokens cannot be marshalled by design so substitute an
	// empty slice.
	if len(tkns) == 0 {
		return []struct{}{}
	}

	return tkns
}

type ResultGetStats struct {
	ParamsToken
	Issuance                 *fat.Issuance
	CirculatingSupply        uint64 `json:"circulating"`
	Burned                   uint64 `json:"burned"`
	Transactions             int    `json:"transactions"`
	IssuanceTimestamp        int64  `json:"issuancets"`
	LastTransactionTimestamp int64  `json:"lasttxts,omitempty"`
}

var coinbaseRCDHash = fat.Coinbase()

func getStats(data json.RawMessage) interface{} {
	params := ParamsToken{}
	chain, err := validate(data, &params)
	if err != nil {
		return err
	}

	coinbase, err := chain.GetAddress(&coinbaseRCDHash)
	if err != nil {
		panic(err)
	}
	burned := coinbase.Balance
	txs, err := chain.GetEntries(nil, nil, nil, "", "", 0, 0)
	if err != nil {
		panic(err)
	}

	var lastTxTs int64
	if len(txs) > 0 {
		lastTxTs = txs[len(txs)-1].Timestamp.Unix()
	}
	res := ResultGetStats{
		CirculatingSupply:        chain.Issued - burned,
		Burned:                   burned,
		Transactions:             len(txs),
		IssuanceTimestamp:        chain.Issuance.Timestamp.Unix(),
		LastTransactionTimestamp: lastTxTs,
	}
	if chain.IsIssued() {
		res.Issuance = &chain.Issuance
	}
	res.ChainID = chain.ID
	res.TokenID = chain.Token
	res.IssuerChainID = chain.Issuer
	return res
}

type ResultGetNFToken struct {
	NFTokenID fat1.NFTokenID    `json:"id"`
	Owner     *factom.FAAddress `json:"owner"`
	Metadata  json.RawMessage   `json:"metadata,omitempty"`
}

func getNFToken(data json.RawMessage) interface{} {
	params := ParamsGetNFToken{}
	chain, err := validate(data, &params)
	if err != nil {
		return err
	}

	if chain.Type != fat1.Type {
		err := ErrorTokenNotFound
		err.Data = "Token Chain is not FAT-1"
		return err
	}

	tkn := state.NFToken{NFTokenID: *params.NFTokenID}
	if err := chain.GetNFToken(&tkn); err != nil {
		if err == gorm.ErrRecordNotFound {
			err := ErrorTokenNotFound
			err.Data = "No such NFTokenID has been issued"
			return err
		}
		panic(err)
	}
	return ResultGetNFToken{
		NFTokenID: tkn.NFTokenID,
		Metadata:  tkn.Metadata,
		Owner:     tkn.Owner.RCDHash,
	}
}

func getNFTokens(data json.RawMessage) interface{} {
	params := ParamsGetAllNFTokens{}
	chain, err := validate(data, &params)
	if err != nil {
		return err
	}

	if chain.Type != fat1.Type {
		err := ErrorTokenNotFound
		err.Data = "Token Chain is not FAT-1"
		return err
	}

	tkns, err := chain.GetAllNFTokens(*params.Page, *params.Limit, params.Order)
	if err != nil {
		panic(err)
	}

	res := make([]ResultGetNFToken, len(tkns))
	for i, tkn := range tkns {
		res[i].NFTokenID = tkn.NFTokenID
		res[i].Metadata = tkn.Metadata
		res[i].Owner = tkn.Owner.RCDHash
	}

	return res
}

func sendTransaction(data json.RawMessage) interface{} {
	var zero factom.EsAddress
	if flag.EsAdr == zero {
		return ErrorNoEC
	}
	params := ParamsSendTransaction{}
	chain, err := validate(data, &params)
	if err != nil {
		return err
	}

	entry := params.Entry()
	hash, _ := entry.ComputeHash()
	transaction, err := chain.GetEntry(&hash)
	if transaction.IsPopulated() {
		err := ErrorInvalidTransaction
		err.Data = "duplicate transaction"
		return err
	}
	if err != gorm.ErrRecordNotFound {
		panic(err)
	}

	switch chain.Type {
	case fat0.Type:
		if err := validFAT0Transaction(chain, entry); err != nil {
			return err
		}
	case fat1.Type:
		if err := validFAT1Transaction(chain, entry); err != nil {
			return err
		}
	default:
		panic("invalid FAT type")
	}

	balance, err := flag.ECAdr.GetBalance(c)
	if err != nil {
		panic(err)
	}
	cost, err := entry.Cost()
	if err != nil {
		rerr := ErrorInvalidTransaction
		rerr.Data = err
		return rerr
	}
	if balance < uint64(cost) {
		return ErrorNoEC
	}
	txID, err := entry.ComposeCreate(c, flag.EsAdr)
	if err != nil {
		log.Error(err)
		panic(err)
	}

	return struct {
		ChainID *factom.Bytes32 `json:"chainid"`
		TxID    *factom.Bytes32 `json:"txid"`
		Hash    *factom.Bytes32 `json:"entryhash"`
	}{ChainID: chain.ID, TxID: txID, Hash: entry.Hash}
}

func validFAT0Transaction(chain *state.Chain, entry factom.Entry) error {
	tx := fat0.NewTransaction(entry)
	rpcErr := ErrorInvalidTransaction
	if err := tx.Valid(chain.ID1); err != nil {
		rpcErr.Data = err.Error()
		return rpcErr
	}

	// check balances
	if tx.IsCoinbase() {
		if tx.Inputs.Sum() > uint64(chain.Supply)-chain.Issued {
			rpcErr.Data = "insufficient coinbase supply"
			return rpcErr
		}
		return nil
	}
	for rcdHash, amount := range tx.Inputs {
		adr, err := chain.GetAddress(&rcdHash)
		if err != nil {
			log.Error(err)
			panic(err)
		}
		if amount > adr.Balance {
			rpcErr.Data = fmt.Sprintf("insufficient balance: %v", rcdHash)
			return rpcErr
		}
	}
	return nil
}

func validFAT1Transaction(chain *state.Chain, entry factom.Entry) error {
	tx := fat1.NewTransaction(entry)
	rpcErr := ErrorInvalidTransaction
	if err := tx.Valid(chain.ID1); err != nil {
		rpcErr.Data = err.Error()
		return rpcErr
	}

	for rcdHash, tkns := range tx.Inputs {
		adr, err := chain.GetAddress(&rcdHash)
		if err != nil {
			log.Error(err)
			panic(err)
		}
		if tx.IsCoinbase() {
			if chain.Supply > 0 &&
				uint64(chain.Supply)-chain.Issued < uint64(len(tkns)) {
				// insufficient coinbase supply
				rpcErr.Data = "insufficient coinbase supply"
				return rpcErr
			}
			for tknID := range tkns {
				tkn := state.NFToken{NFTokenID: tknID}
				err := chain.GetNFToken(&tkn)
				if err == nil {
					rpcErr.Data = fmt.Sprintf(
						"NFTokenID(%v) already exists", tknID)
					return rpcErr
				}
				if err != gorm.ErrRecordNotFound {
					log.Error(err)
					panic(err)
				}
			}
			break
		}
		if adr.Balance < uint64(len(tkns)) {
			rpcErr.Data = fmt.Sprintf("insufficient balance: %v", rcdHash)
			return rpcErr
		}
		for tknID := range tkns {
			tkn := state.NFToken{NFTokenID: tknID, OwnerID: adr.ID}
			err := chain.GetNFToken(&tkn)
			if err == gorm.ErrRecordNotFound {
				rpcErr.Data = fmt.Sprintf(
					"NFTokenID(%v) is not owned by %v",
					tknID, rcdHash)
				return rpcErr
			}
			if err != nil {
				log.Error(err)
				panic(err)
			}
		}
	}

	return nil
}

func getDaemonTokens(data json.RawMessage) interface{} {
	if data != nil {
		return ParamsErrorNoParams
	}

	issuedIDs := state.Chains.GetIssued()
	chains := make([]ParamsToken, len(issuedIDs))
	for i, chainID := range issuedIDs {
		chain := state.Chains.Get(chainID)
		chains[i].ChainID = chainID
		chains[i].TokenID = chain.Token
		chains[i].IssuerChainID = chain.Issuer
	}
	return chains
}

type ResultGetDaemonProperties struct {
	FatdVersion string `json:"fatdversion"`
	APIVersion  string `json:"apiversion"`
}

func getDaemonProperties(data json.RawMessage) interface{} {
	if data != nil {
		return ParamsErrorNoParams
	}
	return ResultGetDaemonProperties{FatdVersion: flag.Revision, APIVersion: APIVersion}
}

type ResultGetSyncStatus struct {
	Sync    uint32 `json:"syncheight"`
	Current uint32 `json:"factomheight"`
}

func getSyncStatus(data json.RawMessage) interface{} {
	sync, current := engine.GetSyncStatus()
	return ResultGetSyncStatus{Sync: sync, Current: current}
}

func validate(data json.RawMessage, params Params) (*state.Chain, error) {
	if data == nil {
		return nil, params.Error()
	}
	if err := unmarshalStrict(data, params); err != nil {
		return nil, jrpc.InvalidParams(err.Error())
	}
	chainID := params.ValidChainID()
	if chainID == nil || !params.IsValid() {
		return nil, params.Error()
	}
	chain := state.Chains.Get(chainID)
	if !chain.IsIssued() {
		return nil, ErrorTokenNotFound
	}
	return &chain, nil
}

func unmarshalStrict(data []byte, v interface{}) error {
	b := bytes.NewBuffer(data)
	d := json.NewDecoder(b)
	d.DisallowUnknownFields()
	return d.Decode(v)
}
