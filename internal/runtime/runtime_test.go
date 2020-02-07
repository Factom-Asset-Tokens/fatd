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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"crawshaw.io/sqlite"
	"crawshaw.io/sqlite/sqlitex"
	"github.com/Factom-Asset-Tokens/factom/fat"
	"github.com/Factom-Asset-Tokens/factom/fat0"
	"github.com/Factom-Asset-Tokens/factom/fat104"
	"github.com/Factom-Asset-Tokens/fatd/internal/db"
	"github.com/Factom-Asset-Tokens/fatd/internal/db/address"
	"github.com/Factom-Asset-Tokens/fatd/internal/runtime"
	"github.com/Factom-Asset-Tokens/fatd/internal/runtime/testdata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wasmerio/go-ext-wasm/wasmer"
)

const TotalPoints = 876

var (
	wasm, modBin []byte
	chain        db.FATChain
	sess         *sqlite.Session
	ctx          runtime.Context

	rbErr = fmt.Errorf("rollback")
)

func TestMain(m *testing.M) {
	os.Exit(mainRun(m))
}

func mainRun(m *testing.M) int {
	var err error
	// Load wasm module
	wasm, err = ioutil.ReadFile("./testdata/api_test.wasm")
	if err != nil {
		panic(err)
	}
	mod, err := wasmer.CompileWithGasMetering(wasm)
	if err != nil {
		panic(err)
	}
	defer mod.Close()
	modBin, err = mod.Serialize()
	if err != nil {
		panic(err)
	}

	// Open up the chain so we can set up the runtime.Context
	chain, err = db.OpenFATChain(context.Background(),
		"./testdata/test-fatd.db/",
		"b54c4310530dc4dd361101644fa55cb10aec561e7874a7b786ea3b66f2c6fdfb.sqlite3")
	if err != nil {
		panic(err)
	}
	defer chain.Close()

	// Rollback all changes made during the tests.
	defer sqlitex.Save(chain.Conn)(&rbErr)

	// Set up our context.
	ctx = testdata.Context(&chain)

	// Start a session so we can ensure that changes actually occur to the
	// DB.
	sess, err = chain.Conn.CreateSession("")
	if err != nil {
		panic(err)
	}
	defer sess.Delete()
	if err := sess.Attach(""); err != nil {
		panic(err)
	}

	return m.Run()
}

var (
	runAllFunc           = fat104.Func{Name: "run_all"}
	testRevertFunc       = fat104.Func{Name: "test_revert"}
	testSelfDestructFunc = fat104.Func{Name: "test_self_destruct"}
	testSendFunc         = fat104.Func{Name: "test_send",
		Args: []fat104.Type{fat104.TypeI64}}
)

func TestRunAll(t *testing.T) {
	reqr, vm, cleanUp := setupTest(t, true)
	defer cleanUp()
	asrt := assert.New(t)

	// Changes should have occurred to the database.
	var buf bytes.Buffer
	reqr.NoError(sess.Changeset(&buf))
	reqr.Zero(buf.Len(), "expected no changes")

	// Set the limit to the exact amount of gas required to complete the
	// function call. This must be updated if api.wasm changes.
	vm.SetExecLimit(TotalPoints)
	vm.SetPointsUsed(0)

	// This should return successfully.
	v, txErr, err := vm.Call(&ctx, &runAllFunc)
	reqr.NoError(err, "err")
	reqr.NoErrorf(txErr, "txErr: points used: %v",
		int64(vm.GetPointsUsed()))

	// The return value should be SUCCESS but this tells us what error in
	// the API test we hit.
	reqr.Equal(wasmer.TypeI32, v.GetType())
	asrt.Equalf(testdata.ErrMap[0], testdata.ErrMap[v.ToI32()],
		"ret: %v", v.ToI32())

	// We should have used all of the points we set earlier.
	asrt.Equal(int64(TotalPoints), int64(vm.GetPointsUsed()))

	// All host functions should have been called, except revert and
	// self_destruct.
	reqr.Equal(len(runtime.Cost)-2, len(runtime.Called),
		"Not all host funcs were called")

	// Changes should have occurred to the database.
	buf.Truncate(0)
	reqr.NoError(sess.Changeset(&buf))
	reqr.NotZero(buf.Len(), "expected changes")
}

