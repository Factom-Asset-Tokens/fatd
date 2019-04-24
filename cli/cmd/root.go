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
	"strings"
	"time"

	jrpc "github.com/AdamSLevy/jsonrpc2/v11"
	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat"
	"github.com/Factom-Asset-Tokens/fatd/srv"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/posener/complete"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	//cobra.OnInitialize(initConfig)
	cobra.OnInitialize(initClients)
}
func initClients() {
	// Only use Debug if true to avoid always overriding --debugfactomd and
	// --debugfatd flags.
	if Debug {
		FATClient.DebugRequest = Debug
		FactomClient.Factomd.DebugRequest = Debug
		// Do not use DebugRequest for factom-walletd to avoid leaking
		// private keys.
		// Use --debugwalletd explicitly to debug wallet API calls.
	}

	for _, client := range []*jrpc.Client{
		&FATClient.Client,
		&FactomClient.Factomd,
		&FactomClient.Walletd,
	} {
		// Use of Basic Auth with empty User and Password is not
		// supported.
		// --fatduser "" --fatdpassword "" has no effect.
		if len(client.User)+len(client.Password) > 0 {
			client.BasicAuth = true
		}
		client.Timeout = FATClient.Timeout
	}

	for _, url := range []*string{
		&FATClient.FatdServer,
		&FactomClient.FactomdServer,
		&FactomClient.WalletdServer,
	} {
		// Add "http://" if no scheme was specified.
		addHTTPScheme(url)
	}
}
func addHTTPScheme(url *string) {
	strs := strings.Split(*url, "://")
	if len(strs) == 1 {
		*url = "http://" + *url
	}
}

var (
	cfgFile      string
	FATClient    = srv.NewClient()
	FactomClient = factom.NewClient()
	Debug        bool

	paramsToken = srv.ParamsToken{
		ChainID:       new(factom.Bytes32),
		IssuerChainID: new(factom.Bytes32)}
)

var apiFlags = func() *flag.FlagSet {
	flags := flag.NewFlagSet("", flag.ContinueOnError)
	flags.ParseErrorsWhitelist.UnknownFlags = true

	flags.StringVarP(&FATClient.FatdServer, "fatd", "d",
		"localhost:8078", "scheme://host:port for fatd")
	flags.StringVarP(&FactomClient.FactomdServer, "factomd", "s",
		"localhost:8088", "scheme://host:port for factomd")
	flags.StringVarP(&FactomClient.WalletdServer, "walletd", "w",
		"localhost:8089", "scheme://host:port for factom-walletd")

	flags.StringVar(&FATClient.User, "fatduser", "",
		"Basic HTTP Auth User for fatd")
	flags.StringVar(&FactomClient.Factomd.User, "factomduser", "",
		"Basic HTTP Auth User for factomd")
	flags.StringVar(&FactomClient.Walletd.User, "walletduser", "",
		"Basic HTTP Auth User for factom-walletd")

	flags.StringVar(&FATClient.Password, "fatdpass", "",
		"Basic HTTP Auth Password for fatd")
	flags.StringVar(&FactomClient.Factomd.Password, "factomdpass", "",
		"Basic HTTP Auth Password for factomd")
	flags.StringVar(&FactomClient.Walletd.Password, "walletdpass", "",
		"Basic HTTP Auth Password for factom-walletd")

	flags.DurationVar(&FATClient.Timeout, "timeout", 3*time.Second,
		"Timeout for all API requests (i.e. 10s, 1m)")

	flags.BoolVar(&Debug, "debug", false, "Print fatd and factomd API calls")
	flags.BoolVar(&FATClient.DebugRequest, "debugfatd", false,
		"Print fatd API calls")
	flags.BoolVar(&FactomClient.Factomd.DebugRequest, "debugfactomd", false,
		"Print factomd API calls")
	flags.BoolVar(&FactomClient.Walletd.DebugRequest, "debugwalletd", false,
		"Print factom-walletd API calls")
	flags.MarkHidden("debugfatd")
	flags.MarkHidden("debugfactomd")
	flags.MarkHidden("debugwalletd")
	return flags
}()

