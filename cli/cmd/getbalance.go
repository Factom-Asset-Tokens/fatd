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
	"fmt"
	"math"

	"github.com/Factom-Asset-Tokens/fatd/factom"
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
balance --chainid <chain-id> ADDRESS...`[1:],
		Aliases: []string{"balances"},
		Short:   "Get balances for addresses",
		Long: `
Get the balance of each ADDRESS on the given --chainid.

Returns the total balance for FAT-0 tokens or a list of NF Token IDs for FAT-1
tokens.
`[1:],
		Args:    getBalanceArgs,
		PreRunE: validateChainIDFlags,
		Run:     getBalance,
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
	vrbLog.Printf("Fetching token chain details... %v", paramsToken.ChainID)
	params := srv.ParamsToken{ChainID: paramsToken.ChainID}
	var stats srv.ResultGetStats
	if err := FATClient.Request("get-stats", params, &stats); err != nil {
		errLog.Fatal(err)
	}
	switch stats.Issuance.Type {
	case fat0.Type:
		params := srv.ParamsGetBalance{}
		params.ChainID = paramsToken.ChainID
		vrbLog.Println("Fetching balances...")
		for _, adr := range addresses {
			params.Address = &adr
			var balance uint64
			if err := FATClient.Request("get-balance", params,
				&balance); err != nil {
				errLog.Fatal(err)
			}
			fmt.Println(adr, balance)
		}
	case fat1.Type:
		limit := uint64(math.MaxUint64)
		params := srv.ParamsGetNFBalance{Limit: &limit}
		params.ChainID = paramsToken.ChainID
		vrbLog.Println("Fetching NF balances...")
		for _, adr := range addresses {
			params.Address = &adr
			var balance fat1.NFTokens
			if err := FATClient.Request("get-nf-balance", params,
				&balance); err != nil {
				errLog.Fatal(err)
			}
			fmt.Println(adr, balance)
		}
	}
}
