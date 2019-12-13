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
	"context"
	"fmt"
	"math"
	"unsafe"

	"github.com/Factom-Asset-Tokens/factom"
	"github.com/Factom-Asset-Tokens/factom/fat0"
	"github.com/Factom-Asset-Tokens/fatd/fat"
	"github.com/Factom-Asset-Tokens/fatd/internal/db"
	"github.com/Factom-Asset-Tokens/fatd/internal/db/addresses"
	"github.com/Factom-Asset-Tokens/fatd/internal/db/contracts"
	"github.com/wasmerio/go-ext-wasm/wasmer"
)

type Context struct {
	db.Chain
	factom.DBlock
	fat0.Transaction

	Err error

	context.Context
	*wasmer.InstanceContext
}

func intoContext(ptr unsafe.Pointer) *Context {
	instanceCtx := wasmer.IntoInstanceContext(ptr)
	ctx := instanceCtx.Data().(*Context)
	ctx.InstanceContext = &instanceCtx
	return ctx
}

func (ctx *Context) Sender() factom.FAAddress {
	for sender := range ctx.Transaction.Inputs {
		return sender
	}
	panic("empty Transaction.Inputs!")
}

func (ctx *Context) ContractAddress() (factom.FAAddress, error) {
	for contract := range ctx.Transaction.Outputs {
		return contract, nil
	}
	return factom.FAAddress{}, ctx.Error(fmt.Errorf("empty Transaction.Outputs!"))
}

func (ctx *Context) Amount() (uint64, error) {
	for _, amount := range ctx.Transaction.Outputs {
		return amount, nil
	}
	return 0, ctx.Error(fmt.Errorf("empty Transaction.Outputs!"))
}
func (ctx *Context) ContractBalance() (uint64, error) {
	contract, err := ctx.ContractAddress()
	if err != nil {
		return 0, err
	}
	if contract == fat.Coinbase() {
		if ctx.Chain.Issuance.Supply < 0 {
			return math.MaxUint64, nil
		}
		return uint64(ctx.Chain.Issuance.Supply) - ctx.Chain.NumIssued, nil
	}
	_, bal, err := addresses.SelectIDBalance(ctx.Chain.Conn, &contract)
	if err != nil {
		return 0, ctx.Error(fmt.Errorf(
			"get_balance: addresses.SelectIDBalance: %w", err))
	}
	return bal, nil
}

func (ctx *Context) Send(amount uint64, adr *factom.FAAddress) error {
	contract, err := ctx.ContractAddress()
	if err != nil {
		return err
	}
	if contract == fat.Coinbase() {
		if ctx.Chain.Issuance.Supply > 0 &&
			int64(ctx.Chain.NumIssued+amount) > ctx.Chain.Issuance.Supply {
			return ctx.Revert("send: max supply exceeded")
		}
		if err := ctx.Chain.AddNumIssued(amount); err != nil {
			return ctx.Error(err)
		}
	} else {
		_, txErr, err := addresses.Sub(ctx.Chain.Conn, &contract, amount)
		if err != nil {
			return ctx.Error(fmt.Errorf("addresses.Sub: %w", err))
		}
		if txErr != nil {
			return ctx.Revert("send: insufficient balance")
		}
	}

	_, err = addresses.Add(ctx.Chain.Conn, adr, amount)
	if err != nil {
		return ctx.Error(fmt.Errorf("addresses.Add: %w", err))
	}
	return nil
}

func (ctx *Context) SelfDestruct() error {
	ctx.ConsumeAllGas()
	adr, err := ctx.ContractAddress()
	if err != nil {
		return err
	}
	var id int64
	id, ctx.Err = addresses.SelectID(ctx.Chain.Conn, &adr)
	if ctx.Err != nil {
		return ctx.Err
	}
	ctx.Err = contracts.DeleteAddressContract(ctx.Chain.Conn, id)
	if ctx.Err != nil {
		return ctx.Err
	}
	// Marks a successful self destruct. Will not be set as a Tx Err.
	ctx.Err = ErrorSelfDestruct{}
	return ctx.Err
}

func (ctx *Context) Revert(reason string) error {
	ctx.ConsumeAllGas()
	ctx.Err = ErrorRevert{reason}
	return ctx.Err
}

func (ctx *Context) ConsumeAllGas() {
	ctx.SetPointsUsed(ctx.GetExecLimit())
}

func (ctx *Context) Error(err error) error {
	if err != nil {
		ctx.ConsumeAllGas()
	}
	ctx.Err = err
	return err
}

func (ctx *Context) ReadAddress(adr_buf int32) (factom.FAAddress, error) {
	var adr factom.FAAddress
	if 32 != copy(adr[:], ctx.Memory().Data()[adr_buf:]) {
		return adr, ctx.Error(
			fmt.Errorf("Context.ReadAddress: invalid copy length"))
	}
	return adr, nil
}

func (ctx *Context) ReadString(str_buf int32, size uint32) string {
	str := make([]byte, size)
	n := copy(str[:], ctx.Memory().Data()[str_buf:])
	return string(str[:n])
}

func (ctx *Context) WriteAddress(adr *factom.FAAddress, adr_buf int32) error {
	if 32 != copy(ctx.Memory().Data()[adr_buf:], adr[:]) {
		return ctx.Error(
			fmt.Errorf("Context.WriteAddress: invalid copy length"))
	}
	return nil
}
