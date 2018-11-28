package srv

import (
	"bytes"
	"encoding/json"

	jrpc "github.com/AdamSLevy/jsonrpc2/v7"
	"github.com/Factom-Asset-Tokens/fatd/state"
)

var jrpcMethods = jrpc.MethodMap{
	"get-issuance": getIssuance,
	//"get-transaction":  getTransaction,
	//"get-transactions": getTransactions,
	//"get-balance":      getBalance,
	//"get-stats":        getStats,
	//"get-nf-token":     getNFToken,

	//"send-transaction": sendTransaction,

	//"get-daemon-tokens":     getDaemonTokens,
	//"get-daemon-properties": getDaemonProperties,
}

var tokenParamsErr = `"params" required: either "chain-id" or both "token-id" and "issuer-id"`

var (
	TokenNotFoundError = jrpc.NewErrorResponse(-32800, "Token Not Found",
		"not yet issued or not tracked by this instance of fatd")
)

func getIssuance(params json.RawMessage) jrpc.Response {
	if params == nil {
		return jrpc.NewInvalidParamsErrorResponse(tokenParamsErr)
	}
	token := TokenParams{}
	if err := unmarshalStrict(params, &token); err != nil {
		return jrpc.NewInvalidParamsErrorResponse(err.Error())
	}
	chainID := token.ValidChainID()
	if chainID == nil {
		return jrpc.NewInvalidParamsErrorResponse(tokenParamsErr)
	}

	issuance := state.GetIssuance(chainID)
	if issuance == nil {
		return TokenNotFoundError
	}

	return jrpc.NewResponse(issuance)
}

func unmarshalStrict(data []byte, v interface{}) error {
	b := bytes.NewBuffer(data)
	d := json.NewDecoder(b)
	d.DisallowUnknownFields()
	return d.Decode(v)
}
