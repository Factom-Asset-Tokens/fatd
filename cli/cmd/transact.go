// MIT License
//
// Copyright 2018 Canonical Ledgers, LLC
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

package cmd

import (
	"encoding/json"
	"fmt"
	"math"

	"github.com/Factom-Asset-Tokens/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat"
	"github.com/Factom-Asset-Tokens/fatd/fat/fat0"
	"github.com/Factom-Asset-Tokens/fatd/fat/fat1"
	"github.com/Factom-Asset-Tokens/fatd/srv"

	jrpc "github.com/AdamSLevy/jsonrpc2/v11"
	"github.com/posener/complete"
	"github.com/spf13/cobra"
)

var (
	metadata json.RawMessage
)

// transactCmd represents the transact command
var transactCmd = func() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "transact",
		Aliases: []string{"send", "distribute"},
		Short:   "Send or distribute FAT tokens",
		Long: `
Send or distribute FAT-0 or FAT-1 tokens.

Submitting a FAT transaction involves submitting a signed transaction entry to
the given FAT Token Chain. The flags for 'transact fat0' and 'transact fat1'
are the same except for the arguments to the --input and --output flags differ
slightly.

Inputs and Outputs
        Both --input and --output may be used multiple times. For both flags,
        the argument must be either a public (FA) or private (Fs) Factoid
        address, followed by a ":" and some specifier defined by the subcommand
        for the token type.

        For FAT-0, the argument to --input or --output could be,
                FA3SjebEevRe964p4tQ6eieEvzi7puv9JWF3S3Wgw2v3WGKueL3R:150
                Fs2mGpZiHMwiEfe7kBD5ZYpXJsaxb3gUX258PJsAcNJ8GxFy8pBt:150

        For FAT-1, the argument to --input or --output could be,
                FA3SjebEevRe964p4tQ6eieEvzi7puv9JWF3S3Wgw2v3WGKueL3R:[1,2,5-100]
                Fs2mGpZiHMwiEfe7kBD5ZYpXJsaxb3gUX258PJsAcNJ8GxFy8pBt:[1,2,5-100]

        The private keys for all --input addresses must be either known to
        factom-walletd or directly supplied.

        The keyword "coinbase" may be used in place of an address to specify
        the coinbase address in an --output.

        See 'fat-cli transact fat0 --help' or 'fat-cli transact fat1 --help'
        for more information about the --input and --output argument format.

Normal Transactions
        Normal transactions are multi --input and multi --output between
        virtually any number of Factoid addresses. These are generated and
        signed by the users that control the private keys for the input
        addresses in the transaction.

        For normal transactions, supply at least one --input, and at least one
        --output. The coinbase address may not be an an --input, but may be an
        --output. Tokens sent to the coinbase address are provably burned.

        The --output addresses must not include any address used as an --input.

        Every token sent as an --input must also be part of some --output.

Coinbase Transactions
        Coinbase transactions distribute new tokens to user addresses. These
        are multi --output transactions from the coinbase address, and are
        generated and signed by the Issuer of the token, who controls the --sk1
        key. The coinbase address may not be used as an --output.

        Coinbase transactions that distribute an amount that causes the total
        supply to exceed the max supply declared in the Token Initialization
        Entry, are invalid and ignored. Burned tokens are still counted towards
        the total supply.

Entry Credits
        Creating entries on the Factom blockchain costs Entry Credits. Most
        transactions with minimal metadata costs only 1 EC. You must specify a
        funded Entry Credit address with --ecadr, which may be either a private
        Es address, or a pubilc EC address that can be fetched from
        factom-walletd.

Sanity Checks
        Transactions are always sanity checked for valid data and signatures
        locally.

        Additionally, prior to composing the Transaction Entry, a number of
        calls to fatd and factomd are made to ensure that transaction will be
        considered valid by fatd. These network checks are skipped if --force
        is used.

        - The Token Chain has been issued as the correct FAT type.
        - All inputs have sufficient balance.
        - For coinbase transactions, the --sk1 key corresponds to  the Identity
          Chain's declared ID1 key.
`[1:],
		PersistentPreRunE: validateTransactFlags,
	}
	rootCmd.AddCommand(cmd)
	rootCmplCmd.Sub["transact"] = transactCmplCmd
	rootCmplCmd.Sub["help"].Sub["transact"] = complete.Command{Sub: complete.Commands{}}

	flags := cmd.PersistentFlags()
	flags.AddFlagSet(composeFlags)
	flags.VarPF(&sk1, "sk1", "",
		"Secret Identity Key 1 to sign coinbase txs").DefValue = ""
	flags.VarPF((*RawMessage)(&metadata), "metadata", "m",
		"JSON metadata to include in tx")

	generateCmplFlags(cmd, transactCmplCmd.Flags)
	return cmd
}()

