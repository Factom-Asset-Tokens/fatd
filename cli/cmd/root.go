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
	"time"

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

var (
	cfgFile      string
	FATClient    = srv.NewClient()
	FactomClient = factom.NewClient()
	Debug        bool

	paramsToken = srv.ParamsToken{
		ChainID:       new(factom.Bytes32),
		IssuerChainID: new(factom.Bytes32)}
)

func init() {
	//cobra.OnInitialize(initConfig)
	cobra.OnInitialize(initClients)
}

// initClients sets the same timeout and debug settings for all Clients.
func initClients() {
	FATClient.DebugRequest = Debug
	FactomClient.Factomd.DebugRequest = Debug
	FactomClient.Walletd.DebugRequest = Debug
	FactomClient.Factomd.Timeout = FATClient.Timeout
	FactomClient.Walletd.Timeout = FATClient.Timeout
}

var apiFlags = func() *flag.FlagSet {
	flags := flag.NewFlagSet("", flag.ContinueOnError)
	flags.ParseErrorsWhitelist.UnknownFlags = true
	flags.StringVarP(&FATClient.FatdServer, "fatd", "d", "http://localhost:8078",
		"scheme://host:port for fatd")
	flags.StringVarP(&FactomClient.FactomdServer, "factomd", "s",
		"http://localhost:8088",
		"scheme://host:port for factomd")
	flags.StringVarP(&FactomClient.WalletdServer, "walletd", "w",
		"http://localhost:8089",
		"scheme://host:port for factom-walletd")
	flags.DurationVar(&FATClient.Timeout, "timeout", 6*time.Second,
		"Timeout for all API requests (i.e. 10s, 1m)")
	flags.BoolVar(&Debug, "debug", false, "Print all RPC requests and responses")
	return flags
}()

// rootCmd represents the base command when called without any subcommands
var rootCmd = func() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fat-cli",
		Short: "Factom Asset Tokens CLI",
		Long: `fat-cli allows users to explore and interact with FAT chains.

fat-cli can be used to explore FAT chains to get balances, issuance, and
transaction data. It can also be used to send transactions on existing FAT
chains, and issue new FAT-0 or FAT-1 chains.

Chain ID Settings

Most sub-commands need to be scoped to a specific FAT chain using --chainid or
both the --tokenid and --identity.

API Settings

fat-cli needs to be able to query the API of a running fatd node to explore FAT
chains. Use --fatd to specify the fatd endpoint, if not on
http://localhost:8078.

fat-cli needs to be able to query factom-walletd in order to access private
keys for transaction signing and paying for Factom entries. Use --walletd to
set the factom-walletd endpoint, if not on http://localhost:8089.

fat-cli needs to be able to query factomd in order to submit signed transaction
or issuance entries. Use --factomd to specify the factomd endpoint, if not on
http://localhost:8088.`,
		Args:    cobra.ExactArgs(0),
		PreRunE: validateRunCompletionFlags,
		Run:     runCompletion,
	}

	cmd.Flags().AddFlagSet(installCompletionFlags)
	flags := cmd.PersistentFlags()
	// API Flags
	flags.AddFlagSet(apiFlags)
	// Chain ID Flags
	flags.VarP(paramsToken.ChainID, "chainid", "c",
		"Chain ID of a FAT chain")
	flags.Lookup("chainid").DefValue = "none"
	flags.StringVarP(&paramsToken.TokenID, "tokenid", "t", "",
		"Token ID of a FAT chain")
	flags.VarP(paramsToken.IssuerChainID, "identity", "i",
		"Issuer Identity Chain ID of a FAT chain")
	flags.Lookup("identity").DefValue = "none"

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
	return fmt.Errorf("either --chainid or --tokenid and --identity must be specified")
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
