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

package runtime

// #include <stdint.h>
// #define int64_t long long
//
// extern int32_t get_height(void *ctx);
// extern int32_t get_precision(void *ctx);
// extern int64_t get_amount(void *ctx);
// extern int64_t get_timestamp(void *ctx);
//
// extern int64_t get_balance(void *ctx, int32_t buf);
//
// extern void get_sender(void *ctx, int32_t buf);
// extern void get_entry_hash(void *ctx, int32_t buf);
// extern void get_address(void *ctx, int32_t buf);
import "C"

//send 	Send FAT-0 tokens from the contracts balance 	char * - The Factoid address string destination, int - The amount of tokens to send in base units 	int - The boolean success value of the operation
//burn 	Burn the specified amount of tokens from the contracts balance 	int - The amount of tokens to burn 	void
//
//revert 	Revert the current contract calls state changes and abort the call. Will still charge the input amount 	void 	void
//invalidate 	Invalidate the calling transaction and abort state changes. Refunds input amount to caller
//selfDestruct 	Terminate the current contract, liquidating the FAT-0 balance to a Factoid address 	char * - the liquidation destination Factoid address 	void
import (
	"context"
	"fmt"
	"unsafe"

	"github.com/Factom-Asset-Tokens/factom"
	"github.com/Factom-Asset-Tokens/factom/fat0"
	"github.com/Factom-Asset-Tokens/fatd/internal/db"
	"github.com/wasmerio/go-ext-wasm/wasmer"
)

type Context struct {
	db.Chain
	factom.DBlock
	fat0.Transaction

	ctx context.Context
}

func intoContext(ctx interface{}) Context {
	return ctx.(Context)
}

const (
	CostGetHeight    = 1
	CostGetPrecision = 1
	CostGetAmount    = 1
	CostGetTimestamp = 1

	CostGetEntryHash = 1
	CostGetSender    = 1
	CostGetAddress   = 1
)

//export get_height
func get_height(ctx unsafe.Pointer) int32 {
	instanceCtx := wasmer.IntoInstanceContext(ctx)
	Meter(instanceCtx, CostGetHeight)

	context := intoContext(instanceCtx.Data())
	return int32(context.DBlock.Height)
}

//export get_precision
func get_precision(ctx unsafe.Pointer) int32 {
	instanceCtx := wasmer.IntoInstanceContext(ctx)
	Meter(instanceCtx, CostGetPrecision)

	context := intoContext(instanceCtx.Data())
	return int32(context.Chain.Issuance.Precision)
}

//export get_amount
func get_amount(ctx unsafe.Pointer) int64 {
	instanceCtx := wasmer.IntoInstanceContext(ctx)
	Meter(instanceCtx, CostGetAmount)

	context := intoContext(instanceCtx.Data())

	var amount uint64
	for _, amount = range context.Transaction.Outputs {
	}
	return int64(amount)
}

//export get_timestamp
func get_timestamp(ctx unsafe.Pointer) int64 {
	instanceCtx := wasmer.IntoInstanceContext(ctx)
	Meter(instanceCtx, CostGetTimestamp)

	context := intoContext(instanceCtx.Data())

	return context.Transaction.Entry.Timestamp.Unix()
}

//export get_entry_hash
func get_entry_hash(ctx unsafe.Pointer, buf int32) {
	instanceCtx := wasmer.IntoInstanceContext(ctx)
	Meter(instanceCtx, CostGetEntryHash)

	context := intoContext(instanceCtx.Data())

	mem := instanceCtx.Memory()
	copy(mem.Data()[buf:], context.Transaction.Entry.Hash[:])
}

//export get_sender
func get_sender(ctx unsafe.Pointer, buf int32) {
	instanceCtx := wasmer.IntoInstanceContext(ctx)
	Meter(instanceCtx, CostGetSender)

	context := intoContext(instanceCtx.Data())

	var sender factom.FAAddress
	for sender, _ = range context.Transaction.Inputs {
	}

	mem := instanceCtx.Memory()
	copy(mem.Data()[buf:], sender[:])
}

//export get_address
func get_address(ctx unsafe.Pointer, buf int32) {
	instanceCtx := wasmer.IntoInstanceContext(ctx)
	Meter(instanceCtx, CostGetAddress)

	context := intoContext(instanceCtx.Data())

	var adr factom.FAAddress
	for adr, _ = range context.Transaction.Outputs {
	}

	mem := instanceCtx.Memory()
	copy(mem.Data()[buf:], adr[:])
}

type hostFunc struct {
	Func  interface{}
	cFunc unsafe.Pointer
}

var hostFuncs = map[string]hostFunc{
	"ext_get_timestamp": hostFunc{get_timestamp, C.get_timestamp},
	"ext_get_height":    hostFunc{get_height, C.get_height},
	"ext_get_precision": hostFunc{get_precision, C.get_precision},
	"ext_get_amount":    hostFunc{get_amount, C.get_amount},

	"ext_get_sender":     hostFunc{get_sender, C.get_sender},
	"ext_get_address":    hostFunc{get_address, C.get_address},
	"ext_get_entry_hash": hostFunc{get_entry_hash, C.get_entry_hash},
}

var NumHostFuncs = len(hostFuncs)

func imports() (*wasmer.Imports, error) {
	i := wasmer.NewImports()
	i = i.Namespace("env")
	for name, f := range hostFuncs {
		var err error
		i, err = i.Append(name, f.Func, f.cFunc)
		if err != nil {
			return nil, fmt.Errorf(
				"wasmer.Imports.Append(%q): %w", name, err)
		}
	}
	return i, nil
}

//
