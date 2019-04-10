package srv

import jrpc "github.com/AdamSLevy/jsonrpc2/v10"

var (
	ParamsErrorNoParams = jrpc.NewInvalidParamsError(`no "params" accepted`)
	ParamsErrorToken    = jrpc.NewInvalidParamsError(
		`required: either "chainid" or both "tokenid" and "issuerid"`)
	ParamsErrorGetTransaction = jrpc.NewInvalidParamsError(
		`required: "entryhash" and either "chainid" or both "tokenid" and "issuerid"`)
	ParamsErrorGetTransactions = jrpc.NewInvalidParamsError(
		`required: "hash" or "start" and either "chainid" or both "tokenid" and "issuerid", "limit" must be greater than 0 if provided`)
	ParamsErrorGetNFToken = jrpc.NewInvalidParamsError(
		`required: "nftokenid" and either "chainid" or both "tokenid" and "issuerid"`)
	ParamsErrorGetBalance = jrpc.NewInvalidParamsError(
		`required: "address" and either "chainid" or both "tokenid" and "issuerid"`)
	ParamsErrorSendTransaction = jrpc.NewInvalidParamsError(
		`required: "rcd-sigs" and "tx" and either "chainid" or both "tokenid" and "issuerid"`)

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
