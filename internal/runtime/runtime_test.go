// MIT License
//
// Copyright 2019 Canonical Ledgers, LLC
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
// IN THE SOFTWARE.

package runtime_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"testing"

	"crawshaw.io/sqlite/sqlitex"
	"github.com/Factom-Asset-Tokens/fatd/internal/db"
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
	const PointsUsed = 878
	vm.SetExecLimit(PointsUsed)

	chain, err := db.Open(context.Background(), "./testdata/test-fatd.db/",
		"b54c4310530dc4dd361101644fa55cb10aec561e7874a7b786ea3b66f2c6fdfb.sqlite3")
	if err != nil {
		panic(err)
	}
	defer chain.Close()
	release := sqlitex.Save(chain.Conn)
	rbErr := fmt.Errorf("rollback")
	defer release(&rbErr)

	ctx := testdata.Context(chain)

	runtime.Called = make(map[string]struct{}, len(runtime.Cost))

	vm.SetContextData(&ctx)
	v, err := vm.Call("run_all")
	require.NoErrorf(err, "points used: %v", int64(vm.GetPointsUsed()))
	require.Equal(wasmer.TypeI32, v.GetType())
	assert.Equalf(testdata.ErrMap[0], testdata.ErrMap[v.ToI32()],
		"ret: %v", v.ToI32())
	assert.Equal(int64(PointsUsed), int64(vm.GetPointsUsed()))
	require.Equal(len(runtime.Called), len(runtime.Cost),
		"Not all host funcs were called")

	runtime.Called = nil

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
	require.Equal(int64(pointsUsed+runtime.Cost["get_height"]),
		int64(vm.GetPointsUsed()))
}
