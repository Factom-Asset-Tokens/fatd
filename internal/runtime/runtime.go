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

import (
	"encoding/json"
	"fmt"

	"github.com/Factom-Asset-Tokens/factom"
	"github.com/Factom-Asset-Tokens/factom/fat104"
	"github.com/wasmerio/go-ext-wasm/wasmer"
)

type VM struct {
	wasmer.Instance
}

func NewVM(mod *wasmer.Module) (*VM, error) {
	imp, err := imports()
	if err != nil {
		return nil, err
	}

	inst, err := mod.InstantiateWithImports(imp)
	if err != nil {
		return nil, fmt.Errorf("wasmer.Module.InstantiateWithImports(): %w", err)
	}
	return &VM{inst}, nil
}

func (vm *VM) Call(ctx *Context,
	abi *fat104.Func, args ...json.RawMessage) (v wasmer.Value, txErr, err error) {

	f, ok := vm.Exports[abi.Name]
	if !ok {
		err = fmt.Errorf("function %q not defined", abi.Name)
		return
	}

	if len(abi.Args) != len(args) {
		txErr = fmt.Errorf("invalid number of args")
		return
	}

	var wasmArgs []interface{}
	var ptr int32
	for i, arg := range abi.Args {
		switch arg {
		case fat104.TypeI32:
			var val int32
			if txErr = json.Unmarshal(args[i], &val); txErr != nil {
				return
			}
			wasmArgs = append(wasmArgs, val)
		case fat104.TypeI64:
			var val int64
			if txErr = json.Unmarshal(args[i], &val); txErr != nil {
				return
			}
			wasmArgs = append(wasmArgs, val)
		case fat104.TypeString:
			var val string
			if txErr = json.Unmarshal(args[i], &val); txErr != nil {
				return
			}

			if copy(vm.Memory.Data()[ptr:], append([]byte(val), 0)) !=
				len(val)+1 {
				err = fmt.Errorf("couldn't copy arg %v into mem", i)
			}
			wasmArgs = append(wasmArgs, ptr)
			ptr += int32(len(val) + 1)
		case fat104.TypeBytes:
			var val factom.Bytes
			if txErr = json.Unmarshal(args[i], &val); txErr != nil {
				return
			}

			if copy(vm.Memory.Data()[ptr:], val) !=
				len(val) {
				err = fmt.Errorf("couldn't copy arg %v into mem", i)
			}
			n := int32(len(val))
			wasmArgs = append(wasmArgs, ptr, n)
			ptr += n
		}
	}

	vm.SetContextData(ctx)

	v, err = f(wasmArgs...)
	if err != nil {
		if err.Error() == fmt.Sprintf(
			"Failed to call the `%s` exported function.", abi.Name) {
			var errStr string
			errStr, err = wasmer.GetLastError()
			if err == nil {
				if errStr != ErrorExecLimitExceededString {
					err = fmt.Errorf(errStr)
					return
				}
				txErr = ErrorExecLimitExceeded{}
			}
		}
	}
	if ctx.Err != nil {
		switch ctx.Err.(type) {
		case ErrorRevert, ErrorExecLimitExceeded:
			txErr = ctx.Err
		case ErrorSelfDestruct:
			txErr = nil
		default:
			err = ctx.Err
		}
	}
	return
}

func (vm *VM) ValidateABI(ctx *Context, abi fat104.ABI) error {
	for name, f := range abi {
		args := dummyArgs(f)
		if args == nil {
			return fmt.Errorf("invalid ABI: args for %q", name)
		}

		ret, _, err := vm.Call(ctx, &f, args...)
		if err != nil {
			return fmt.Errorf("invalid ABI: %w", err)
		}

		switch ret.GetType() {
		case wasmer.TypeVoid:
			if f.Ret != fat104.TypeUndefined {
				return fmt.Errorf("invalid ABI: return type for %q",
					name)
			}
		case wasmer.TypeI32:
			if f.Ret != fat104.TypeI32 {
				return fmt.Errorf("invalid ABI: return type for %q",
					name)
			}
		case wasmer.TypeI64:
			if f.Ret != fat104.TypeI64 {
				return fmt.Errorf("invalid ABI: return type for %q",
					name)
			}
		default:
			return fmt.Errorf("invalid ABI: return type for %q", name)
		}

	}
	return nil
}

func dummyArgs(f fat104.Func) []json.RawMessage {
	args := make([]json.RawMessage, 0, len(f.Args))
	for _, arg := range f.Args {
		a := dummyArgsType(arg)
		if a == nil {
			return nil
		}
		args = append(args, a)
	}
	return args
}

func dummyArgsType(t fat104.Type) json.RawMessage {
	switch t {
	case fat104.TypeI32, fat104.TypeI64:
		return json.RawMessage(`0`)
	case fat104.TypeString:
		return json.RawMessage(`"test"`)
	case fat104.TypeBytes:
		return json.RawMessage(`"deadbeef"`)
	default:
		return nil
	}
}
