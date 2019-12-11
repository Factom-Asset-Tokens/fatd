package runtime

import (
	"fmt"

	"github.com/wasmerio/go-ext-wasm/wasmer"
)

const ErrorExecLimitExceededString = "Execution limit exceeded."

type ErrorExecLimitExceeded struct{}

func (err ErrorExecLimitExceeded) Error() string {
	return ErrorExecLimitExceededString
}

var Called map[string]struct{}
var Cost = map[string]uint64{
	"get_height":     1,
	"get_precision":  1,
	"get_amount":     1,
	"get_timestamp":  1,
	"get_entry_hash": 1,
	"get_sender":     1,
	"get_address":    1,
	"get_coinbase":   1,
	"get_balance":    1,
	"get_balance_of": 1,
	"send":           1,
	"burn":           1,
}

func Meter(runtimeCtx wasmer.InstanceContext, fname string) {
	if Called != nil {
		Called[fname] = struct{}{}
	}

	cost, ok := Cost[fname]
	if !ok {
		panic(fmt.Errorf("missing cost for %q", fname))
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
