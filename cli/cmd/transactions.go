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
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat/fat1"
	"github.com/Factom-Asset-Tokens/fatd/srv"
	"github.com/posener/complete"
	"github.com/spf13/cobra"
)

var (
	paramsGetTxs = srv.ParamsGetTransactions{
		Page: new(uint64), Limit: new(uint64),
		StartHash:   new(factom.Bytes32),
		NFTokenID:   new(fat1.NFTokenID),
		ParamsToken: srv.ParamsToken{ChainID: paramsToken.ChainID},
	}
	to, from       bool
	transactionIDs []factom.Bytes32
)

// getTxsCmd represents the transactions command
var getTxsCmd = func() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "transactions [flags]|[TXID...]",
		Aliases:               []string{"transaction", "txs", "tx"},
		DisableFlagsInUseLine: true,
		Short:                 "List txs and their data",
		Long: `Get tx data for each TXID or list txs scoped by the search criteria provided by
flags on the given --chainid (or --tokenid and --identity).

If at least one TXID is provided, then the data for each tx is returned. Only
global flags are accepted with TXIDs.

If no TXID is provided, then a paginated list of txs is returned. The list can
be scoped down to txs --to or --from one --address or more, and in the case of
a FAT-1 chain, by a single --nftokenid. Use --page and --limit to scroll
through txs.`,
		Args:    getTxsArgs,
		PreRunE: validateGetTxsFlags,
		Run:     getTxs,
	}
	getCmd.AddCommand(cmd)
	getCmplCmd.Sub["transactions"] = getTxsCmplCmd
	rootCmplCmd.Sub["help"].Sub["get"].Sub["transactions"] = complete.Command{}

	flags := cmd.Flags()
	flags.Uint64VarP(paramsGetTxs.Page, "page", "p", 1, "Page of returned txs")
	flags.Uint64VarP(paramsGetTxs.Limit, "limit", "l", 10, "Limit of returned txs")
	flags.Var((*txOrder)(&paramsGetTxs.Order), "order", "Order of returned txs")
	flags.Lookup("order").DefValue = "asc"
	flags.BoolVar(&to, "to", false, "Request only txs TO the given --address set")
	flags.BoolVar(&from, "from", false, "Request only txs FROM the given --address set")
	flags.Var(paramsGetTxs.StartHash, "starttxhash",
		"Hash of transaction to start indexing from")
	flags.Lookup("starttxhash").DefValue = "none"
	flags.Uint64Var((*uint64)(paramsGetTxs.NFTokenID), "nftokenid", 0,
		"Request only txs involving this NF Token ID")
	flags.VarP((*FAAddressList)(&paramsGetTxs.Addresses), "address", "a",
		"Add to the set of addresses to lookup txs for")
	flags.Lookup("address").DefValue = "none"

	generateCmplFlags(cmd, getTxsCmplCmd.Flags)
	return cmd
}()

var getTxsCmplCmd = complete.Command{
	Flags: mergeFlags(rootCmplCmd.Flags,
		complete.Flags{
			"--order":   complete.PredictSet("asc", "desc"),
			"--address": PredictFAAddresses,
			"-a":        PredictFAAddresses,
		}),
	Args: complete.PredictAnything,
}

func getTxsArgs(_ *cobra.Command, args []string) error {
	transactionIDs = make([]factom.Bytes32, len(args))
	dupl := make(map[factom.Bytes32]struct{}, len(args))
	for i, arg := range args {
		id := &transactionIDs[i]
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

func validateGetTxsFlags(cmd *cobra.Command, args []string) error {
	if err := validateChainIDFlags(cmd, args); err != nil {
		return err
	}
	flags := cmd.LocalFlags()
	if len(transactionIDs) > 0 {
		for _, flgName := range []string{"page", "order", "page", "limit",
			"starttxhash", "to", "from", "nftokenid", "address"} {
			if flags.Changed(flgName) {
				return fmt.Errorf("flag --%v incompatible with TXID arguments",
					flgName)
			}
		}
		return nil
	}

	if flags.Changed("to") || flags.Changed("from") {
		if len(paramsGetTxs.Addresses) == 0 {
			return fmt.Errorf(
				"flags --to and --from require at least one --address")
		}
		if to != from {
			if to {
				paramsGetTxs.ToFrom = "to"
			} else {
				paramsGetTxs.ToFrom = "from"
			}
		}
	}

	if !flags.Changed("starttxhash") {
		paramsGetTxs.StartHash = nil
	}

	if !flags.Changed("nftokenid") {
		paramsGetTxs.NFTokenID = nil
	}

	return nil
}

func getTxs(_ *cobra.Command, _ []string) {
	var stats srv.ResultGetStats
	if err := FATClient.Request("get-stats",
		paramsGetTxs.ParamsToken, &stats); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if len(transactionIDs) > 0 {
		params := srv.ParamsGetTransaction{ParamsToken: paramsGetTxs.ParamsToken}
		result := srv.ResultGetTransaction{}
		tx := json.RawMessage{}
		result.Tx = &tx
		for _, txID := range transactionIDs {
			params.Hash = &txID
			if err := FATClient.Request("get-transaction",
				params, &result); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Println("TXID:", result.Hash)
			fmt.Println("Timestamp:", result.Timestamp.Time())
			fmt.Println("TX:", (string)(*result.Tx.(*json.RawMessage)))
			fmt.Println("")
		}
		return
	}

	result := make([]srv.ResultGetTransaction, *paramsGetTxs.Limit)
	for i := range result {
		result[i].Tx = &json.RawMessage{}
	}
	if err := FATClient.Request("get-transactions", paramsGetTxs, &result); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for _, result := range result {
		fmt.Println("TXID:", result.Hash)
		fmt.Println("Timestamp:", result.Timestamp.Time())
		fmt.Println("TX:", (string)(*result.Tx.(*json.RawMessage)))
		fmt.Println("")
	}
}

type FAAddressList []factom.FAAddress

func (adrs *FAAddressList) Set(adrStr string) error {
	adr, err := factom.NewFAAddress(adrStr)
	if err != nil {
		return err
	}
	*adrs = append(*adrs, adr)
	return nil
}
func (adrs FAAddressList) String() string {
	return fmt.Sprintf("%#v", adrs)
}
func (adrs FAAddressList) Type() string {
	return "FAAddress"
}

type txOrder string

func (o *txOrder) Set(str string) error {
	str = strings.ToLower(str)
	switch str {
	case "asc", "ascending", "earliest":
		*o = "asc"
	case "desc", "descending", "latest":
		*o = "desc"
	default:
		return fmt.Errorf(`must be "asc" or "desc"`)
	}
	return nil
}
func (o txOrder) String() string {
	return string(o)
}
func (o txOrder) Type() string {
	return "asc|desc"
}
