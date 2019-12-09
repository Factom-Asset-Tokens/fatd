package runtime

import (
	"io/ioutil"
	"testing"

	"github.com/Factom-Asset-Tokens/factom"
	"github.com/stretchr/testify/require"
	"github.com/wasmerio/go-ext-wasm/wasmer"
)

func TestRuntime(t *testing.T) {
	require := require.New(t)

	// Load wasm module
	wasm, err := ioutil.ReadFile("./testdata/api.wasm")
	require.NoError(err)
	mod, err := wasmer.CompileWithGasMetering(wasm)
	require.NoError(err)

	vm, err := NewVM(&mod)
	require.NoError(err)
	defer vm.Close()

	// Set the limit to the exact amount of gas required to complete the
	// function call.
	vm.SetExecLimit(5)

	ctx := Context{DBlock: factom.DBlock{Height: 1001}}
	vm.SetContextData(ctx)
	v, err := vm.Call("run_all")
	require.NoError(err)
	require.Equal(wasmer.TypeI32, v.GetType())
	require.Equal(int32(0), v.ToI32())
	require.Equal(int64(5), int64(vm.GetPointsUsed()))

	// This should cause the exec limit to be exceeded immediately.
	_, err = vm.Call("run_all")
	require.EqualError(err, ErrorExecLimitExceededString)
	require.Equal(int64(9), int64(vm.GetPointsUsed()))

	// This should cause the limit to be exceeded in the first host func
	// call.
	vm.SetPointsUsed(1)
	_, err = vm.Call("run_all")
	require.EqualError(err, ErrorExecLimitExceededString)
	require.Equal(int64(6), int64(vm.GetPointsUsed()))
}
