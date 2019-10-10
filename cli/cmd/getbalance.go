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
	"context"
	"fmt"
	"math"

	"github.com/Factom-Asset-Tokens/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat/fat0"
	"github.com/Factom-Asset-Tokens/fatd/fat/fat1"
	"github.com/Factom-Asset-Tokens/fatd/srv"
	"github.com/posener/complete"
	"github.com/spf13/cobra"
)

var addresses []factom.FAAddress

// getBalanceCmd represents the balance command
var getBalanceCmd = func() *cobra.Command {
	cmd := &cobra.Command{
		DisableFlagsInUseLine: true,
		Use: `
balance [--chainid <chain-id>] ADDRESS...`[1:],
		Aliases: []string{"balances"},
		Short:   "Get balances for addresses",
		Long: `
Get the balance of each ADDRESS on the given --chainid, if given, otherwise
return all non-zero total balances.

The list of NF Token IDs for FAT-1 tokens are displayed if --chainid is used.
`[1:],
		Args: getBalanceArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			flags := cmd.Flags()
			chainIDSet := flags.Changed("chainid")
			tokenIDSet := flags.Changed("tokenid")
			identitySet := flags.Changed("identity")
			if chainIDSet || tokenIDSet || identitySet {
				return validateChainIDFlags(cmd, args)
			}
			paramsToken.ChainID = nil
			return nil
		},
		Run: getBalance,
	}
	getCmd.AddCommand(cmd)
	getCmplCmd.Sub["balance"] = getBalanceCmplCmd
	rootCmplCmd.Sub["help"].Sub["get"].Sub["balance"] = complete.Command{}
	generateCmplFlags(cmd, getBalanceCmplCmd.Flags)
	return cmd
}()

var getBalanceCmplCmd = complete.Command{
	Flags: mergeFlags(apiCmplFlags, tokenCmplFlags),
	Args:  PredictFAAddresses,
}

func getBalanceArgs(cmd *cobra.Command, args []string) error {
	if err := cobra.MinimumNArgs(1)(cmd, args); err != nil {
		return err
	}
	addresses = make([]factom.FAAddress, len(args))
	dupl := make(map[factom.FAAddress]struct{}, len(args))
	for i := range addresses {
		adr := &addresses[i]
		if err := adr.Set(args[i]); err != nil {
			return err
		}
		if _, ok := dupl[*adr]; ok {
			return fmt.Errorf("duplicate: %v", adr)
		}
		dupl[*adr] = struct{}{}
	}
	return nil
}

func getBalance(cmd *cobra.Command, _ []string) {
	if paramsToken.ChainID == nil {
		var params srv.ParamsGetBalances
		params.IncludePending = paramsToken.IncludePending
		vrbLog.Println("Fetching balances for all chains...")
		for _, adr := range addresses {
			params.Address = &adr
			var balances srv.ResultGetBalances
			if err := FATClient.Request(context.Background(),
				"get-balances", params, &balances); err != nil {
				errLog.Fatal(err)
			}
			fmt.Printf("%v:", adr)
			if len(balances) == 0 {
				fmt.Println(" none")
				continue
			}
			fmt.Println()
			for chainID, balance := range balances {
				vrbLog.Printf("Fetching token chain details... %v", chainID)
				params := srv.ParamsToken{ChainID: &chainID}
				var stats srv.ResultGetStats
				if err := FATClient.Request(context.Background(),
					"get-stats", params, &stats); err != nil {
					errLog.Fatal(err)
				}
				var bal interface{}
				if stats.Issuance.Precision > 1 {
					bal = float64(balance) / math.Pow10(
						int(stats.Issuance.Precision))
				} else {
					bal = balance
				}
				fmt.Printf("\t%v: %v\n", chainID, bal)
			}
		}
		return
	}

	vrbLog.Printf("Fetching token chain details... %v", paramsToken.ChainID)
	params := srv.ParamsToken{ChainID: paramsToken.ChainID}
	var stats srv.ResultGetStats
	if err := FATClient.Request(context.Background(),
		"get-stats", params, &stats); err != nil {
		errLog.Fatal(err)
	}
	switch stats.Issuance.Type {
	case fat0.Type:
		params := srv.ParamsGetBalance{}
		params.ChainID = paramsToken.ChainID
		params.IncludePending = paramsToken.IncludePending
		vrbLog.Println("Fetching balances...")
		for _, adr := range addresses {
			params.Address = &adr
			var balance uint64
			if err := FATClient.Request(context.Background(),
				"get-balance", params, &balance); err != nil {
				errLog.Fatal(err)
			}
			if stats.Issuance.Precision > 1 {
				fmt.Println(adr, float64(balance)/math.Pow10(
					int(stats.Issuance.Precision)))
			} else {
				fmt.Println(adr, balance)
			}
		}
	case fat1.Type:
		var params srv.ParamsGetNFBalance
		params.Limit = math.MaxUint64
		params.ChainID = paramsToken.ChainID
		params.IncludePending = paramsToken.IncludePending
		vrbLog.Println("Fetching NF balances...")
		for _, adr := range addresses {
			params.Address = &adr
			var balance fat1.NFTokens
			if err := FATClient.Request(context.Background(),
				"get-nf-balance", params, &balance); err != nil {
				errLog.Fatal(err)
			}
			fmt.Println(adr, balance)
		}
	}
}
