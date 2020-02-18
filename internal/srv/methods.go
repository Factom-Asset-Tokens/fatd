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
	"context"
	"encoding/json"
	"fmt"

	jsonrpc2 "github.com/AdamSLevy/jsonrpc2/v14"

	"github.com/Factom-Asset-Tokens/factom"
	"github.com/Factom-Asset-Tokens/factom/fat"
	"github.com/Factom-Asset-Tokens/factom/fat0"
	"github.com/Factom-Asset-Tokens/factom/fat1"
	"github.com/Factom-Asset-Tokens/fatd/api"
	"github.com/Factom-Asset-Tokens/fatd/internal/db/address"
	"github.com/Factom-Asset-Tokens/fatd/internal/db/entry"
	"github.com/Factom-Asset-Tokens/fatd/internal/db/nftoken"
	"github.com/Factom-Asset-Tokens/fatd/internal/engine"
	"github.com/Factom-Asset-Tokens/fatd/internal/flag"
	"github.com/Factom-Asset-Tokens/fatd/internal/state"
)

var c = flag.FactomClient

var jsonrpc2Methods = jsonrpc2.MethodMap{
	"get-issuance":           getIssuance(false),
	"get-issuance-entry":     getIssuance(true),
	"get-transaction":        getTransaction(false),
	"get-transaction-entry":  getTransaction(true),
	"get-transactions":       getTransactions(false),
	"get-transactions-entry": getTransactions(true),
	"get-balance":            getBalance,
	"get-balances":           getBalances,
	"get-nf-balance":         getNFBalance,
	"get-stats":              getStats,
	"get-nf-token":           getNFToken,
	"get-nf-tokens":          getNFTokens,

	"send-transaction": sendTransaction,

	"get-daemon-tokens":     getDaemonTokens,
	"get-daemon-properties": getDaemonProperties,
	"get-sync-status":       getSyncStatus,
}

func getIssuance(entry bool) jsonrpc2.MethodFunc {
	return func(ctx context.Context, data json.RawMessage) interface{} {
		var params api.ParamsToken
		chain, put, err := validate(ctx, data, &params)
		if err != nil {
			return err
		}
		defer put()

		if entry {
			return chain.Issuance.Entry
		}
		return api.ResultGetIssuance{
			ParamsToken: api.ParamsToken{
				ChainID:       chain.ID,
				TokenID:       chain.TokenID,
				IssuerChainID: chain.Identity.ChainID,
			},
			Hash:      chain.Issuance.Entry.Hash,
			Timestamp: chain.Issuance.Entry.Timestamp.Unix(),
			Issuance:  chain.Issuance,
		}
	}
}

func getTransaction(getEntry bool) jsonrpc2.MethodFunc {
	return func(ctx context.Context, data json.RawMessage) interface{} {
		var params api.ParamsGetTransaction
		chain, put, err := validate(ctx, data, &params)
		if err != nil {
			return err
		}
		defer put()

		entry, err := entry.SelectValidByHash(chain.Conn, params.Hash)
		if err != nil {
			panic(err)
		}
		if !entry.IsPopulated() {
			return api.ErrorTransactionNotFound
		}

		if getEntry {
			return entry
		}

		result := api.ResultGetTransaction{
			Hash:      entry.Hash,
			Timestamp: entry.Timestamp.Unix(),
			Pending:   chain.LatestEntryTimestamp().Before(entry.Timestamp),
		}

		var tx interface{}
		switch chain.Issuance.Type {
		case fat0.Type:
			tx, err = fat0.NewTransaction(entry,
				(*factom.Bytes32)(chain.Identity.ID1Key))
		case fat1.Type:
			tx, err = fat1.NewTransaction(entry,
				(*factom.Bytes32)(chain.Identity.ID1Key))
		default:
			panic(fmt.Sprintf("unknown FAT type: %v", chain.Issuance.Type))
		}
		if err != nil {
			panic(err)
		}
		result.Tx = tx
		return result
	}
}

