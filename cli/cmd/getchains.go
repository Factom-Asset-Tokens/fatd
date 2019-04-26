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

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/srv"
	"github.com/posener/complete"
	"github.com/spf13/cobra"
)

// getChainsCmd represents the chains command
var getChainsCmd = func() *cobra.Command {
	cmd := &cobra.Command{
		DisableFlagsInUseLine: true,
		Use: `
chains [CHAINID...]`[1:],
		Aliases: []string{"chain", "stats", "stat", "issuance", "issuances"},
		Short:   "List chains and their stats",
		Long: `
Get info about each CHAINID.

If at least one CHAINID is provided, then the stats and issuance info for each
chain is returned.

If no CHAINID is given, then the complete list of Issued Token Chains that fatd
is tracking is returned.
`[1:],
		Args: getChainsArgs,
		Run:  getChains,
	}
	getCmd.AddCommand(cmd)
	getCmplCmd.Sub["chains"] = getChainsCmplCmd
	rootCmplCmd.Sub["help"].Sub["get"].Sub["chains"] = complete.Command{}

	generateCmplFlags(cmd, getChainsCmplCmd.Flags)
	// Don't complete these global flags as they are ignored by this
	// command.
	for _, flg := range []string{"-C", "--chainid",
		"-I", "--identity", "-T", "--tokenid"} {
		delete(getChainsCmplCmd.Flags, flg)
	}
	usage := cmd.UsageFunc()
	cmd.SetUsageFunc(func(cmd *cobra.Command) error {
		cmd.Flags().MarkHidden("chainid")
		return usage(cmd)
	})
	return cmd
}()

var getChainsCmplCmd = complete.Command{
	Flags: mergeFlags(apiCmplFlags),
	Args:  PredictChainIDs,
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
		vrbLog.Println("Fetching list of issued token chains...")
		var chains []srv.ParamsToken
		if err := FATClient.Request("get-daemon-tokens", nil,
			&chains); err != nil {
			errLog.Fatal(err)
		}
		for _, chain := range chains {
			fmt.Printf(`Chain ID: %v
Issuer Identity Chain ID: %v
Token ID: %q

`,
				chain.ChainID, chain.IssuerChainID, chain.TokenID)
		}
	}

	for _, chainID := range chainIDs {
		vrbLog.Printf("Fetching token chain details... %v", chainID)
		params := srv.ParamsToken{ChainID: &chainID}
		var stats srv.ResultGetStats
		if err := FATClient.Request("get-stats", params, &stats); err != nil {
			errLog.Fatal(err)
		}
		printStats(&chainID, stats)
	}
}

func printStats(chainID *factom.Bytes32, stats srv.ResultGetStats) {
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
		chainID, stats.IssuerChainID, stats.TokenID,
		stats.Issuance.Type, stats.Issuance.Symbol,
		stats.Issuance.Supply, stats.CirculatingSupply, stats.Burned,
		stats.Transactions,
		stats.IssuanceTimestamp.Time())
	if stats.LastTransactionTimestamp != nil {
		fmt.Printf("Last Tx Timestamp: %v\n",
			stats.LastTransactionTimestamp.Time())
	}
	fmt.Println()

}
