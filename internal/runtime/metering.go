package runtime

import (
	"sync/atomic"

	"github.com/wasmerio/go-ext-wasm/wasmer"
)

const ErrorExecLimitExceededString = "Execution limit exceeded."

type ErrorExecLimitExceeded struct{}

func (err ErrorExecLimitExceeded) Error() string {
	return ErrorExecLimitExceededString
}

var CallCount int64 = -1

func Meter(runtimeCtx wasmer.InstanceContext, cost uint64) {
	if CallCount != -1 {
		atomic.AddInt64(&CallCount, 1)
	}
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
