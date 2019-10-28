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

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Factom-Asset-Tokens/factom"
	"github.com/Factom-Asset-Tokens/fatd/api"
	"github.com/Factom-Asset-Tokens/fatd/fat1"

	"github.com/posener/complete"
	"github.com/spf13/cobra"
)

var (
	paramsGetTxs = api.ParamsGetTransactions{
		StartHash:   new(factom.Bytes32),
		NFTokenID:   new(fat1.NFTokenID),
		ParamsToken: api.ParamsToken{ChainID: paramsToken.ChainID},
		ParamsPagination: api.ParamsPagination{Page: new(uint),
			Order: "desc"},
	}
	to, from       bool
	transactionIDs []factom.Bytes32
)

// getTxsCmd represents the transactions command
var getTxsCmd = func() *cobra.Command {
	cmd := &cobra.Command{
		DisableFlagsInUseLine: true,
		Use: `
transactions --chainid <chain-id> TXHASH...

  fat-cli get transactions --chainid <chain-id> [--starttx <tx-hash>]
        [--page <page>] [--limit <limit>] [--order <"asc" | "desc">]
        [--address <FA> [--address <FA>]... [--to] [--from]]
        [--nftokenid <nf-token-id>]
`[1:],
		Aliases: []string{"transaction", "txs", "tx"},
		Short:   "List transactions and their data",
		Long: `
For the given --chainid, get transaction data for each TXID or list
transactions scoped by the search criteria provided by flags.

If at least one TXID is provided, then the data for each transaction is
returned. Only global flags are accepted with TXIDs.

If no TXID is provided, then a paginated list of all transactions is returned.
The list can be scoped down to transactions --to or --from one --address or
more, and in the case of a FAT-1 chain, by a single --nftokenid. Use --page and
--limit to scroll through transactions.
`[1:],
		Args:    getTxsArgs,
		PreRunE: validateGetTxsFlags,
		Run:     getTxs,
	}
	getCmd.AddCommand(cmd)
	getCmplCmd.Sub["transactions"] = getTxsCmplCmd
	rootCmplCmd.Sub["help"].Sub["get"].Sub["transactions"] = complete.Command{}

	flags := cmd.Flags()
	flags.UintVarP(paramsGetTxs.Page, "page", "p", 1, "Page of returned txs")
	flags.UintVarP(&paramsGetTxs.Limit, "limit", "l", 10, "Limit of returned txs")
	flags.VarPF((*txOrder)(&paramsGetTxs.Order), "order", "", "Order of returned txs").
		DefValue = "desc"
	flags.BoolVar(&to, "to", false, "Request only txs TO the given --address set")
	flags.BoolVar(&from, "from", false, "Request only txs FROM the given --address set")
	flags.VarPF(paramsGetTxs.StartHash, "starttx", "",
		"Hash of tx to start indexing from").DefValue = ""
	flags.Uint64Var((*uint64)(paramsGetTxs.NFTokenID), "nftokenid", 0,
		"Request only txs involving this NF Token ID")
	flags.VarPF((*FAAddressList)(&paramsGetTxs.Addresses), "address", "a",
		"Add to the set of addresses to lookup txs for").DefValue = ""

	generateCmplFlags(cmd, getTxsCmplCmd.Flags)
	return cmd
}()

var getTxsCmplCmd = complete.Command{
	Flags: mergeFlags(apiCmplFlags, tokenCmplFlags,
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
	paramsGetTxs.IncludePending = paramsToken.IncludePending
	flags := cmd.LocalFlags()
	if len(transactionIDs) > 0 {
		for _, flgName := range []string{"page", "order", "page", "limit",
			"starttx", "to", "from", "nftokenid", "address"} {
			if flags.Changed(flgName) {
				return fmt.Errorf("--%v is incompatible with TXID arguments",
					flgName)
			}
		}
		return nil
	}

	if flags.Changed("to") || flags.Changed("from") {
		if len(paramsGetTxs.Addresses) == 0 {
			return fmt.Errorf(
				"--to and --from require at least one --address")
		}
		if to != from { // Setting --to and --from is the same as omitting both.
			if to {
				paramsGetTxs.ToFrom = "to"
			} else {
				paramsGetTxs.ToFrom = "from"
			}
		}
	}

	if !flags.Changed("starttx") {
		paramsGetTxs.StartHash = nil
	}

	if !flags.Changed("nftokenid") {
		paramsGetTxs.NFTokenID = nil
	}
	if flags.Changed("page") {
		if *paramsGetTxs.Page == 0 {
			return fmt.Errorf("--page cannot be 0, starts at 1")
		}
		if *paramsGetTxs.Page == 1 {
			// No need to explicitly send "page": 1
			paramsGetTxs.Page = nil
		}
	}

	return nil
}

func getTxs(_ *cobra.Command, _ []string) {
	vrbLog.Printf("Fetching txs for chain... %v",
		paramsToken.ChainID)
	if len(transactionIDs) == 0 {
		result := make([]api.ResultGetTransaction, paramsGetTxs.Limit)
		for i := range result {
			result[i].Tx = &json.RawMessage{}
		}
		if err := FATClient.Request(context.Background(),
			"get-transactions", paramsGetTxs, &result); err != nil {
			errLog.Fatal(err)
		}
		for _, result := range result {
			printTx(result)
		}
		return
	}
	params := api.ParamsGetTransaction{ParamsToken: paramsGetTxs.ParamsToken}
	var result api.ResultGetTransaction
	var tx json.RawMessage
	result.Tx = &tx
	for _, txID := range transactionIDs {
		vrbLog.Printf("Fetching tx details... %v", txID)
		params.Hash = &txID
		if err := FATClient.Request(context.Background(),
			"get-transaction", params, &result); err != nil {
			errLog.Fatal(err)
		}
		printTx(result)
	}
	return
}

func printTx(result api.ResultGetTransaction) {
	if result.Pending {
		fmt.Println("PENDING TX")
	}
	fmt.Println("TXID:", result.Hash)
	fmt.Println("Timestamp:", time.Unix(result.Timestamp, 0))
	fmt.Println("TX:", (string)(*result.Tx.(*json.RawMessage)))

	fmt.Println()
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
	case "des", "desc", "descending", "latest":
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
