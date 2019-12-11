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

	"github.com/Factom-Asset-Tokens/factom"
	"github.com/Factom-Asset-Tokens/factom/fat0"
	"github.com/Factom-Asset-Tokens/fatd/fat"
	"github.com/Factom-Asset-Tokens/fatd/internal/db"
	"github.com/Factom-Asset-Tokens/fatd/internal/db/addresses"
	"github.com/wasmerio/go-ext-wasm/wasmer"
)

type Context struct {
	db.Chain
	factom.DBlock
	fat0.Transaction

	ctx context.Context
}

func intoContext(ctx *wasmer.InstanceContext) *Context {
	return ctx.Data().(*Context)
}

func (ctx *Context) Sender() factom.FAAddress {
	for sender := range ctx.Transaction.Inputs {
		return sender
	}
	panic("empty Transaction.Inputs!")
}

func (ctx *Context) ContractAddress() factom.FAAddress {
	for contract := range ctx.Transaction.Outputs {
		return contract
	}
	panic("empty Transaction.Outputs!")
}

func (ctx *Context) Amount() uint64 {
	for _, amount := range ctx.Transaction.Outputs {
		return amount
	}
	panic("empty Transaction.Outputs!")
}

func (ctx *Context) Send(amount uint64, adr *factom.FAAddress) {
	chain := ctx.Chain
	contract := ctx.ContractAddress()
	if contract == fat.Coinbase() {
		if chain.Issuance.Supply > 0 &&
			int64(chain.NumIssued+amount) > chain.Issuance.Supply {
			panic(fmt.Errorf("coinbase exceeds max supply"))
		}
		if err := chain.AddNumIssued(amount); err != nil {
			panic(err)
		}
	} else {
		_, txErr, err := addresses.Sub(chain.Conn, &contract, amount)
		if err != nil {
			panic(err)
		}
		if txErr != nil {
			panic(txErr)
		}
	}

	_, err := addresses.Add(chain.Conn, adr, amount)
	if err != nil {
		panic(err)
	}
}
