package srv

import (
	"bytes"
	"encoding/json"
	"time"

	jrpc "github.com/AdamSLevy/jsonrpc2/v6"
	"github.com/Factom-Asset-Tokens/fatd/fat0"
	"github.com/Factom-Asset-Tokens/fatd/state"
)

var jrpcHandler = jrpc.HTTPRequestHandler(jrpc.MethodMap{
	"version":          version,
	"get-issuance":     getIssuance,
	"get-transaction":  getTransaction,
	"get-transactions": getTransactions,
	"get-balance":      getBalance,
	"get-stats":        getStats,
})

func version(params json.RawMessage) jrpc.Response {
	if params != nil {
		return jrpc.NewInvalidParamsErrorResponse("Unexpected parameters")
	}
	return jrpc.NewResponse("0.0.0")
}

var requiredParamsErr = `required params: "chain-id", or "token-id" and "issuer-id"`

func getIssuance(params json.RawMessage) jrpc.Response {
	if params == nil {
		return jrpc.NewInvalidParamsErrorResponse(requiredParamsErr)
	}
	token := TokenParams{}
	if err := unmarshalStrict(params, &token); err != nil {
		return jrpc.NewInvalidParamsErrorResponse(err.Error())
	}
	chainID := token.ChainID
	if token.ChainID == nil {
		if token.TokenID == nil || token.IssuerChainID == nil {
			return jrpc.NewInvalidParamsErrorResponse(requiredParamsErr)
		}
		chainID = fat0.ChainID(*token.TokenID, token.IssuerChainID)
	}

	issuance := state.GetIssuance(chainID)
	if issuance == nil {
		return jrpc.NewErrorResponse(-32800, "Token Not Found",
			"not yet issued or not tracked by this instance of fatd")
	}

	return jrpc.NewResponse(issuance)
}

func getTransaction(params json.RawMessage) jrpc.Response {
	if params == nil {
		return jrpc.NewInvalidParamsErrorResponse("Parameters are required for this method")
	}

	token := TokenParams{}
	if err := unmarshalStrict(params, &token); err != nil {
		return jrpc.NewInvalidParamsErrorResponse("unrecognized params")

	}
	if token.ChainID == nil &&
		(token.TokenID == nil || token.IssuerChainID == nil) {
		return jrpc.NewInvalidParamsErrorResponse(requiredParamsErr)
	}

	if token.TransactionID == nil {
		return jrpc.NewInvalidParamsErrorResponse("A transaction ID ('tx-id') is required for method 'get-transaction'")
	}

	input := map[string]interface{}{
		"address": "FA1zT4aFpEvcnPqPCigB3fvGu4Q4mTXY22iiuV69DqE1pNhdF2MC",
		"amount":  302,
	}

	transaction := map[string]interface{}{
		"inputs":      [1]map[string]interface{}{input},
		"outputs":     [1]map[string]interface{}{input},
		"blockheight": 153745,
		"salt":        "80d87a8bd5cf2a3eca9037c2229f3701eed29360caa975531ef5fe476b1b70b5",
		"timestamp":   time.Now().Unix() - 3600, //unix timestamp Added by this lib, based on Factom entry timestamp
		"extIds":      [2]string{"874220a808090fb736f345dd5d67ac26eab94c9c9f51b708b05cdc4d42f65aae", "nPqPCigB3fvGu4Q4mTXY22iiuV69DqE1pNhd"},
	}

	return jrpc.NewResponse(transaction)
}

var getTransactions jrpc.MethodFunc = func(params json.RawMessage) jrpc.Response {
	if params == nil {
		return jrpc.NewInvalidParamsErrorResponse("Parameters are required for this method")
	}

	token := TokenParams{}
	if err := unmarshalStrict(params, &token); err != nil {
		return jrpc.NewInvalidParamsErrorResponse("unrecognized params")

	}
	if token.ChainID == nil &&
		(token.TokenID == nil || token.IssuerChainID == nil) {
		return jrpc.NewInvalidParamsErrorResponse(requiredParamsErr)
	}

	input := map[string]interface{}{
		"address": "FA1zT4aFpEvcnPqPCigB3fvGu4Q4mTXY22iiuV69DqE1pNhdF2MC",
		"amount":  302,
	}

	transaction := map[string]interface{}{
		"inputs":      [1]map[string]interface{}{input},
		"outputs":     [1]map[string]interface{}{input},
		"blockheight": 153745,
		"salt":        "80d87a8bd5cf2a3eca9037c2229f3701eed29360caa975531ef5fe476b1b70b5",
		"timestamp":   time.Now().Unix() - 3600, //unix timestamp Added by this lib, based on Factom entry timestamp
		"extIds":      [2]string{"874220a808090fb736f345dd5d67ac26eab94c9c9f51b708b05cdc4d42f65aae", "nPqPCigB3fvGu4Q4mTXY22iiuV69DqE1pNhd"},
	}

	return jrpc.NewResponse([2]map[string]interface{}{transaction, transaction})
}

var getBalance jrpc.MethodFunc = func(params json.RawMessage) jrpc.Response {
	if params == nil {
		return jrpc.NewInvalidParamsErrorResponse("Parameters are required for this method")
	}

	token := TokenParams{}
	if err := unmarshalStrict(params, &token); err != nil {
		return jrpc.NewInvalidParamsErrorResponse("unrecognized params")

	}
	if token.ChainID == nil &&
		(token.TokenID == nil || token.IssuerChainID == nil) {
		return jrpc.NewInvalidParamsErrorResponse(requiredParamsErr)
	}

	return jrpc.NewResponse(302)
}

func getStats(params json.RawMessage) jrpc.Response {
	if params == nil {
		return jrpc.NewInvalidParamsErrorResponse("Parameters are required for this method")
	}

	token := TokenParams{}
	if err := unmarshalStrict(params, &token); err != nil {
		return jrpc.NewInvalidParamsErrorResponse("unrecognized params")

	}
	if token.ChainID == nil &&
		(token.TokenID == nil || token.IssuerChainID == nil) {
		return jrpc.NewInvalidParamsErrorResponse(requiredParamsErr)
	}

	stats := map[string]interface{}{
		"supply":                   10000000,
		"circulatingSupply":        53024,
		"transactions":             7745,
		"issuanceTimestamp":        1518286500,
		"lastTransactionTimestamp": time.Now().Unix(),
	}

	return jrpc.NewResponse(stats)
}

func unmarshalStrict(data []byte, v interface{}) error {
	b := bytes.NewBuffer(data)
	d := json.NewDecoder(b)
	d.DisallowUnknownFields()
	return d.Decode(v)
}