func TestOutOfGas(t *testing.T) {
	t.Run("zero_limit", func(t *testing.T) {
		reqr, vm, cleanUp := setupTest(t, false)
		defer cleanUp()

		// No execution should occur with a 0 limit.
		vm.SetExecLimit(0)
		vm.SetPointsUsed(0)
		_, txErr, err := vm.Call(&ctx, &runAllFunc)
		reqr.NoError(err)
		reqr.EqualError(txErr, runtime.ErrorExecLimitExceededString)
	})

	t.Run("from_host", func(t *testing.T) {
		reqr, vm, cleanUp := setupTest(t, true)
		defer cleanUp()

		// This is enough points to make it into the first host
		// function, but not through it.
		vm.SetExecLimit(12)
		vm.SetPointsUsed(0)

		runtime.Cost["get_timestamp"] = 5000
		defer func() { runtime.Cost["get_timestamp"] = 1 }()
		_, txErr, err := vm.Call(&ctx, &runAllFunc)
		reqr.NoError(err)
		reqr.EqualError(txErr, runtime.ErrorExecLimitExceededString)
		// The points should be equal to the last pointsUsed plus the cost of
		// the first called host function.
		reqr.Equal(int64(12+runtime.Cost["get_timestamp"]),
			int64(vm.GetPointsUsed()))
	})
	t.Run("revert_changes", func(t *testing.T) {
		reqr, vm, cleanUp := setupTest(t, true)
		defer cleanUp()

		vm.SetExecLimit(46)
		vm.SetPointsUsed(0)

		_, txErr, err := vm.Call(&ctx, &testSendFunc, json.RawMessage(`1`))
		reqr.NoError(err)
		// Ensure send was successfully called
		reqr.Contains(runtime.Called, "send")
		reqr.EqualErrorf(txErr, runtime.ErrorExecLimitExceededString,
			"txErr: points used: %v", int64(vm.GetPointsUsed()))
	})
}

func TestSend(t *testing.T) {
	t.Run("low_balance", func(t *testing.T) {
		reqr, vm, cleanUp := setupTest(t, true)
		defer cleanUp()

		_, txErr, err := vm.Call(&ctx, &testSendFunc,
			json.RawMessage(`999999999999999`))
		reqr.NoError(err)
		// Ensure send was successfully called
		reqr.EqualErrorf(txErr,
			runtime.ErrorRevert{"send: insufficient balance"}.Error(),
			"txErr: points used: %v", int64(vm.GetPointsUsed()))
		reqr.Contains(runtime.Called, "send")
		// Ensure no changes.
		var buf bytes.Buffer
		reqr.NoError(sess.Changeset(&buf))
		reqr.Zero(buf.Len(), "changes not reverted")
	})
	t.Run("success", func(t *testing.T) {
		reqr, vm, cleanUp := setupTest(t, true)
		defer cleanUp()

		bal, err := ctx.ContractBalance()
		reqr.NoError(err)
		reqr.NotZero(bal)

		_, txErr, err := vm.Call(&ctx, &testSendFunc, json.RawMessage(`1`))
		reqr.NoError(err)
		reqr.NoError(txErr)
		// Ensure send was successfully called
		reqr.Contains(runtime.Called, "send")
		newBal, err := ctx.ContractBalance()
		reqr.NoError(err)
		reqr.Equal(bal-1, newBal)
	})
	t.Run("coinbase/low_supply", func(t *testing.T) {
		reqr, vm, cleanUp := setupTest(t, true)
		defer cleanUp()

		outputs := ctx.Transaction.Outputs
		ctx.Transaction.Outputs = fat0.AddressAmountMap{fat.Coinbase(): 0}
		defer func() {
			ctx.Transaction.Outputs = outputs
		}()

		_, txErr, err := vm.Call(&ctx, &testSendFunc,
			json.RawMessage(`999999999999999`))
		reqr.NoError(err)
		// Ensure send was successfully called
		reqr.Contains(runtime.Called, "send")
		reqr.Equalf(runtime.ErrorRevert{"send: max supply exceeded"}, txErr,
			"txErr: points used: %v", int64(vm.GetPointsUsed()))
		// Ensure no changes.
		var buf bytes.Buffer
		reqr.NoError(sess.Changeset(&buf))
		reqr.Zero(buf.Len(), "changes not reverted")
	})

	t.Run("coinbase/success", func(t *testing.T) {
		reqr, vm, cleanUp := setupTest(t, true)
		defer cleanUp()

		outputs := ctx.Transaction.Outputs
		ctx.Transaction.Outputs = fat0.AddressAmountMap{fat.Coinbase(): 0}
		defer func() {
			ctx.Transaction.Outputs = outputs
		}()

		bal, err := ctx.ContractBalance()
		reqr.NoError(err)
		remSupply := ctx.Chain.Issuance.Supply - int64(ctx.Chain.NumIssued)
		reqr.Equal(remSupply, int64(bal))

		_, txErr, err := vm.Call(&ctx, &testSendFunc, json.RawMessage(`1`))
		reqr.NoError(err)
		reqr.NoError(txErr)
		// Ensure send was successfully called
		reqr.Contains(runtime.Called, "send")
		newBal, err := ctx.ContractBalance()
		reqr.NoError(err)
		reqr.Equal(int(bal-1), int(newBal))
	})
}

