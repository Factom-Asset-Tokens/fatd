package srv

import (
	"bytes"
	"encoding/json"

	jrpc "github.com/AdamSLevy/jsonrpc2/v9"
	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat0"
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
		return chain.Issuance
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
		return transaction
	}
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
		transactions, err := chain.GetTransactions(params.Hash, params.FactoidAddress,
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

		txs := make([]struct {
			Hash *factom.Bytes32  `json:"entryhash"`
			Time int64            `json:"timestamp"`
			Tx   fat0.Transaction `json:"data"`
		}, len(transactions))
		for i := range txs {
			txs[i].Hash = transactions[i].Hash
			txs[i].Time = transactions[i].Timestamp.Unix()
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

func getStats(data json.RawMessage) interface{} {
	params := ParamsToken{}
	chainID, res := validate(data, &params)
	if chainID == nil {
		return res
	}

	return struct {
		Supply                   int `json:"supply"`
		CirculatingSupply        int `json:"circulating-supply"`
		Transactions             int `json:"transactions"`
		IssuanceTimestamp        int `json:"issuance-timestamp"`
		LastTransactionTimestamp int `json:"last-transaction-timestamp"`
	}{}
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
	params := ParamsSendTransaction{}
	chainID, err := validate(data, &params)
	if chainID == nil {
		return err
	}

	return ErrorTokenNotFound
}

func getDaemonTokens(data json.RawMessage) interface{} {
	if data != nil {
		return ParamsErrorNoParams
	}

	return []struct {
		TokenID  string          `json:"token-id"`
		IssuerID *factom.Bytes32 `json:"issuer-id"`
		ChainID  *factom.Bytes32 `json:"chain-id"`
	}{{}}
}

func getDaemonProperties(data json.RawMessage) interface{} {
	if data != nil {
		return ParamsErrorNoParams
	}
	return struct {
		FatdVersion string `json:"fatd-version"`
		APIVersion  string `json:"api-version"`
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
