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

	jrpc "github.com/AdamSLevy/jsonrpc2/v11"
	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat"
	"github.com/posener/complete"
	"github.com/spf13/cobra"
)

// issueChainCmd represents the createchain command
var issueChainCmd = func() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chain",
		Short: "Create a new FAT token chain",
		Long: `Compose or submit the Chain Creation Entry for a new FAT token chain.

Creating a new chain with the correct Name IDs is the first of two steps to
issue a new FAT token.`,
		Args:    cobra.ExactArgs(0),
		PreRunE: validateIssueChainFlags,
		Run:     issueChain,
	}
	issueCmd.AddCommand(cmd)
	issueCmplCmd.Sub["chain"] = issueChainCmplCmd
	rootCmplCmd.Sub["help"].Sub["issue"].Sub["chain"] = complete.Command{}

	generateCmplFlags(cmd, issueChainCmplCmd.Flags)
	// Don't complete these global flags as they are ignored by this
	// command.
	for _, flg := range []string{"-c", "--chainid"} {
		delete(issueChainCmplCmd.Flags, flg)
	}
	usage := cmd.UsageFunc()
	cmd.SetUsageFunc(func(cmd *cobra.Command) error {
		cmd.Flags().MarkHidden("chainid")
		return usage(cmd)
	})
	return cmd
}()

var issueChainCmplCmd = complete.Command{
	Flags: mergeFlags(apiCmplFlags, tokenCmplFlags, ecAdrCmplFlags),
}

func validateIssueChainFlags(cmd *cobra.Command, _ []string) error {
	flags := cmd.Flags()
	if flags.Changed("tokenid") || flags.Changed("identity") {
		if !flags.Changed("tokenid") || !flags.Changed("identity") {
			return fmt.Errorf("--tokenid and --identity must be used together")
		}
		if flags.Changed("chainid") {
			return fmt.Errorf(
				"--chainid may not be used with --tokenid or --identity")
		}
		*paramsToken.ChainID = fat.ChainID(paramsToken.TokenID,
			*paramsToken.IssuerChainID)

		return nil
	}
	return fmt.Errorf("--tokenid and --identity are required")
}

var (
	missingChainHeadErr      = jrpc.Error{Code: -32009, Message: "Missing Chain Head"}
	newChainInProcessListErr = jrpc.Error{Message: "new chain in process list"}
)

func issueChain(_ *cobra.Command, _ []string) {
	var first factom.Entry
	first.ExtIDs = fat.NameIDs(paramsToken.TokenID, *paramsToken.IssuerChainID)
	if !force {
		eb := factom.EBlock{ChainID: paramsToken.ChainID}
		err := eb.GetChainHead(FactomClient)
		if err == nil {
			fmt.Printf("Chain %v already exists.\n", eb.ChainID)
			// We can consider this a success. Exit code 0.
			return
		}
		rpcErr, ok := err.(jrpc.Error)
		if ok && rpcErr == newChainInProcessListErr {
			fmt.Printf("New chain %v is in process list. Wait ~10 mins.\n",
				eb.ChainID)
			// We can consider this a success. Exit code 0.
			return
		}
		if !ok || rpcErr != missingChainHeadErr {
			// If err was anything other than the missingChainHeadErr...
			fmt.Println(err)
			os.Exit(1)
		}

		cost, err := first.Cost()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if uint64(cost) > ecBalance {
			fmt.Println("Insufficient EC balance")
			os.Exit(1)
		}
	}

	if curl {
		if err := printCurl(first, ecEsAdr.Es); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		return
	}

	txID, err := first.ComposeCreate(FactomClient, ecEsAdr.Es)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("Chain created: %v\n", first.ChainID)
	fmt.Printf("Factom Tx ID:  %v\n", txID)
	return
}

func printCurl(entry factom.Entry, es factom.EsAddress) error {
	newChain := entry.ChainID == nil
	commit, reveal, _, err := entry.Compose(es)
	if err != nil {
		return err
	}

	commitMethod := "commit"
	revealMethod := "reveal"
	if newChain {
		commitMethod += "-chain"
		revealMethod += "-chain"
	}

	commitHex, _ := factom.Bytes(commit).MarshalJSON()
	fmt.Printf(`curl -X POST --data-binary '{"jsonrpc": "2.0", "id": 0, "method": "%v", "params":{"message":%v}}' -H 'content-type:text/plain;' %v/v2`,
		commitMethod, string(commitHex), FactomClient.FactomdServer)
	fmt.Println()

	revealHex, _ := factom.Bytes(reveal).MarshalJSON()
	fmt.Printf(`curl -X POST --data-binary '{"jsonrpc": "2.0", "id": 0, "method": "%v", "params":{"entry":%v}}' -H 'content-type:text/plain;' %v/v2`,
		revealMethod, string(revealHex), FactomClient.FactomdServer)
	fmt.Println()
	return nil
}
