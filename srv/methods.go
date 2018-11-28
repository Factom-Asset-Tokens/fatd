package srv

import (
	"bytes"
	"encoding/json"

	jrpc "github.com/AdamSLevy/jsonrpc2/v7"
	"github.com/Factom-Asset-Tokens/fatd/factom"
)

var jrpcMethods = jrpc.MethodMap{
	"get-issuance":     getIssuance,
	"get-transaction":  getTransaction,
	"get-transactions": getTransactions,
	"get-balance":      getBalance,
	//"get-stats":        getStats,
	//"get-nf-token":     getNFToken,

	//"send-transaction": sendTransaction,

	//"get-daemon-tokens":     getDaemonTokens,
	//"get-daemon-properties": getDaemonProperties,
}

var (
	TokenParamsRes = jrpc.NewInvalidParamsErrorResponse(
		`"params" required: either "chain-id" or both "token-id" and "issuer-id"`)
	TokenNotFoundRes = jrpc.NewErrorResponse(-32800, "Token Not Found",
		"not yet issued or not tracked by this instance of fatd")
	TransactionNotFoundRes = jrpc.NewErrorResponse(-32800, "Token Not Found",
		"not yet issued or not tracked by this instance of fatd")
	GetTransactionParamsRes = jrpc.NewInvalidParamsErrorResponse(
		`"params" required: "hash" and either "chain-id" or both "token-id" and "issuer-id"`)
	GetTransactionsParamsRes = jrpc.NewInvalidParamsErrorResponse(
		`"params" required: "hash" or "start" and either "chain-id" or both "token-id" and "issuer-id", "limit" must be greater than 0 if provided`)
	GetBalanceParamsRes = jrpc.NewInvalidParamsErrorResponse(
		`"params" required: "fa-address" and either "chain-id" or both "token-id" and "issuer-id"`)
)

func getIssuance(data json.RawMessage) jrpc.Response {
	params := TokenParams{}
	chainID, res := validate(data, &params, TokenParamsRes)
	if chainID == nil {
		return res
	}

	// Look up issuance

	return TokenNotFoundRes
}

func getTransaction(data json.RawMessage) jrpc.Response {
	params := GetTransactionParams{}
	chainID, res := validate(data, &params, GetTransactionParamsRes)
	if chainID == nil {
		return res
	}

	// Lookup Tx by Hash

	return TransactionNotFoundRes
}

func getTransactions(data json.RawMessage) jrpc.Response {
	params := GetTransactionsParams{}
	chainID, res := validate(data, &params, GetTransactionsParamsRes)
	if chainID == nil {
		return res
	}

	// Lookup Txs

	return TransactionNotFoundRes
}

func getBalance(data json.RawMessage) jrpc.Response {
	params := GetBalanceParams{}
	chainID, res := validate(data, &params, GetBalanceParamsRes)
	if chainID == nil {
		return res
	}

	// Lookup Txs

	return jrpc.NewResponse(0)
}

func validate(data json.RawMessage, params Params,
	invalidParamsErrorRes jrpc.Response) (*factom.Bytes32, jrpc.Response) {
	if data == nil {
		return nil, invalidParamsErrorRes
	}
	if err := unmarshalStrict(data, params); err != nil {
		return nil, jrpc.NewInvalidParamsErrorResponse(err.Error())
	}
	chainID := params.ValidChainID()
	if chainID == nil || !params.IsValid() {
		return nil, invalidParamsErrorRes
	}
	return chainID, jrpc.Response{}
}

func unmarshalStrict(data []byte, v interface{}) error {
	b := bytes.NewBuffer(data)
	d := json.NewDecoder(b)
	d.DisallowUnknownFields()
	return d.Decode(v)
}
