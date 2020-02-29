package main

import (
	"github.com/Factom-Asset-Tokens/factom"
	"github.com/Factom-Asset-Tokens/factom/fat0"
	"github.com/posener/complete"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
)

var contractDelegateCmd = func() *cobra.Command {
	dgt := NewContractDelegator()
	cmd := &cobra.Command{
		DisableFlagsInUseLine: true,
		Use: `
delegate --ecadr <EC | Es> --chainid <chain-id>
        --faadr <FA | Fs> --contract <contract-chain-id>
        [--metadata <metadata-json>]
`[1:],
		Short: "Delegate control of an address to a smart contract",
		Long: `
Delegate control of an FAAddress on a single FAT chain to a smart contract.
After successfully delegating control of an address, normal transactions from
this address will not be possible on that chain. Only valid contract calls to
this address are allowed.
`[1:],
		Args: cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := validateChainIDFlags(cmd, args); err != nil {
				return err
			}
			return dgt.ValidateFlagStructure(cmd.Flags())
		},
		RunE: func(*cobra.Command, []string) error { return dgt.RunE() },
	}
	contractCmd.AddCommand(cmd)
	contractCmplCmd.Sub["delegate"] = contractDelegatorCmplCmd
	rootCmplCmd.Sub["help"].Sub["contract"].Sub["delegate"] = complete.Command{}

	cmd.Flags().AddFlagSet(dgt.Flags())

	generateCmplFlags(cmd, contractPublishCmplCmd.Flags)
	return cmd
}()

var contractDelegatorCmplCmd = complete.Command{
	Flags: mergeFlags(apiCmplFlags, ecAdrCmplFlags,
		complete.Flags{
			"--faadr":    PredictFAAddresses,
			"--contract": complete.PredictAnything,
			"--metadata": complete.PredictFiles("*.json"),
			"-m":         complete.PredictFiles("*.json"),
		}),
	Args: complete.PredictAnything,
}

type ContractDelegator struct {
	EntryCreator
	FAFsAddress
	fat0.Transaction
}

func NewContractDelegator() ContractDelegator {
	var dgt ContractDelegator
	dgt.Contract = new(factom.Bytes32)
	return dgt
}

func (dgt *ContractDelegator) Flags() *flag.FlagSet {
	flags := flag.NewFlagSet("Contract Delegation", flag.ContinueOnError)
	flags.VarPF(&dgt.FAFsAddress, "faadr", "",
		"FA or Fs Address to delegate control of").DefValue = ""
	flags.VarPF(dgt.Contract, "contract", "",
		"Contract Chain ID to delegate control to").DefValue = ""
	flags.VarPF((*JSONOrFile)(&dgt.Metadata), "metadata", "m", "Contract Metadata")
	flags.AddFlagSet(dgt.EntryCreator.Flags())
	return flags
}

func (dgt *ContractDelegator) ValidateFlagStructure(flags *flag.FlagSet) error {
	if err := required(flags, "chainid", "contract", "faadr"); err != nil {
		return err
	}

	if err := dgt.EntryCreator.ValidateFlagStructure(flags); err != nil {
		return err
	}

	return nil
}

func (dgt *ContractDelegator) RunE() error {
	if err := dgt.PopulateEsAddress(); err != nil {
		return err
	}
	if err := dgt.PopulateFsAddress(); err != nil {
		return err
	}

	tx := dgt.Transaction
	tx.Entry.ChainID = paramsToken.ChainID
	tx.Inputs = fat0.AddressAmountMap{dgt.FAFsAddress.FA: 0}
	tx.Outputs = tx.Inputs
	e, err := tx.Sign(dgt.Fs)
	if err != nil {
		return err
	}

	// TODO: add full validation checks
	// - FAT Chain is FAT-0
	// - Contract Chain is valid          (needs fatd api)
	// - Address is not already delegated (needs fatd api)

	if !dgt.Curl {
		vrbLog.Printf(
			"Delegating control of %v to contract %v on FAT chain %v...",
			dgt.FA, dgt.Contract, paramsToken.ChainID)
	}
	factomTxID, err := dgt.ComposeCreate(&e)
	if err != nil {
		return err
	}

	_ = factomTxID

	return nil
}
