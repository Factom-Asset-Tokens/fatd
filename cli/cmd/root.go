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

	ChainID         factom.Bytes32
	TokenID         string
	IdentityChainID factom.Bytes32
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

fat-cli queries the fatd API to explore FAT chains. It can compose the entries
required to create new FAT chains or transact FAT tokens and submit them
directly to the Factom blockchain via a factomd node.
`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
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

	flags.VarP(&ChainID, "chainid", "c", "the Chain ID of a FAT chain tracked by fatd")
	flags.Lookup("chainid").DefValue = "none"
	flags.StringVarP(&TokenID, "tokenid", "t", "",
		"Token ID of a FAT chain tracked by fatd")
	flags.VarP(&IdentityChainID, "identity", "i",
		"Chain ID of the Identity Chain for a FAT chain tracked by fatd")
	flags.Lookup("identity").DefValue = "none"

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
	FactomClient.Factomd.Timeout = FATClient.Timeout
	FactomClient.Walletd.Timeout = FATClient.Timeout
}

// validateChainID validates "chainid", "tokenid" and "identity", and
// initializes the ChainID.
func validateChainID(cmd *cobra.Command, _ []string) error {
	flags := cmd.Flags()
	chainidF := flags.Lookup("chainid")
	tokenidF := flags.Lookup("tokenid")
	identityF := flags.Lookup("identity")
	if chainidF.Changed {
		if tokenidF.Changed || identityF.Changed {
			return fmt.Errorf("--chainid may not be used with --tokenid or --identity")
		}
		return nil
	}
	if tokenidF.Changed || identityF.Changed {
		if !tokenidF.Changed || !identityF.Changed {
			return fmt.Errorf("--tokenid and --identity must be used together")
		}
		ChainID = fat.ChainID(TokenID, IdentityChainID)
		return nil
	}
	return fmt.Errorf("either --chainid or --tokenid and --identity must be specified")
}
