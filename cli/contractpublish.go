package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/AdamSLevy/jsonrpc2/v14"
	"github.com/Factom-Asset-Tokens/factom"
	"github.com/Factom-Asset-Tokens/factom/fat104"
	"github.com/Factom-Asset-Tokens/factom/fat107"
	"github.com/Factom-Asset-Tokens/fatd/internal/runtime"
	"github.com/posener/complete"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/wasmerio/go-ext-wasm/wasmer"
)

var contractPublishCmd = func() *cobra.Command {
	var pub ContractPublisher
	cmd := &cobra.Command{
		DisableFlagsInUseLine: true,
		Use: `
publish --wasm <binary-wasm-file> --abi <abi-json>
        [--metadata <metadata-json>]
`[1:],
		Aliases: []string{"pub"},
		Short:   "Publish a WASM smart contract",
		Long: `
Publish the WASM binary code and its public ABI.
`[1:],
		Args: cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			return pub.ValidateFlagStructure(cmd.Flags())
		},
		RunE: func(*cobra.Command, []string) error { return pub.Publish() },
	}
	contractCmd.AddCommand(cmd)
	contractCmplCmd.Sub["publish"] = contractPublishCmplCmd
	rootCmplCmd.Sub["help"].Sub["contract"].Sub["publish"] = complete.Command{}

	cmd.Flags().AddFlagSet(pub.Flags())

	generateCmplFlags(cmd, contractPublishCmplCmd.Flags)
	// Don't complete these global flags as they are ignored by this
	// command.
	for _, flg := range []string{"-C", "--chainid"} {
		delete(contractPublishCmplCmd.Flags, flg)
	}
	usage := cmd.UsageFunc()
	cmd.SetUsageFunc(func(cmd *cobra.Command) error {
		cmd.Flags().MarkHidden("chainid")
		return usage(cmd)
	})
	return cmd
}()

var contractPublishCmplCmd = complete.Command{
	Flags: mergeFlags(apiCmplFlags, ecAdrCmplFlags,
		complete.Flags{
			"--api":      complete.PredictFiles("*.json"),
			"-a":         complete.PredictFiles("*.json"),
			"--metadata": complete.PredictFiles("*.json"),
			"-m":         complete.PredictFiles("*.json"),
			"--wasm":     complete.PredictFiles("*.wasm"),
		}),
	Args: complete.PredictAnything,
}

type ContractPublisher struct {
	fat104.Contract
	Wasm BinaryFile
	EntryCreator
}

func (pub *ContractPublisher) Flags() *flag.FlagSet {
	flags := flag.NewFlagSet("Contract Publication", flag.ContinueOnError)
	flags.VarPF((*ABI)(&pub.ABI), "abi", "a", "Contract ABI")
	flags.VarPF((*JSONOrFile)(&pub.Metadata), "metadata", "m", "Contract Metadata")
	flags.VarPF(&pub.Wasm, "wasm", "W", "Contract Wasm Binary File")
	flags.AddFlagSet(pub.EntryCreator.Flags())
	return flags
}

func (pub *ContractPublisher) ValidateFlagStructure(flags *flag.FlagSet) error {
	if err := required(flags, "abi", "wasm"); err != nil {
		return err
	}

	if err := pub.EntryCreator.ValidateFlagStructure(flags); err != nil {
		return err
	}

	return nil
}

func (pub *ContractPublisher) Publish() error {
	vrbLog.Println("Compiling WASM...")
	mod, err := wasmer.CompileWithGasMetering(pub.Wasm.Data)
	if err != nil {
		return fmt.Errorf("--wasm %q: %w", pub.Wasm, err)
	}

	vrbLog.Println("Linking WASM...")
	vm, err := runtime.NewVM(&mod)
	if err != nil {
		return fmt.Errorf("--wasm %q: %w", pub.Wasm, err)
	}

	// Currently we cannot easily validate ABIs with the current design of
	// the runtime. Improvements in the future could allow this.
	// TODO: Improve ABI validation in runtime.
	errLog.Println("Warning: Full ABI validation checks are not implemented!")
	if len(vm.Exports) != len(pub.ABI) {
		return fmt.Errorf("ABI and Wasm do not match")
	}
	for fname := range vm.Exports {
		_, ok := pub.ABI[fname]
		if !ok {
			return fmt.Errorf("ABI and Wasm do not match")
		}
	}

	vrbLog.Println("Compressing WASM...")
	data := pub.Wasm.Data
	size := uint64(len(data))
	dataHash := factom.Bytes32(sha256.Sum256(data))
	dataHash = sha256.Sum256(dataHash[:])

	dataBuf := bytes.NewBuffer(data)
	cDataBuf := bytes.NewBuffer(make([]byte, 0, len(data)))

	gz := gzip.NewWriter(cDataBuf)
	_, err = dataBuf.WriteTo(gz)
	if err != nil {
		return err
	}
	err = gz.Close()
	if err != nil {
		return err
	}

	vrbLog.Println("Generating Contract Chain...")
	compression := fat107.Compression{Format: "gzip", Size: uint64(cDataBuf.Len())}
	chainID, _, hashes, commits, reveals, totalCost, err := fat107.Generate(
		context.Background(), pub.Es,
		cDataBuf, &compression, size, &dataHash,
		pub.Metadata)
	if err != nil {
		return err
	}

	if err := pub.CheckECBalance(totalCost); err != nil {
		return err
	}

	if !pub.Force {
		vrbLog.Println("Checking if contract chain already exists...")
		eb := factom.EBlock{ChainID: &chainID}
		inProcessList, err := eb.GetChainHead(context.Background(), FactomClient)
		if err != nil {
			if err, ok := err.(jsonrpc2.Error); ok {
				if err != (jsonrpc2.Error{
					Code: -32009, Message: "Missing Chain Head"}) {
					return err
				}
			} else {
				return err
			}
		}
		// TODO: make this more forgiving so that contract publication
		// can be resumed if interrupted.
		if inProcessList || eb.KeyMR != nil {
			return fmt.Errorf("Chain already exists.")
		}
	}

	if !pub.Curl {
		vrbLog.Printf("Publishing Contract Chain %v...", chainID)
	}
	for i, commit := range commits {
		if err := pub.Create(commit, reveals[i], &hashes[i]); err != nil {
			return err
		}
	}
	if !pub.Curl {
		fmt.Println("Contract Chain:", chainID)
	}

	return nil
}
