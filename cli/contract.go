package main

import (
	"github.com/posener/complete"
	"github.com/spf13/cobra"
)

// contractCmd represents the get command
var contractCmd = func() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "contracts",
		Aliases: []string{"contract", "con"},
		Short:   "publish|lookup|delegate|call",
		Long: `
Publish, lookup, delegate control of an address to, or call a FAT Smart
Contract.

Publish
        Smart contracts are just wasm code compiled for the FAT-SC runtime
        environment that have been published on a data store chain.

Lookup
        Any FAT-0 chain can reference any published smart contract by its
        ChainID. You can view the metadata or download the contract bytecode
        for a given Chain ID if it is a valid FAT-SC data store chain.

Delegate
        On a given FAT-0 chain, you can delegate control of an address to a
        published contract. This instantiates the contract code on this
        address.

Call
        Once control of an address has been delegated to a contract, that
        contract's public functions can be called by any other address on that
        FAT-0 chain.
`[1:],
	}
	rootCmd.AddCommand(cmd)
	rootCmplCmd.Sub["contract"] = contractCmplCmd
	rootCmplCmd.Sub["help"].Sub["contract"] =
		complete.Command{Sub: complete.Commands{}}
	generateCmplFlags(cmd, contractCmplCmd.Flags)
	return cmd
}()

var contractCmplCmd = complete.Command{
	Flags: apiCmplFlags,
	Sub:   complete.Commands{},
}