func getTransactions(getEntry bool) jsonrpc2.MethodFunc {
	return func(ctx context.Context, data json.RawMessage) interface{} {
		var params api.ParamsGetTransactions
		chain, put, err := validate(ctx, data, &params)
		if err != nil {
			return err
		}
		defer put()

		if params.NFTokenID != nil && chain.Issuance.Type != fat1.Type {
			err := api.ErrorTokenNotFound
			err.Data = "Token Chain is not FAT-1"
			return err
		}

		// Lookup Txs
		var nfTkns fat1.NFTokens
		if params.NFTokenID != nil {
			nfTkns = fat1.NFTokens{*params.NFTokenID: struct{}{}}
		}
		entry, err := entry.SelectByAddress(chain.Conn, params.StartHash,
			params.Addresses, nfTkns,
			params.ToFrom, params.Order,
			*params.Page, params.Limit)
		if err != nil {
			panic(err)
		}
		if len(entry) == 0 {
			return api.ErrorTransactionNotFound
		}
		if getEntry {
			// Omit the ChainID from the response since the client
			// already knows it.
			for i := range entry {
				entry[i].ChainID = nil
			}
			return entry
		}

		txs := make([]api.ResultGetTransaction, len(entry))
		for i := range txs {
			entry := entry[i]
			var tx interface{}
			switch chain.Issuance.Type {
			case fat0.Type:
				tx, err = fat0.NewTransaction(entry,
					(*factom.Bytes32)(chain.Identity.ID1Key))
			case fat1.Type:
				tx, err = fat1.NewTransaction(entry,
					(*factom.Bytes32)(chain.Identity.ID1Key))
			default:
				panic(fmt.Sprintf("unknown FAT type: %v",
					chain.Issuance.Type))
			}
			if err != nil {
				panic(err)
			}
			txs[i].Hash = entry.Hash
			txs[i].Timestamp = entry.Timestamp.Unix()
			txs[i].Pending = chain.LatestEntryTimestamp().
				Before(entry.Timestamp)
			txs[i].Tx = tx
		}
		return txs

	}
}

func getBalance(ctx context.Context, data json.RawMessage) interface{} {
	var params api.ParamsGetBalance
	chain, put, err := validate(ctx, data, &params)
	if err != nil {
		return err
	}
	defer put()

	_, balance, err := address.SelectIDBalance(chain.Conn, params.Address)
	if err != nil {
		panic(err)
	}
	return balance
}

func getBalances(ctx context.Context, data json.RawMessage) interface{} {
	var params api.ParamsGetBalances
	if _, _, err := validate(ctx, data, &params); err != nil {
		return err
	}

	issuedIDs := engine.IssuedIDs()
	balances := make(api.ResultGetBalances, len(issuedIDs))
	for _, chainID := range issuedIDs {
		chain, put, err := engine.Get(ctx, chainID, params.GetIncludePending())
		if err != nil {
			// ctx is done
			return err
		}
		defer put()
		fatChain, ok := state.ToFATChain(chain)
		if !ok {
			panic("not a FAT chain")
		}
		_, balance, err := address.SelectIDBalance(fatChain.Conn, params.Address)
		if err != nil {
			panic(err)
		}
		if balance > 0 {
			balances[*chainID] = balance
		}
	}
	return balances
}

