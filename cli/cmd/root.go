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
	"github.com/spf13/viper"
)

var (
	cfgFile      string
	FATClient    = srv.NewClient()
	FactomClient = factom.NewClient()
	Debug        bool

	paramsToken = srv.ParamsToken{
		ChainID:       new(factom.Bytes32),
		IssuerChainID: new(factom.Bytes32)}
)

// Complete runs the CLI completion.
func Complete() bool {
	comp := complete.New("fat-cli", rootCmplCmd)
	return comp.Complete()
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "fat-cli",
	Short: "Factom Asset Tokens CLI",
	Long: `fat-cli allows users to explore and interact with FAT chains.

fat-cli can be used to explore FAT chains to view balances, issuance, and
transaction data. It can also be used to send transactions on existing FAT
chains, and issue new FAT-0 or FAT-1 chains.

API Settings

fat-cli needs to be able to query the API of a running fatd node to explore FAT
chains. Use --fatd to specify the fatd endpoint, if not on
http://localhost:8078.

fat-cli needs to be able to query factom-walletd in order to access private
keys for transaction signing and paying for Factom entries. Use --walletd to
set the factom-walletd endpoint, if not on http://localhost:8089.

fat-cli needs to be able to query factomd in order to submit signed transaction
or issuance entries. Use --factomd to specify the factomd endpoint, if not on
http://localhost:8088.

Chain ID Settings

Most sub-commands need to be scoped to a specific FAT chain. This can be done
by specifying both the --tokenid and --identity Chain ID, or just the
--chainid.
`,
}

var rootCmplCmd = complete.Command{
	Flags: mergeFlags(apiFlags, tokenFlags),
	Sub:   complete.Commands{},
}

var apiFlags = complete.Flags{
	"--fatd":    complete.PredictAnything,
	"--factomd": complete.PredictAnything,
	"--walletd": complete.PredictAnything,
	"--timeout": complete.PredictAnything,
	"--debug":   complete.PredictNothing,
}
var tokenFlags = complete.Flags{
	"--chainid":  PredictChainIDs,
	"-c":         PredictChainIDs,
	"--tokenid":  complete.PredictAnything,
	"--identity": complete.PredictAnything,
}

func mergeFlags(flgs ...complete.Flags) complete.Flags {
	var size int
	for _, flg := range flgs {
		size += len(flg)
	}
	f := make(complete.Flags, size)
	for _, flg := range flgs {
		for k, v := range flg {
			f[k] = v
		}
	}
	return f
}

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
	cobra.OnInitialize(initTimeouts)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	//rootCmd.PersistentFlags().
	//	StringVar(&cfgFile, "config", "",
	//		"config file (default is $HOME/.fat-cli.yaml)")
	flags := rootCmd.PersistentFlags()
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

	flags.VarP(paramsToken.ChainID, "chainid", "c",
		"Chain ID of a FAT chain tracked by fatd")
	flags.Lookup("chainid").DefValue = "none"
	flags.StringVarP(&paramsToken.TokenID, "tokenid", "t", "",
		"Token ID of a FAT chain tracked by fatd")
	flags.VarP(paramsToken.IssuerChainID, "identity", "i",
		"Issuer Identity Chain ID of a FAT chain tracked by fatd")
	flags.Lookup("identity").DefValue = "none"
	flags.BoolVar(&Debug, "debug", false, "Print all RPC requests and responses")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	//rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
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

// initTimeouts set the same timeout for all Clients.
func initTimeouts() {
	FATClient.DebugRequest = Debug
	FactomClient.Factomd.DebugRequest = Debug
	FactomClient.Walletd.DebugRequest = Debug
	FactomClient.Factomd.Timeout = FATClient.Timeout
	FactomClient.Walletd.Timeout = FATClient.Timeout
}

// validateChainIDFlags validates "chainid", "tokenid" and "identity", and
// initializes the ChainID.
func validateChainIDFlags(cmd *cobra.Command, _ []string) error {
	flags := cmd.Flags()
	defer func() {
		paramsToken.TokenID = ""
		paramsToken.IssuerChainID = nil
	}()
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
