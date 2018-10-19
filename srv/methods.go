package srv

import (
	"bytes"
	"encoding/json"
	jrpc "github.com/AdamSLevy/jsonrpc2/v4"
	"time"
)

var version jrpc.MethodFunc = func(params json.RawMessage) *jrpc.Response {
	if params != nil {
		return jrpc.NewInvalidParamsErrorResponse("Unexpected parameters")
	}
	return jrpc.NewResponse("0.0.0")
}

var requiredParamsErr = `required params: "chain-id", or "token-id" and "issuer-id"`
var getIssuance jrpc.MethodFunc = func(params json.RawMessage) *jrpc.Response {
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

	issuance := map[string]interface{}{
		"type":      "FAT-0",
		"name":      "Test Token",
		"symbol":    "TTK",
		"supply":    10000000,
		"salt":      "874220a808090fb736f345dd5d67ac26eab94c9c9f51b708b05cdc4d42f65aae",
		"timestamp": time.Now().Unix() - 3600, //unix timestamp Added by this lib, based on Factom entry timestamp
		"extIds":    [1]string{"874220a808090fb736f345dd5d67ac26eab94c9c9f51b708b05cdc4d42f65aae"},
	}

	return jrpc.NewResponse(issuance)
}

var getTransaction jrpc.MethodFunc = func(params json.RawMessage) *jrpc.Response {
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

var getTransactions jrpc.MethodFunc = func(params json.RawMessage) *jrpc.Response {
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

var getBalance jrpc.MethodFunc = func(params json.RawMessage) *jrpc.Response {
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

var getStats jrpc.MethodFunc = func(params json.RawMessage) *jrpc.Response {
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