func getNFBalance(ctx context.Context, data json.RawMessage) interface{} {
	var params api.ParamsGetNFBalance
	chain, put, err := validate(ctx, data, &params)
	if err != nil {
		return err
	}
	defer put()

	if chain.Issuance.Type != fat1.Type {
		err := api.ErrorTokenNotFound
		err.Data = "Token Chain is not FAT-1"
		return err
	}

	tkns, err := nftoken.SelectByOwner(chain.Conn, params.Address,
		*params.Page, params.Limit, params.Order)
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

var coinbaseRCDHash = fat.Coinbase()

func getStats(ctx context.Context, data json.RawMessage) interface{} {
	var params api.ParamsToken
	chain, put, err := validate(ctx, data, &params)
	if err != nil {
		return err
	}
	defer put()

	_, burned, err := address.SelectIDBalance(chain.Conn, &coinbaseRCDHash)
	if err != nil {
		panic(err)
	}
	txCount, err := entry.SelectCount(chain.Conn, true)
	e, err := entry.SelectLatestValid(chain.Conn)
	if err != nil {
		panic(err)
	}

	nonZeroBalances, err := address.SelectCount(chain.Conn, true)
	if err != nil {
		panic(err)
	}

	res := api.ResultGetStats{
		CirculatingSupply:        chain.NumIssued - burned,
		Burned:                   burned,
		Transactions:             txCount,
		IssuanceTimestamp:        chain.Issuance.Entry.Timestamp.Unix(),
		LastTransactionTimestamp: e.Timestamp.Unix(),
		NonZeroBalances:          nonZeroBalances,
	}
	if chain.IsIssued() {
		res.Issuance = &chain.Issuance
	}
	res.ChainID = chain.ID
	res.TokenID = chain.TokenID
	res.IssuerChainID = chain.Identity.ChainID
	res.IssuanceHash = chain.Issuance.Entry.Hash
	return res
}

func getNFToken(ctx context.Context, data json.RawMessage) interface{} {
	var params api.ParamsGetNFToken
	chain, put, err := validate(ctx, data, &params)
	if err != nil {
		return err
	}
	defer put()

	if chain.Issuance.Type != fat1.Type {
		err := api.ErrorTokenNotFound
		err.Data = "Token Chain is not FAT-1"
		return err
	}

	owner, creationHash, metadata, err := nftoken.SelectData(
		chain.Conn, *params.NFTokenID)
	if err != nil {
		panic(err)
	}
	if creationHash.IsZero() {
		err := api.ErrorTokenNotFound
		err.Data = "No such NFTokenID has been issued"
		return err
	}

	res := api.ResultGetNFToken{
		NFTokenID:  *params.NFTokenID,
		Metadata:   metadata,
		Owner:      &owner,
		CreationTx: &creationHash,
	}

	if owner == fat.Coinbase() {
		res.Owner = nil
		res.Burned = true
	}
	return res
}

func getNFTokens(ctx context.Context, data json.RawMessage) interface{} {
	var params api.ParamsGetAllNFTokens
	chain, put, err := validate(ctx, data, &params)
	if err != nil {
		return err
	}
	defer put()

	if chain.Issuance.Type != fat1.Type {
		err := api.ErrorTokenNotFound
		err.Data = "Token Chain is not FAT-1"
		return err
	}

	tkns, owners, creationHashes, metadata, err := nftoken.SelectDataAll(
		chain.Conn, params.Order, *params.Page, params.Limit)
	if err != nil {
		panic(err)
	}

	res := make([]api.ResultGetNFToken, len(tkns))
	for i := range res {
		res[i].NFTokenID = tkns[i]
		res[i].Metadata = metadata[i]
		res[i].CreationTx = &creationHashes[i]
		res[i].Owner = &owners[i]
		if owners[i] == fat.Coinbase() {
			res[i].Owner = nil
			res[i].Burned = true
		}
	}

	return res
}

func sendTransaction(ctx context.Context, data json.RawMessage) interface{} {
	var params api.ParamsSendTransaction
	chain, put, err := validate(ctx, data, &params)
	if err != nil {
		return err
	}
	defer put()
	if !params.DryRun && factom.Bytes32(flag.EsAdr).IsZero() {
		return api.ErrorNoEC
	}

	entry := params.Entry()
	var txErr error
	switch chain.Issuance.Type {
	case fat0.Type:
		txErr, err = attemptApplyFAT0Tx(chain, entry)
	case fat1.Type:
		txErr, err = attemptApplyFAT1Tx(chain, entry)
	}
	if err != nil {
		panic(err)
	}
	if txErr != nil {
		err := api.ErrorInvalidTransaction
		err.Data = txErr.Error()
		return err
	}

	var txID *factom.Bytes32
	if !params.DryRun {
		balance, err := flag.ECAdr.GetBalance(ctx, c)
		if err != nil {
			panic(err)
		}
		cost, _ := entry.Cost()
		if balance < uint64(cost) {
			return api.ErrorNoEC
		}
		txID = new(factom.Bytes32)
		var commit []byte
		commit, *txID = factom.GenerateCommit(
			flag.EsAdr, params.Raw, entry.Hash, false)
		if err := c.Commit(ctx, commit); err != nil {
			panic(err)
		}
		if err := c.Reveal(ctx, params.Raw); err != nil {
			panic(err)
		}

	}

	return struct {
		ChainID *factom.Bytes32 `json:"chainid"`
		TxID    *factom.Bytes32 `json:"txid,omitempty"`
		Hash    *factom.Bytes32 `json:"entryhash"`
	}{ChainID: chain.ID, TxID: txID, Hash: entry.Hash}
}
func attemptApplyFAT0Tx(chain *state.FATChain, e factom.Entry) (txErr, err error) {
	// Validate tx
	valid, err := entry.CheckUniquelyValid(chain.Conn, 0, e.Hash)
	if err != nil {
		return
	}
	if !valid {
		txErr = fmt.Errorf("replay: hash previously marked valid")
		return
	}

	tx, txErr := fat0.NewTransaction(e, (*factom.Bytes32)(chain.Identity.ID1Key))
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
	} else {
		// Check all input balances
		for adr, amount := range tx.Inputs {
			var bal uint64
			if _, bal, err = address.SelectIDBalance(
				chain.Conn, &adr); err != nil {
				return
			}
			if amount > bal {
				txErr = fmt.Errorf("insufficient balance: %v", adr)
				return
			}
		}
	}
	return
}
func attemptApplyFAT1Tx(chain *state.FATChain, e factom.Entry) (txErr, err error) {
	// Validate tx
	valid, err := entry.CheckUniquelyValid(chain.Conn, 0, e.Hash)
	if err != nil {
		return
	}
	if !valid {
		txErr = fmt.Errorf("replay: hash previously marked valid")
		return
	}

	tx, txErr := fat1.NewTransaction(e, (*factom.Bytes32)(chain.Identity.ID1Key))
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
		for nfID := range nfTkns {
			var ownerID int64
			ownerID, err = nftoken.SelectOwnerID(chain.Conn, nfID)
			if err != nil {
				return
			}
			if ownerID != -1 {
				txErr = fmt.Errorf("NFTokenID{%v} already exists", nfID)
				return
			}
		}
	} else {
		for adr, nfTkns := range tx.Inputs {
			var adrID int64
			var bal uint64
			adrID, bal, err = address.SelectIDBalance(chain.Conn, &adr)
			if err != nil {
				return
			}
			amount := uint64(len(nfTkns))
			if amount > bal {
				txErr = fmt.Errorf("insufficient balance: %v", adr)
				return
			}
			for nfTkn := range nfTkns {
				var ownerID int64
				ownerID, err = nftoken.SelectOwnerID(
					chain.Conn, nfTkn)
				if err != nil {
					return
				}
				if ownerID == -1 {
					txErr = fmt.Errorf("no such NFToken{%v}", nfTkn)
					return
				}
				if ownerID != adrID {
					txErr = fmt.Errorf("NFToken{%v} not owned by %v",
						nfTkn, adr)
					return
				}
			}
		}
	}
	return
}
func getDaemonTokens(ctx context.Context, data json.RawMessage) interface{} {
	if _, _, err := validate(ctx, data, nil); err != nil {
		return err
	}

	issuedIDs := engine.IssuedIDs()
	chains := make([]api.ParamsToken, len(issuedIDs))
	for i, chainID := range issuedIDs {
		// Use pending = true because a chain that has a pending
		// issuance entry will not show up in this list, and no other
		// pending entry will effect the data of interest. Using the
		// pending state is more efficient.
		chain, put, err := engine.Get(ctx, chainID, true)
		if err != nil {
			// ctx is done
			return err
		}
		defer put()
		fatChain, ok := state.ToFATChain(chain)
		if !ok {
			panic("not a FAT chain")
		}
		chainID := chainID
		chains[i].ChainID = chainID
		chains[i].TokenID = fatChain.TokenID
		chains[i].IssuerChainID = fatChain.Identity.ChainID
	}
	return chains
}

