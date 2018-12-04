package srv

import (
	"bytes"
	"encoding/json"

	jrpc "github.com/AdamSLevy/jsonrpc2/v9"
	"github.com/Factom-Asset-Tokens/fatd/factom"
)

var jrpcMethods = jrpc.MethodMap{
	"get-issuance":     getIssuance,
	"get-transaction":  getTransaction,
	"get-transactions": getTransactions,
	"get-balance":      getBalance,
	"get-stats":        getStats,
	"get-nf-token":     getNFToken,

	"send-transaction": sendTransaction,

	"get-daemon-tokens":     getDaemonTokens,
	"get-daemon-properties": getDaemonProperties,
}

func getIssuance(data json.RawMessage) interface{} {
	params := TokenParams{}
	chainID, res := validate(data, &params)
	if chainID == nil {
		return res
	}

	// Look up issuance

	return TokenNotFoundError
}

func getTransaction(data json.RawMessage) interface{} {
	params := GetTransactionParams{}
	chainID, res := validate(data, &params)
	if chainID == nil {
		return res
	}

	// Lookup Tx by Hash

	return TransactionNotFoundError
}

func getTransactions(data json.RawMessage) interface{} {
	params := GetTransactionsParams{}
	chainID, res := validate(data, &params)
	if chainID == nil {
		return res
	}

	// Lookup Txs

	return TransactionNotFoundError
}

func getBalance(data json.RawMessage) interface{} {
	params := GetBalanceParams{}
	chainID, res := validate(data, &params)
	if chainID == nil {
		return res
	}

	// Lookup Txs

	return 0
}

func getStats(data json.RawMessage) interface{} {
	params := TokenParams{}
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
	params := GetNFTokenParams{}
	chainID, err := validate(data, &params)
	if chainID == nil {
		return err
	}

	return TokenNotFoundError
}

func sendTransaction(data json.RawMessage) interface{} {
	params := SendTransactionParams{}
	chainID, err := validate(data, &params)
	if chainID == nil {
		return err
	}

	return TokenNotFoundError
}

func getDaemonTokens(data json.RawMessage) interface{} {
	if data != nil {
		return NoParamsError
	}

	return []struct {
		TokenID  string          `json:"token-id"`
		IssuerID *factom.Bytes32 `json:"issuer-id"`
		ChainID  *factom.Bytes32 `json:"chain-id"`
	}{{}}
}

func getDaemonProperties(data json.RawMessage) interface{} {
	if data != nil {
		return NoParamsError
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
