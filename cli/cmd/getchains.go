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

	"github.com/Factom-Asset-Tokens/factom"
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
		cmd.Flags().MarkHidden("tokenid")
		cmd.Flags().MarkHidden("identity")
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
		if err := FATClient.Request(context.Background(),
			"get-daemon-tokens", nil, &chains); err != nil {
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
		params := srv.ParamsToken{ChainID: &chainID,
			IncludePending: paramsToken.IncludePending}
		var stats srv.ResultGetStats
		if err := FATClient.Request(context.Background(),
			"get-stats", params, &stats); err != nil {
			errLog.Fatal(err)
		}
		printStats(&chainID, stats)
	}
}

func printStats(chainID *factom.Bytes32, stats srv.ResultGetStats) {
	fmt.Printf(`Chain ID: %v
Issuer Identity Chain ID: %v
Issuance Entry Hash: %v
Token ID: %v
Type: %v
Symbol: %q
Precision: %q
Supply:            %v
Ciculating Supply: %v
Burned:            %v
Number of Transactions: %v
Issuance Timestamp: %v
`,
		chainID, stats.IssuerChainID, stats.IssuanceHash, stats.TokenID,
		stats.Issuance.Type, stats.Issuance.Symbol,
		stats.Issuance.Supply, stats.CirculatingSupply, stats.Burned,
		stats.Transactions,
		stats.IssuanceTimestamp)
	if stats.LastTransactionTimestamp > 0 {
		fmt.Printf("Last Tx Timestamp: %v\n",
			stats.LastTransactionTimestamp)
	}
	fmt.Println()

}
