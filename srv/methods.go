package srv

import (
	"bytes"
	"encoding/json"
	"fmt"

	jrpc "github.com/AdamSLevy/jsonrpc2/v10"
	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat0"
	"github.com/Factom-Asset-Tokens/fatd/flag"
	"github.com/Factom-Asset-Tokens/fatd/state"
)

var jrpcMethods = jrpc.MethodMap{
	"get-issuance":           getIssuance(false),
	"get-issuance-entry":     getIssuance(true),
	"get-transaction":        getTransaction(false),
	"get-transaction-entry":  getTransaction(true),
	"get-transactions":       getTransactions(false),
	"get-transactions-entry": getTransactions(true),
	"get-balance":            getBalance,
	"get-stats":              getStats,
	"get-nf-token":           getNFToken,

	"send-transaction": sendTransaction,

	"get-daemon-tokens":     getDaemonTokens,
	"get-daemon-properties": getDaemonProperties,
}

type ResultsGetIssuance struct {
	ParamsToken
	Timestamp *factom.Time  `json:"timestamp"`
	Issuance  fat0.Issuance `json:"issuance"`
}

func getIssuance(entry bool) jrpc.MethodFunc {
	return func(data json.RawMessage) interface{} {
		params := ParamsToken{}
		chainID, res := validate(data, &params)
		if chainID == nil {
			return res
		}

		// Look up issuance
		chain := state.Chains.Get(chainID)
		if !chain.IsIssued() {
			return ErrorTokenNotFound
		}
		if entry {
			return chain.Issuance.Entry.Entry
		}
		return ResultsGetIssuance{
			ParamsToken: ParamsToken{
				ChainID:       chainID,
				TokenID:       chain.Token,
				IssuerChainID: chain.Identity.ChainID,
			},
			Timestamp: chain.Issuance.Timestamp,
			Issuance:  chain.Issuance,
		}
	}
}

func getTransaction(entry bool) jrpc.MethodFunc {
	return func(data json.RawMessage) interface{} {
		params := ParamsGetTransaction{}
		chainID, res := validate(data, &params)
		if chainID == nil {
			return res
		}

		// Lookup Tx by Hash
		chain := state.Chains.Get(chainID)
		transaction, err := chain.GetTransaction(params.Hash)
		if err != nil {
			panic(err)
		}
		if !transaction.IsPopulated() {
			return ErrorTransactionNotFound
		}

		if entry {
			return transaction.Entry.Entry
		}
		if err := transaction.UnmarshalEntry(); err != nil {
			panic(err)
		}
		return ResultsGetTransaction{
			Hash:      transaction.Hash,
			Timestamp: transaction.Timestamp,
			Tx:        transaction,
		}
	}
}

type ResultsGetTransaction struct {
	Hash      *factom.Bytes32  `json:"entryhash"`
	Timestamp *factom.Time     `json:"timestamp"`
	Tx        fat0.Transaction `json:"data"`
}

func getTransactions(entry bool) jrpc.MethodFunc {
	return func(data json.RawMessage) interface{} {
		params := ParamsGetTransactions{}
		chainID, res := validate(data, &params)
		if chainID == nil {
			return res
		}

		// Lookup Txs
		chain := state.Chains.Get(chainID)
		transactions, err := chain.GetTransactions(params.Hash,
			params.FactoidAddress, params.ToFrom,
			*params.Start, *params.Limit)
		if err != nil {
			log.Debug(err)
			panic(err)
		}
		if len(transactions) == 0 {
			return ErrorTransactionNotFound
		}
		if entry {
			txs := make([]factom.Entry, len(transactions))
			for i := range txs {
				txs[i] = transactions[i].Entry.Entry
				txs[i].ChainID = nil
			}
			return txs
		}

		txs := make([]ResultsGetTransaction, len(transactions))
		for i := range txs {
			txs[i].Hash = transactions[i].Hash
			txs[i].Timestamp = transactions[i].Timestamp
			txs[i].Tx = transactions[i]
		}

		return txs
	}
}

func getBalance(data json.RawMessage) interface{} {
	params := ParamsGetBalance{}
	chainID, res := validate(data, &params)
	if chainID == nil {
		return res
	}

	// Lookup Txs
	chain := state.Chains.Get(chainID)
	if !chain.IsIssued() {
		return ErrorTokenNotFound
	}
	balance, err := chain.GetBalance(*params.Address)
	if err != nil {
		panic(err)
	}
	return balance
}

