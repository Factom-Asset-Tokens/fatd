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

// #include <stdlib.h>
//
// extern int32_t get_height(void *ctx);
// extern int32_t get_sender(void *ctx, int32_t adrBuf);
import "C"
import (
	"fmt"
	"unsafe"

	"github.com/Factom-Asset-Tokens/factom"
	"github.com/wasmerio/go-ext-wasm/wasmer"
)

//export get_height
func get_height(ctx unsafe.Pointer) int32 {
	instanceCtx := wasmer.IntoInstanceContext(ctx)
	return int32(instanceCtx.Data().(factom.EBlock).Height)
}

//export get_sender
func get_sender(ctx unsafe.Pointer, adrBuf int32) int32 {
	return -1
}

func imports() (*wasmer.Imports, error) {
	i := wasmer.NewImports()
	i = i.Namespace("env")
	i, err := i.Append("get_height", get_height, C.get_height)
	if err != nil {
		return nil, fmt.Errorf("wasmer.Imports.Append(%q): %w", "get_height", err)
	}
	i, err = i.Append("get_sender", get_sender, C.get_sender)
	if err != nil {
		return nil, fmt.Errorf("wasmer.Imports.Append(%q): %w", "get_sender", err)
	}
	return i, nil
}

//getSender 	Get the calling tx's input Factoid address 	void 	char * 	extern char * getInput(void);
//getAmount 	Get the calling tx's output amount 	void 	int
//getEntryhash 	Get the calling tx's Factom Entryhash 	void 	char *
//getTimestamp 	Get the unix timestamp of the calling tx 	void 	int
//getHeight 	Get the Factom blockheight of the calling tx 	void 	int
//getAddress 	Get the Factoid address of this contract 	void 	char *
//
//getPrecision 	Get the decimal precision of the host token 	void 	int
//getBalance 	Get the FAT-0 balance of a Factoid address on the host token 	char * - The Factoid address string 	int
//send 	Send FAT-0 tokens from the contracts balance 	char * - The Factoid address string destination, int - The amount of tokens to send in base units 	int - The boolean success value of the operation
//burn 	Burn the specified amount of tokens from the contracts balance 	int - The amount of tokens to burn 	void
//
//revert 	Revert the current contract calls state changes and abort the call. Will still charge the input amount 	void 	void
//invalidate 	Invalidate the calling transaction and abort state changes. Refunds input amount to caller
//selfDestruct 	Terminate the current contract, liquidating the FAT-0 balance to a Factoid address 	char * - the liquidation destination Factoid address 	void
