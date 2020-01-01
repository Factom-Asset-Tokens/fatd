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
// extern void revert(void *ctx, int32_t msg, int32_t msg_len);
// extern void self_destruct(void *ctx);
import "C"
import (
	"fmt"
	"unsafe"

	"github.com/Factom-Asset-Tokens/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat"
	"github.com/Factom-Asset-Tokens/fatd/internal/db/address"
	"github.com/wasmerio/go-ext-wasm/wasmer"
)

//export get_height
func get_height(ptr unsafe.Pointer) int32 {
	ctx := intoContext(ptr)
	if ctx.Meter("get_height") != nil {
		return 0
	}
	return int32(ctx.DBlock.Height)
}

//export get_precision
func get_precision(ptr unsafe.Pointer) int32 {
	ctx := intoContext(ptr)
	if ctx.Meter("get_precision") != nil {
		return 0
	}
	return int32(ctx.Chain.Issuance.Precision)
}

//export get_amount
func get_amount(ptr unsafe.Pointer) int64 {
	ctx := intoContext(ptr)
	if ctx.Meter("get_amount") != nil {
		return 0
	}
	amount, _ := ctx.Amount()
	return int64(amount)
}

//export get_coinbase
func get_coinbase(ptr unsafe.Pointer, adr_buf int32) {
	ctx := intoContext(ptr)
	if ctx.Meter("get_coinbase") != nil {
		return
	}

	coinbase := fat.Coinbase()
	ctx.WriteAddress(&coinbase, adr_buf)
}

//export get_timestamp
func get_timestamp(ptr unsafe.Pointer) int64 {
	ctx := intoContext(ptr)
	if ctx.Meter("get_timestamp") != nil {
		return 0
	}

	return ctx.Transaction.Entry.Timestamp.Unix()
}

//export get_entry_hash
func get_entry_hash(ptr unsafe.Pointer, hash_buf int32) {
	ctx := intoContext(ptr)
	if ctx.Meter("get_entry_hash") != nil {
		return
	}

	ctx.WriteAddress(
		(*factom.FAAddress)(ctx.Transaction.Entry.Hash),
		hash_buf)
}

//export get_sender
func get_sender(ptr unsafe.Pointer, adr_buf int32) {
	ctx := intoContext(ptr)
	if ctx.Meter("get_sender") != nil {
		return
	}

	sender := ctx.Sender()
	ctx.WriteAddress(&sender, adr_buf)
}

//export get_address
func get_address(ptr unsafe.Pointer, adr_buf int32) {
	ctx := intoContext(ptr)
	if ctx.Meter("get_address") != nil {
		return
	}
	contract, err := ctx.ContractAddress()
	if err != nil {
		return
	}
	ctx.WriteAddress(&contract, adr_buf)
}

//export get_balance
func get_balance(ptr unsafe.Pointer) int64 {
	ctx := intoContext(ptr)
	if ctx.Meter("get_balance") != nil {
		return 0
	}

	bal, _ := ctx.ContractBalance()
	return int64(bal)
}

//export get_balance_of
func get_balance_of(ptr unsafe.Pointer, adr_buf int32) int64 {
	ctx := intoContext(ptr)
	if ctx.Meter("get_balance_of") != nil {
		return 0
	}

	adr, err := ctx.ReadAddress(adr_buf)
	if err != nil {
		return 0
	}
	_, bal, err := address.SelectIDBalance(ctx.Chain.Conn, &adr)
	if err != nil {
		ctx.Error(fmt.Errorf(
			"get_balance_of: address.SelectIDBalance: %w", err))
	}

	return int64(bal)
}

//export send
func send(ptr unsafe.Pointer, amount int64, adr_buf int32) {
	ctx := intoContext(ptr)
	if ctx.Meter("send") != nil {
		return
	}
	adr, err := ctx.ReadAddress(adr_buf)
	if err != nil {
		return
	}
	ctx.Send(uint64(amount), &adr)
}

//export burn
func burn(ptr unsafe.Pointer, amount int64) {
	ctx := intoContext(ptr)
	if ctx.Meter("burn") != nil {
		return
	}
	adr := fat.Coinbase()
	ctx.Send(uint64(amount), &adr)
}

//export revert
func revert(ptr unsafe.Pointer, msg int32, msg_len int32) {
	ctx := intoContext(ptr)
	if ctx.Meter("revert") != nil {
		return
	}
	if uint32(msg_len) > 256 {
		msg_len = 256
	}
	ctx.Revert(ctx.ReadString(msg, uint32(msg_len)))
}

//export self_destruct
func self_destruct(ptr unsafe.Pointer) {
	ctx := intoContext(ptr)
	if ctx.Meter("self_destruct") != nil {
		return
	}
	ctx.SelfDestruct()
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

	"ext_revert":        hostFunc{revert, C.revert},
	"ext_self_destruct": hostFunc{self_destruct, C.self_destruct},
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