type ResultsGetStats struct {
	Supply                   int64        `json:"supply"`
	CirculatingSupply        uint64       `json:"circulating"`
	Burned                   uint64       `json:"burned"`
	Transactions             int          `json:"transactions"`
	IssuanceTimestamp        *factom.Time `json:"issuancets"`
	LastTransactionTimestamp *factom.Time `json:"lasttxts,omitempty"`
}

func getStats(data json.RawMessage) interface{} {
	params := ParamsToken{}
	chainID, res := validate(data, &params)
	if chainID == nil {
		return res
	}

	chain := state.Chains.Get(chainID)
	if !chain.IsIssued() {
		return ErrorTokenNotFound
	}

	coinbase := factom.Address{}
	burned, err := chain.GetBalance(coinbase)
	if err != nil {
		panic(err)
	}
	txs, err := chain.GetTransactions(nil, nil, "", 0, 0)
	if err != nil {
		panic(err)
	}

	var lastTxTs *factom.Time
	if len(txs) > 0 {
		lastTxTs = txs[len(txs)-1].Timestamp
	}
	return ResultsGetStats{
		Supply:                   chain.Supply,
		CirculatingSupply:        chain.Issued - burned,
		Burned:                   burned,
		Transactions:             len(txs),
		IssuanceTimestamp:        chain.Issuance.Timestamp,
		LastTransactionTimestamp: lastTxTs,
	}
}

func getNFToken(data json.RawMessage) interface{} {
	params := ParamsGetNFToken{}
	chainID, err := validate(data, &params)
	if chainID == nil {
		return err
	}

	return ErrorTokenNotFound
}

func sendTransaction(data json.RawMessage) interface{} {
	if len(flag.ECPub) == 0 {
		return ErrorNoEC
	}
	params := ParamsSendTransaction{}
	chainID, rpcErr := validate(data, &params)
	if chainID == nil {
		return rpcErr
	}

	chain := state.Chains.Get(chainID)
	if !chain.IsIssued() {
		rpcErr = ErrorTokenNotFound
		rpcErr.Data = chainID
		return rpcErr
	}

	tx := fat0.NewTransaction(params.Entry())
	if err := tx.Valid(chain.IDKey); err != nil {
		rpcErr = ErrorInvalidTransaction
		rpcErr.Data = err.Error()
		return rpcErr
	}

	// check balances
	if tx.IsCoinbase() {
		if tx.Inputs.Sum() > uint64(chain.Supply)-chain.Issued {
			rpcErr := ErrorInvalidTransaction
			rpcErr.Data = "insufficient coinbase supply"
			return rpcErr
		}
	} else {
		for rcdHash, amount := range tx.Inputs {
			adr := factom.NewAddress(&rcdHash)
			balance, err := chain.GetBalance(adr)
			if err != nil {
				log.Error(err)
				panic(err)
			}
			if amount > balance {
				rpcErr := ErrorInvalidTransaction
				rpcErr.Data = fmt.Sprintf(
					"insufficient balance: %v", adr)
				return rpcErr
			}
		}
	}

	txID, err := tx.Create(flag.ECPub)
	if err != nil {
		log.Error(err)
		panic(err)
	}

	return struct {
		ChainID *factom.Bytes32 `json:"chainid"`
		TxID    *factom.Bytes32 `json:"txid"`
		Hash    *factom.Bytes32 `json:"entryhash"`
	}{ChainID: chainID, TxID: txID, Hash: tx.Hash}
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

func getDaemonProperties(data json.RawMessage) interface{} {
	if data != nil {
		return ParamsErrorNoParams
	}
	return struct {
		FatdVersion string `json:"fatdversion"`
		APIVersion  string `json:"apiversion"`
	}{FatdVersion: "0.0.0", APIVersion: "v0"}
}

func validate(data json.RawMessage, params Params) (*factom.Bytes32, jrpc.Error) {
	if data == nil {
		return nil, params.Error()
	}
	if err := unmarshalStrict(data, params); err != nil {
		return nil, jrpc.NewInvalidParamsError(err.Error())
	}
	chainID := params.ValidChainID()
	if chainID == nil || !params.IsValid() {
		return nil, params.Error()
	}
	return chainID, jrpc.Error{}
}

func unmarshalStrict(data []byte, v interface{}) error {
	b := bytes.NewBuffer(data)
	d := json.NewDecoder(b)
	d.DisallowUnknownFields()
	return d.Decode(v)
}
