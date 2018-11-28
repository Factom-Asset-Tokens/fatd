package srv

import (
	"bytes"
	"encoding/json"

	jrpc "github.com/AdamSLevy/jsonrpc2/v7"
)

var jrpcMethods = jrpc.MethodMap{
	"get-issuance":    getIssuance,
	"get-transaction": getTransaction,
	//"get-transactions": getTransactions,
	//"get-balance":      getBalance,
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
)

func getIssuance(params json.RawMessage) jrpc.Response {
	if params == nil {
		return TokenParamsRes
	}
	token := TokenParams{}
	if err := unmarshalStrict(params, &token); err != nil {
		return jrpc.NewInvalidParamsErrorResponse(err.Error())
	}
	chainID := token.ValidChainID()
	if chainID == nil {
		return TokenParamsRes
	}

	// Look up issuance

	return TokenNotFoundRes
}

func getTransaction(params json.RawMessage) jrpc.Response {
	if params == nil {
		return GetTransactionParamsRes
	}
	txParams := GetTransactionParams{}
	if err := unmarshalStrict(params, &txParams); err != nil {
		return jrpc.NewInvalidParamsErrorResponse(err.Error())
	}
	chainID := txParams.ValidChainID()
	if chainID == nil || txParams.Hash == nil {
		return GetTransactionParamsRes
	}

	// Lookup Tx by Hash

	return TransactionNotFoundRes
}

func unmarshalStrict(data []byte, v interface{}) error {
	b := bytes.NewBuffer(data)
	d := json.NewDecoder(b)
	d.DisallowUnknownFields()
	return d.Decode(v)
}
