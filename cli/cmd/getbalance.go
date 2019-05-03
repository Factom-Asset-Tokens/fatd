// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
