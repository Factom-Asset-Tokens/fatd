package runtime

import (
	"github.com/wasmerio/go-ext-wasm/wasmer"
)

const ErrorExecLimitExceededString = "Execution limit exceeded."

type ErrorExecLimitExceeded struct{}

func (err ErrorExecLimitExceeded) Error() string {
	return ErrorExecLimitExceededString
}

func Meter(runtimeCtx wasmer.InstanceContext, cost uint64) {
	used := runtimeCtx.GetPointsUsed() + cost
	runtimeCtx.SetPointsUsed(used)

	limit := runtimeCtx.GetExecLimit()
	if used > limit {
		panic(ErrorExecLimitExceeded{})
	}
}

func RecoverOutOfGas(err *error) {
	if ret := recover(); ret != nil {
		var ok bool
		*err, ok = ret.(ErrorExecLimitExceeded)
		if !ok {
			panic(ret)
		}
	}
}
