package srv

import (
	jrpc "github.com/AdamSLevy/jsonrpc2/v2"
)

var version = jrpc.MethodFunc(func(params interface{}) jrpc.Response {
	if params != nil {
		return jrpc.NewInvalidParamsErrorResponse("Unexpected parameters")
	}
	return jrpc.NewResponse("0.0.0")
})
