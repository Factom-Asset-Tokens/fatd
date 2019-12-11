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
