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

// chainsCmd represents the chains command
var chainsCmd = &cobra.Command{
	Use:                   "chains [CHAINID...]",
	Aliases:               []string{"chain", "stats", "stat"},
	DisableFlagsInUseLine: true,
	Short:                 "Get information about Token Chains",
	Long: `Get information about each CHAINID.

chains returns Token ID and Issuer Identity Chain ID for each CHAINID.

If no CHAINID is given, then chains returns a list of the issued Token Chain
IDs that fatd is tracking.`,
	Args: getChainsArgs,
	Run:  getChains,
}

var chainsCmplCmd = complete.Command{
	Flags: apiFlags,
	Args:  PredictChainIDs,
}

func init() {
	getCmd.AddCommand(chainsCmd)
	getCmplCmd.Sub["chains"] = chainsCmplCmd
}

var chainIDs []factom.Bytes32

func getChainsArgs(_ *cobra.Command, args []string) error {
	chainIDs = make([]factom.Bytes32, len(args))
	dupl := make(map[factom.Bytes32]struct{}, len(args))
	for i, arg := range args {
		id := &chainIDs[i]
		if err := id.Set(arg); err != nil {
			return err
		}
		if _, ok := dupl[*id]; ok {
			return fmt.Errorf("duplicate: %v", id)
		}
		dupl[*id] = struct{}{}
	}
	return nil
}

func getChains(_ *cobra.Command, _ []string) {
	if len(chainIDs) == 0 {
		var chains []srv.ParamsToken
		if err := FATClient.Request("get-daemon-tokens", nil, &chains); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		for _, chain := range chains {
			fmt.Printf(`Chain ID: %v
Issuer Identity Chain ID: %v
Token ID: %q

`,
				chain.ChainID, chain.TokenID, chain.IssuerChainID)
		}
	}

	for _, chainID := range chainIDs {
		chain := srv.ParamsToken{ChainID: &chainID}
		var stats srv.ResultGetStats
		if err := FATClient.Request("get-stats", chain, &stats); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Printf(`Chain ID: %v
Issuer Identity Chain ID: %v
Token ID: %v
Type: %v
Symbol: %q
Supply:            %v
Ciculating Supply: %v
Burned:            %v
Number of Transactions: %v
Issuance Timestamp: %v
`,
			chainID, stats.TokenID, stats.IssuerChainID,
			stats.Issuance.Type, stats.Issuance.Symbol,
			stats.Issuance.Supply, stats.CirculatingSupply, stats.Burned,
			stats.Transactions,
			stats.IssuanceTimestamp.Time())
		if stats.LastTransactionTimestamp != nil {
			fmt.Printf("Last Tx Timestamp: %v\n",
				stats.LastTransactionTimestamp.Time())
		}
		fmt.Println("")

	}

}
