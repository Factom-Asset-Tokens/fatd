package srv

import (
	"encoding/json"

	jrpc "github.com/AdamSLevy/jsonrpc2/v4"
)

var version jrpc.MethodFunc = func(params json.RawMessage) *jrpc.Response {
	if params != nil {
		return jrpc.NewInvalidParamsErrorResponse("Unexpected parameters")
	}
	return jrpc.NewResponse("0.0.0")
}

var getIssuance jrpc.MethodFunc = func(params json.RawMessage) *jrpc.Response {
	if params == nil {
		return jrpc.NewInvalidParamsErrorResponse("Params are required for this method")
	}

	issuance := map[string]interface{}{
		"type":   "FAT-0",
		"name":   "Test Token",
		"symbol": "TTK",
		"supply": 10000000,
		"salt":   "874220a808090fb736f345dd5d67ac26eab94c9c9f51b708b05cdc4d42f65aae",
	}

	return jrpc.NewResponse(issuance)
}

var getTransaction jrpc.MethodFunc = func(params json.RawMessage) *jrpc.Response {
	if params == nil {
		return jrpc.NewInvalidParamsErrorResponse("Params are required for this method")
	}

	input := map[string]interface{}{
		"address": "FA1zT4aFpEvcnPqPCigB3fvGu4Q4mTXY22iiuV69DqE1pNhdF2MC",
		"amount":  302,
	}

	transaction := map[string]interface{}{
		"inputs":         [1]map[string]interface{}{input},
		"outputs":        [1]map[string]interface{}{input},
		"milliTimestamp": 1537450868,
		"salt":           "80d87a8bd5cf2a3eca9037c2229f3701eed29360caa975531ef5fe476b1b70b5",
	}

	return jrpc.NewResponse(transaction)
}

var getBalance jrpc.MethodFunc = func(params json.RawMessage) *jrpc.Response {
	if params == nil {
		return jrpc.NewInvalidParamsErrorResponse("Params are required for this method")
	}

	return jrpc.NewResponse(302)
}

var getStats jrpc.MethodFunc = func(params json.RawMessage) *jrpc.Response {
	if params != nil {
		return jrpc.NewInvalidParamsErrorResponse("Unexpected parameters")
	}

	stats := map[string]interface{}{
		"supply":                   10000000,
		"circulatingSupply":        53024,
		"transactions":             7745,
		"issuanceTimestamp":        1518286500,
		"lastTransactionTimestamp": 1518286899,
	}

	return jrpc.NewResponse(stats)
}