func getDaemonProperties(ctx context.Context, data json.RawMessage) interface{} {
	if _, _, err := validate(ctx, data, nil); err != nil {
		return err
	}
	return api.ResultGetDaemonProperties{
		FatdVersion: flag.Revision,
		APIVersion:  APIVersion,
		NetworkID:   flag.NetworkID,
	}
}

func getSyncStatus(ctx context.Context, data json.RawMessage) interface{} {
	sync, current := engine.GetSyncStatus()
	return api.ResultGetSyncStatus{Sync: sync, Current: current}
}

func validate(ctx context.Context,
	data json.RawMessage, params api.Params) (*state.FATChain, func(), error) {
	if params == nil {
		if len(data) > 0 {
			return nil, nil, jsonrpc2.ErrorInvalidParams(
				`no "params" accepted`)
		}
		return nil, nil, nil
	}
	if len(data) == 0 {
		return nil, nil, params.IsValid()
	}
	if err := unmarshalStrict(data, params); err != nil {
		return nil, nil, jsonrpc2.ErrorInvalidParams(err)
	}
	if err := params.IsValid(); err != nil {
		return nil, nil, err
	}
	if params.GetIncludePending() && flag.DisablePending {
		return nil, nil, api.ErrorPendingDisabled
	}
	chainID := params.ValidChainID()
	if chainID != nil {
		ctx, cancel := context.WithTimeout(ctx, flag.APITimeout)
		chain, put, err := engine.Get(ctx, chainID, params.GetIncludePending())
		if err != nil {
			// ctx is done
			cancel()
			log.Debugf("err")
			return nil, nil, err
		}
		if put == nil {
			cancel()
			return nil, nil, api.ErrorTokenNotFound
		}
		fatChain, ok := state.ToFATChain(chain)
		if !ok {
			panic("not a FAT chain")
		}
		if !fatChain.IsIssued() {
			cancel()
			put()
			return nil, nil, api.ErrorTokenNotFound
		}
		return fatChain, func() { cancel(); put() }, nil
	}
	return nil, nil, nil
}

func unmarshalStrict(data []byte, v interface{}) error {
	b := bytes.NewBuffer(data)
	d := json.NewDecoder(b)
	d.DisallowUnknownFields()
	return d.Decode(v)
}
