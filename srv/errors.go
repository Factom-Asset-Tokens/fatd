package srv

import jrpc "github.com/AdamSLevy/jsonrpc2/v10"

var (
	ParamsErrorNoParams = jrpc.NewInvalidParamsError(`no "params" accepted`)
	ParamsErrorToken    = jrpc.NewInvalidParamsError(
		`required: either "chain-id" or both "token-id" and "issuer-id"`)
	ParamsErrorGetTransaction = jrpc.NewInvalidParamsError(
		`required: "hash" and either "chain-id" or both "token-id" and "issuer-id"`)
	ParamsErrorGetTransactions = jrpc.NewInvalidParamsError(
		`required: "hash" or "start" and either "chain-id" or both "token-id" and "issuer-id", "limit" must be greater than 0 if provided`)
	ParamsErrorGetNFToken = jrpc.NewInvalidParamsError(
		`required: "nf-token-id" and either "chain-id" or both "token-id" and "issuer-id"`)
	ParamsErrorGetBalance = jrpc.NewInvalidParamsError(
		`required: "fa-address" and either "chain-id" or both "token-id" and "issuer-id"`)
	ParamsErrorSendTransaction = jrpc.NewInvalidParamsError(
		`required: "rcd-sigs" and "tx" and either "chain-id" or both "token-id" and "issuer-id"`)

	ErrorTokenNotFound = jrpc.NewError(-32800, "Token Not Found",
		"token may be invalid, or not yet issued or tracked")
	ErrorInvalidAddress = jrpc.NewError(-32801, "Token Not Found",
		"token may be invalid, or not yet issued or tracked")
	ErrorTransactionNotFound = jrpc.NewError(-32803, "Transaction Not Found",
		"no matching tx-id was found")
	ErrorInvalidTransaction = jrpc.NewError(-32804, "Invalid Transaction", nil)
	ErrorTokenSyncing       = jrpc.NewError(-32805, "Token Syncing",
		"token is in the process of syncing")
	ErrorNoEC = jrpc.NewError(-32806, "No Entry Credits",
		"not configured with entry credits")
)