var transactCmplCmd = complete.Command{
	Flags: mergeFlags(apiCmplFlags, tokenCmplFlags, ecAdrCmplFlags),
	Sub:   complete.Commands{},
}

var signingSet []factom.RCDPrivateKey

func validateTransactFlags(cmd *cobra.Command, args []string) error {
	if err := validateChainIDFlags(cmd, args); err != nil {
		return err
	}
	var cmdType fat.Type
	if err := (*Type)(&cmdType).Set(cmd.Name()); err != nil {
		panic(err) // This should never happen.
	}

	if err := validateECAdrFlag(cmd, args); err != nil {
		return err
	}

	flags := cmd.Flags()
	if !flags.Changed("output") {
		return fmt.Errorf("at least one --output is required")
	}
	inputSet := flags.Changed("input")
	sk1Set := flags.Changed("sk1")
	if !inputSet && !sk1Set {
		return fmt.Errorf("--sk1 or at least one --input is required")
	}
	if inputSet && sk1Set {
		return fmt.Errorf("--sk1 and --input may not be used at the same time")
	}

	// All subsequent errors are not issues with correct use of flags, so
	// avoid printing Usage() by calling os.Fata() instead of returning.

	// Populate all private keys
	var numInputs int = 1
	var inputAdrs []factom.FAAddress
	if inputSet {
		numInputs = len(fat0Tx.Inputs) + len(fat1Tx.Inputs)
		inputAdrs = make([]factom.FAAddress, 0, numInputs)
		switch cmdType {
		case fat0.Type:
			for fa := range fat0Tx.Inputs {
				inputAdrs = append(inputAdrs, fa)
			}
		case fat1.Type:
			for fa := range fat1Tx.Inputs {
				inputAdrs = append(inputAdrs, fa)
			}
		}
		signingSet = make([]factom.RCDPrivateKey, numInputs)
		for i, fa := range inputAdrs {
			fs, ok := privateAddress[fa]
			if !ok {
				var err error
				vrbLog.Println("Fetching secret address...", fa)
				fs, err = fa.GetFsAddress(FactomClient)
				if err != nil {
					if err, ok := err.(jrpc.Error); ok {
						errLog.Fatal(err.Data, fa)
					}
					errLog.Fatal(err)
				}
			}
			signingSet[i] = fs
		}
	} else {
		signingSet = append(signingSet, sk1)
		switch cmdType {
		case fat0.Type:
			fat0Tx.Inputs = make(fat0.AddressAmountMap, 1)
			fat0Tx.Inputs[fat.Coinbase()] = fat0Tx.Outputs.Sum()
		case fat1.Type:
			fat1Tx.Inputs = make(fat1.AddressNFTokensMap, 1)
			fat1Tx.Inputs[fat.Coinbase()] = fat1Tx.Outputs.AllNFTokens()
		}
	}

	vrbLog.Printf("Preparing %v Transaction Entry...", cmdType)
	var tx interface {
		Sign(...factom.RCDPrivateKey)
		MarshalEntry() error
		Cost() (uint8, error)
	}
	switch cmdType {
	case fat0.Type:
		fat0Tx.ChainID = paramsToken.ChainID
		fat0Tx.Metadata = metadata
		tx = &fat0Tx
	case fat1.Type:
		fat1Tx.ChainID = paramsToken.ChainID
		fat1Tx.Metadata = metadata
		tx = &fat1Tx
	}
	if err := tx.MarshalEntry(); err != nil {
		errLog.Fatal(err)
	}
	vrbLog.Println("Transaction Entry Content: ", tx)
	tx.Sign(signingSet...)
	cost, err := tx.Cost()
	if err != nil {
		errLog.Fatal(err)
	}

	if !force {
		vrbLog.Println("Checking token chain status...")
		params := srv.ParamsToken{ChainID: paramsToken.ChainID}
		var stats srv.ResultGetStats
		if err := FATClient.Request("get-stats", params, &stats); err != nil {
			errLog.Fatal(err)
		}
		// Verify we are using the right command for this token type.
		if cmdType != stats.Issuance.Type {
			errLog.Fatalf("incorrect token type: expected %v, but chain is %v",
				cmdType, stats.Issuance.Type)
		}

		if inputSet {
			paramsGetBalance := srv.ParamsGetBalance{ParamsToken: params}
			for _, adr := range inputAdrs {
				vrbLog.Println("Checking FAT Token balance...", adr)
				paramsGetBalance.Address = &adr
				var balance uint64
				if err := FATClient.Request("get-balance",
					paramsGetBalance, &balance); err != nil {
					errLog.Fatal(err)
				}
				var inputAmount uint64
				switch cmdType {
				case fat0.Type:
					inputAmount = fat0Tx.Inputs[adr]
				case fat1.Type:
					inputAmount = uint64(len(fat1Tx.Inputs[adr]))
				}
				if inputAmount > balance {
					errLog.Fatalf(
						"--input %v:%v has insufficient balance (%v)",
						adr, addressValueStrMap[adr], balance)
				}
			}
		}
		if inputSet && cmdType == fat1.Type {
			params := srv.ParamsGetNFBalance{ParamsToken: params}
			params.Limit = math.MaxUint64
			for _, adr := range inputAdrs {
				vrbLog.Println("Checking FAT NF Token ownership...", adr)
				params.Address = &adr
				var balance fat1.NFTokens
				if err := FATClient.Request("get-nf-balance",
					params, &balance); err != nil {
					errLog.Fatal(err)
				}
				if err := balance.ContainsAll(fat1Tx.Inputs[adr]); err != nil {
					tknID := fat1.NFTokenID(
						err.(fat1.ErrorMissingNFTokenID))
					errLog.Fatalf(
						"--input %v:%v does not own NFTokenID %v",
						adr, addressValueStrMap[adr], tknID)
				}
			}
		}

		// Validate coinbase transaction
		if sk1Set {
			verifySK1Key(&sk1, stats.IssuerChainID)

			vrbLog.Println("Validating coinbase transaction...")
			var issuing uint64
			switch cmdType {
			case fat0.Type:
				issuing = fat0Tx.Inputs.Sum()
			case fat1.Type:
				issuing = uint64(len(fat1Tx.Inputs.AllNFTokens()))
			}
			issued := stats.CirculatingSupply + stats.Burned
			if stats.Issuance.Supply != -1 &&
				issuing+issued > uint64(stats.Issuance.Supply) {
				errLog.Fatal(
					"invalid coinbase transaction: exceeds max supply")
			}
			if cmdType == fat1.Type {
				params := srv.ParamsGetNFToken{ParamsToken: params}
				for tknID := range fat1Tx.Inputs.AllNFTokens() {
					params.NFTokenID = &tknID
					err := FATClient.Request("get-nf-token", params, nil)
					if err == nil {
						errLog.Fatalf("invalid coinbase transaction: NFTokenID (%v) already exists",
							tknID)
					}
					rpcErr, _ := err.(jrpc.Error)
					if rpcErr.Code != srv.ErrorTokenNotFound.Code {
						errLog.Fatal(err)
					}
				}
			}
		}

		verifyECBalance(&ecEsAdr.EC, cost)
		vrbLog.Printf("Transaction Entry Cost: %v EC", cost)
		vrbLog.Println()
	}

	var entry factom.Entry
	switch cmdType {
	case fat0.Type:
		entry = fat0Tx.Entry.Entry
	case fat1.Type:
		entry = fat1Tx.Entry.Entry
	}
	entry.ChainID = paramsToken.ChainID
	if curl {
		if err := printCurl(entry, ecEsAdr.Es); err != nil {
			errLog.Fatal(err)
		}
		return nil
	}

	vrbLog.Printf("Submitting the %v Transaction Entry to the Factom blockchain...",
		cmdType)
	txID, err := entry.ComposeCreate(FactomClient, ecEsAdr.Es)
	if err != nil {
		errLog.Fatal(err)
	}
	fmt.Printf("%v Transaction Entry Created: %v\n", cmdType, entry.Hash)
	fmt.Printf("Chain ID: %v\n", entry.ChainID)
	fmt.Printf("Factom Tx ID: %v\n", txID)
	return nil
}
