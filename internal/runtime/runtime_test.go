package runtime_test

import (
	"io/ioutil"
	"testing"

	"github.com/Factom-Asset-Tokens/fatd/internal/runtime"
	"github.com/Factom-Asset-Tokens/fatd/internal/runtime/testdata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wasmerio/go-ext-wasm/wasmer"
)

func TestRuntime(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	// Load wasm module
	wasm, err := ioutil.ReadFile("./testdata/api_test.wasm")
	require.NoError(err)
	mod, err := wasmer.CompileWithGasMetering(wasm)
	require.NoError(err)

	vm, err := runtime.NewVM(&mod)
	require.NoError(err)
	defer vm.Close()

	// Set the limit to the exact amount of gas required to complete the
	// function call. This must be updated if api.wasm changes.
	const PointsUsed = 427
	vm.SetExecLimit(PointsUsed)

	ctx := testdata.Context()

	vm.SetContextData(ctx)
	v, err := vm.Call("run_all")
	require.NoErrorf(err, "points used: %v", int64(vm.GetPointsUsed()))
	require.Equal(wasmer.TypeI32, v.GetType())
	assert.Equal(testdata.ErrMap[0], testdata.ErrMap[v.ToI32()])
	require.Equal(int64(PointsUsed), int64(vm.GetPointsUsed()))

	vm.SetPointsUsed(0)
	vm.SetExecLimit(0)
	_, err = vm.Call("run_all")
	require.EqualError(err, runtime.ErrorExecLimitExceededString)

	// By setting the limit to the points used, this should cause the
	// execution limit to be exceeded within the first host function call.
	pointsUsed := vm.GetPointsUsed()
	vm.SetExecLimit(pointsUsed)
	vm.SetPointsUsed(0)
	_, err = vm.Call("run_all")
	require.EqualError(err, runtime.ErrorExecLimitExceededString)
	require.Equal(int64(pointsUsed+runtime.GetHeightCost),
		int64(vm.GetPointsUsed()))
}
