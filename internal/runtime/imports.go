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
// extern int64_t get_balance(void *ctx);
// extern int64_t get_balance_of(void *ctx, int32_t adr_buf);
//
// extern void get_sender(void *ctx, int32_t adr_buf);
// extern void get_entry_hash(void *ctx, int32_t hash_buf);
// extern void get_address(void *ctx, int32_t adr_buf);
// extern void get_coinbase(void *ctx, int32_t adr_buf);
//
// extern void send(void *ctx, int64_t amount, int32_t adr_buf);
// extern void burn(void *ctx, int64_t amount);
//
// extern void revert(void *ctx);
// extern void self_destruct(void *ctx);
import "C"
import (
	"fmt"
	"unsafe"

	"github.com/Factom-Asset-Tokens/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat"
	"github.com/Factom-Asset-Tokens/fatd/internal/db/addresses"
	"github.com/wasmerio/go-ext-wasm/wasmer"
)

func readAddress(instanceCtx *wasmer.InstanceContext, adr_buf int32) factom.FAAddress {
	var adr factom.FAAddress
	if 32 != copy(adr[:], instanceCtx.Memory().Data()[adr_buf:]) {
		panic(fmt.Errorf("readAddress: invalid copy length"))
	}
	return adr
}

func writeAddress(instanceCtx *wasmer.InstanceContext,
	adr *factom.FAAddress, adr_buf int32) {
	if 32 != copy(instanceCtx.Memory().Data()[adr_buf:], adr[:]) {
		panic(fmt.Errorf("writeAddress: invalid copy length"))
	}
}

//export get_height
func get_height(ctx unsafe.Pointer) int32 {
	instanceCtx := wasmer.IntoInstanceContext(ctx)
	Meter(instanceCtx, "get_height")

	context := intoContext(&instanceCtx)
	return int32(context.DBlock.Height)
}

//export get_precision
func get_precision(ctx unsafe.Pointer) int32 {
	instanceCtx := wasmer.IntoInstanceContext(ctx)
	Meter(instanceCtx, "get_precision")

	context := intoContext(&instanceCtx)
	return int32(context.Chain.Issuance.Precision)
}

//export get_amount
func get_amount(ctx unsafe.Pointer) int64 {
	instanceCtx := wasmer.IntoInstanceContext(ctx)
	Meter(instanceCtx, "get_amount")

	context := intoContext(&instanceCtx)
	return int64(context.Amount())
}

//export get_coinbase
func get_coinbase(ctx unsafe.Pointer, adr_buf int32) {
	instanceCtx := wasmer.IntoInstanceContext(ctx)
	Meter(instanceCtx, "get_coinbase")

	coinbase := fat.Coinbase()
	writeAddress(&instanceCtx, &coinbase, adr_buf)
}

//export get_timestamp
func get_timestamp(ctx unsafe.Pointer) int64 {
	instanceCtx := wasmer.IntoInstanceContext(ctx)
	Meter(instanceCtx, "get_timestamp")

	context := intoContext(&instanceCtx)

	return context.Transaction.Entry.Timestamp.Unix()
}

//export get_entry_hash
func get_entry_hash(ctx unsafe.Pointer, hash_buf int32) {
	instanceCtx := wasmer.IntoInstanceContext(ctx)
	Meter(instanceCtx, "get_entry_hash")

	context := intoContext(&instanceCtx)

	writeAddress(&instanceCtx, (*factom.FAAddress)(context.Transaction.Entry.Hash),
		hash_buf)
}

//export get_sender
func get_sender(ctx unsafe.Pointer, adr_buf int32) {
	instanceCtx := wasmer.IntoInstanceContext(ctx)
	Meter(instanceCtx, "get_sender")

	context := intoContext(&instanceCtx)

	sender := context.Sender()
	writeAddress(&instanceCtx, &sender, adr_buf)
}

//export get_address
func get_address(ctx unsafe.Pointer, adr_buf int32) {
	instanceCtx := wasmer.IntoInstanceContext(ctx)
	Meter(instanceCtx, "get_address")

	context := intoContext(&instanceCtx)

	contract := context.ContractAddress()

	writeAddress(&instanceCtx, &contract, adr_buf)
}

//export get_balance
func get_balance(ctx unsafe.Pointer) int64 {
	instanceCtx := wasmer.IntoInstanceContext(ctx)
	Meter(instanceCtx, "get_balance")

	context := intoContext(&instanceCtx)
	contract := context.ContractAddress()

	_, bal, err := addresses.SelectIDBalance(context.Chain.Conn, &contract)
	if err != nil {
		panic(fmt.Errorf("get_balance: addresses.SelectIDBalance: %w", err))
	}

	return int64(bal)
}

//export get_balance_of
func get_balance_of(ctx unsafe.Pointer, adr_buf int32) int64 {
	instanceCtx := wasmer.IntoInstanceContext(ctx)
	Meter(instanceCtx, "get_balance_of")

	context := intoContext(&instanceCtx)

	adr := readAddress(&instanceCtx, adr_buf)

	_, bal, err := addresses.SelectIDBalance(context.Chain.Conn, &adr)
	if err != nil {
		panic(fmt.Errorf("get_balance: addresses.SelectIDBalance: %w", err))
	}

	return int64(bal)
}

//export send
func send(ctx unsafe.Pointer, amount int64, adr_buf int32) {
	instanceCtx := wasmer.IntoInstanceContext(ctx)
	Meter(instanceCtx, "send")

	adr := readAddress(&instanceCtx, adr_buf)

	context := intoContext(&instanceCtx)
	context.Send(uint64(amount), &adr)
}

//export burn
func burn(ctx unsafe.Pointer, amount int64) {
	instanceCtx := wasmer.IntoInstanceContext(ctx)
	Meter(instanceCtx, "burn")

	context := intoContext(&instanceCtx)
	adr := fat.Coinbase()
	context.Send(uint64(amount), &adr)
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
	"ext_get_coinbase":   hostFunc{get_coinbase, C.get_coinbase},
	"ext_get_entry_hash": hostFunc{get_entry_hash, C.get_entry_hash},

	"ext_get_balance":    hostFunc{get_balance, C.get_balance},
	"ext_get_balance_of": hostFunc{get_balance_of, C.get_balance_of},

	"ext_send": hostFunc{send, C.send},
	"ext_burn": hostFunc{burn, C.burn},
}

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
