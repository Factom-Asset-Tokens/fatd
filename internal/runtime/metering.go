package runtime

import (
	"fmt"

	"github.com/wasmerio/go-ext-wasm/wasmer"
)

type ErrorOutOfGas struct{}

func (err ErrorOutOfGas) Error() string {
	return fmt.Sprintf("out of gas")
}

func Meter(runtimeCtx wasmer.InstanceContext, cost uint64) {
	used := runtimeCtx.GetPointsUsed() + cost
	runtimeCtx.SetPointsUsed(used)

	limit := runtimeCtx.GetExecLimit()
	if used > limit {
		panic(ErrorOutOfGas{})
	}
}

func RecoverOutOfGas(err *error) {
	if ret := recover(); ret != nil {
		var ok bool
		*err, ok = ret.(ErrorOutOfGas)
		if !ok {
			panic(ret)
		}
	}
}