func TestRevert(t *testing.T) {
	reqr, vm, cleanUp := setupTest(t, true)
	defer cleanUp()

	_, txErr, err := vm.Call(&ctx, &testRevertFunc)
	reqr.NoError(err)
	// Ensure send was successfully called
	reqr.Contains(runtime.Called, "send")
	reqr.EqualErrorf(txErr,
		runtime.ErrorRevert{"test_revert"}.Error(),
		"txErr: points used: %v", int64(vm.GetPointsUsed()))
}

func TestSelfDestruct(t *testing.T) {
	reqr, vm, cleanUp := setupTest(t, true)
	defer cleanUp()

	_, txErr, err := vm.Call(&ctx, &testSelfDestructFunc)
	reqr.NoError(err)
	reqr.NoError(txErr)
	// Ensure send was successfully called
	reqr.Contains(runtime.Called, "send")
	reqr.Contains(runtime.Called, "self_destruct")

	var buf bytes.Buffer
	reqr.NoError(sess.Changeset(&buf))
	reqr.NotZero(buf.Len(), "changes reverted")

	ctrct, err := ctx.ContractAddress()
	reqr.NoError(err)
	id, err := address.SelectID(ctx.Chain.Conn, &ctrct)
	reqr.NoError(err)
	_, chainID, err := address.SelectContract(ctx.Chain.Conn, id)
	reqr.NoError(err)
	reqr.True(chainID.IsZero())

}

func setupTest(t *testing.T, trackCalled bool) (reqr *require.Assertions, vm *runtime.VM, cleanUp func()) {
	reqr = require.New(t)
	// Rollback all changes made during the tests.
	release := sqlitex.Save(chain.Conn)
	defer func() {
		if cleanUp == nil {
			release(&rbErr)
		}
	}()
	mod, err := wasmer.CompileWithGasMetering(wasm)
	reqr.NoError(err)
	defer func() {
		if cleanUp == nil {
			mod.Close()
		}
	}()

	vm, err = runtime.NewVM(&mod)
	reqr.NoError(err)
	if trackCalled {
		runtime.Called = make(map[string]struct{}, len(runtime.Cost))
	}
	vm.SetPointsUsed(0)
	vm.SetExecLimit(1000)
	ctx.Err = nil
	return reqr, vm, func() {
		runtime.Called = nil
		vm.Close()
		mod.Close()
		release(&rbErr)
	}
}
