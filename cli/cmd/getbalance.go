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
	"os"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/srv"
	"github.com/posener/complete"
	"github.com/spf13/cobra"
)

var addresses []factom.FAAddress

// getBalanceCmd represents the balance command
var getBalanceCmd = func() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "balance ADDRESS...",
		Aliases:               []string{"balances"},
		DisableFlagsInUseLine: true,
		Short:                 "Get balances for addresses",
		Long: `Get the balance of each ADDRESS.

The balance of each ADDRESS on the given --chainid (or --tokenid and
--identity) is returned.`,
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
	params := srv.ParamsGetBalance{}
	params.ChainID = paramsToken.ChainID
	balances := make([]uint64, len(addresses))
	for i, adr := range addresses {
		params.Address = &adr
		if err := FATClient.Request("get-balance", params, &balances[i]); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
	for i, adr := range addresses {
		fmt.Println(adr, balances[i])
	}
}