// rootCmd represents the base command when called without any subcommands
var rootCmd = func() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fat-cli",
		Short: "Factom Asset Tokens CLI",
		Long: `
fat-cli allows users to explore and interact with FAT chains.

fat-cli can be used to explore FAT chains to get balances, issuance, and
transaction data. It can also be used to send transactions on existing FAT
chains, and issue new FAT-0 or FAT-1 tokens.

Chain ID Settings
        Most sub-commands need to be scoped to a specific FAT chain identified
        by a --chainid. Alternatively, this can be specified by using both the
        --tokenid and --identity, which together determine the chain ID.

API Settings
        fat-cli makes use of the fatd, factomd, and factom-walletd JSON-RPC 2.0
        APIs for various operations. Trust in these API endpoints is imperative
        to secure operation.

        The --fatd API is used to explore issuance, transactions, and balances
        for existing FAT chains.

        The --factomd API is used to submit entries directly to the Factom
        blockchain, as well as for checking EC balances, chain existence, and
        identity keys.

        The --walletd API is used to access private keys for FA and EC
        addresses. To avoid use of factom-walletd, use private Fs or Es keys
        directly on the CLI instead.

        If --debug is set, all fatd and factomd API calls will be printed to
        stdout. API calls to factom-walletd are omitted to avoid leaking
        private key data.

Offline Mode
        For increased security requirements to protect private keys, it is
        possible to run fat-cli such that it makes no network calls when
        generating Factom entries for FAT transactions or token issuance.

        Use --curl to skip submitting the entry directly to Factom, and instead
        print out the curl commands to commit and reveal the entry. These curl
        commands contain the encoded signed data and may be safely copied to,
        and run from, a computer with access to factomd.

        Use --force to skip all sanity checks that involve API calls out
        factomd or fatd. As a result, this may result in generating a Factom
        Entry that is invalid for Factom or FAT, but may still use up Entry
        Credits to submit.

        Use private keys for --ecadr and --input directly to avoid any network
        calls to factom-walletd.

Entry Credits
        Making FAT transactions or issuing new FAT tokens requires creating
        entries on the Factom blockchain. Creating Factom entries costs Entry
        Credits. Entry Credits have a relatively fixed price of about $0.001
        USD. Entry Credits can be obtained by burning Factoids which can be
        done using the official factom-cli.

        FAT transactions normally cost 1 EC. The full FAT Token Issuance
        process normally costs 12 EC.
`[1:],
		Args:    cobra.ExactArgs(0),
		PreRunE: validateRunCompletionFlags,
		Run:     runCompletion,
	}

	cmd.Flags().AddFlagSet(installCompletionFlags)

	flags := cmd.PersistentFlags()
	flags.AddFlagSet(apiFlags)
	flags.VarPF(paramsToken.ChainID, "chainid", "C",
		"Chain ID of a FAT chain").DefValue = ""
	flags.StringVarP(&paramsToken.TokenID, "tokenid", "T", "",
		"Token ID of a FAT chain")
	flags.VarPF(paramsToken.IssuerChainID, "identity", "I",
		"Issuer Identity Chain ID of a FAT chain").DefValue = ""

	generateCmplFlags(cmd, rootCmplCmd.Flags)
	return cmd
}()

var rootCmplCmd = complete.Command{
	Flags: mergeFlags(apiCmplFlags, tokenCmplFlags),
	Sub:   complete.Commands{"help": complete.Command{Sub: complete.Commands{}}},
}
var apiCmplFlags = complete.Flags{
	"--help": complete.PredictNothing,
}
var tokenCmplFlags = complete.Flags{
	"--chainid": PredictChainIDs,
	"-c":        PredictChainIDs,
}

func validateRunCompletionFlags(cmd *cobra.Command, _ []string) error {
	// Ensure that the install completion flags are not ever used with any
	// other flags.
	flags := cmd.Flags()
	installCompletionMode := false
	otherFlags := false
	flags.Visit(func(flg *flag.Flag) {
		switch flg.Name {
		case "install", "uninstall":
			installCompletionMode = true
		default:
			otherFlags = true
		}
	})
	if installCompletionMode && otherFlags {
		return fmt.Errorf(
			"--install and --uninstall may not be used with any other flags")
	}
	return nil
}

func runCompletion(cmd *cobra.Command, _ []string) {
	// Complete() returns true if it attempts to install completion,
	// otherwise just output the help page.
	if !Complete() {
		cmd.Help()
	}
}

// validateChainIDFlags validates "chainid", "tokenid" and "identity", and
// initializes the ChainID.
func validateChainIDFlags(cmd *cobra.Command, _ []string) error {
	flags := cmd.Flags()
	if flags.Changed("chainid") {
		if flags.Changed("tokenid") || flags.Changed("identity") {
			return fmt.Errorf(
				"--chainid may not be used with --tokenid or --identity")
		}
		return nil
	}
	if flags.Changed("tokenid") || flags.Changed("identity") {
		if !flags.Changed("tokenid") || !flags.Changed("identity") {
			return fmt.Errorf("--tokenid and --identity must be used together")
		}
		*paramsToken.ChainID = fat.ChainID(paramsToken.TokenID,
			*paramsToken.IssuerChainID)

		return nil
	}
	return fmt.Errorf("--chainid or both --tokenid and --identity is required")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".cli" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".fat-cli")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
