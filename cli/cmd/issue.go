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
	"github.com/posener/complete"
	"github.com/spf13/cobra"
)

var (
	ecEsAdr   ECEsAddress
	ecBalance uint64
	force     bool
	curl      bool
)

// issueCmd represents the issue command
var issueCmd = func() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "issue",
		Short: "Issue a new token chain",
		Long: `Issue a new FAT-0 or FAT-1 token chain.

Issuing a new FAT token chain is a two step process. First, the token chain
must be created with the correct Name IDs in the first entry. Second, the
signed Initialization Entry must be submitted. Chain creation takes a full
Factom block ~10 mins.
`,
		PersistentPreRunE: validateECAdrFlag,
	}
	rootCmd.AddCommand(cmd)
	rootCmplCmd.Sub["issue"] = issueCmplCmd
	rootCmplCmd.Sub["help"].Sub["issue"] = complete.Command{Sub: complete.Commands{}}

	flags := cmd.PersistentFlags()
	flags.VarPF(&ecEsAdr, "ecadr", "e", "EC or Es address to pay for entries").
		DefValue = "none"
	flags.BoolVarP(&force, "force", "f", false,
		"Skip sanity checks like EC balance, chain existence, and identity")
	flags.BoolVar(&curl, "curl", false, "Do not submit Factom entry; print curl commands")
	generateCmplFlags(cmd, issueCmplCmd.Flags)
	return cmd
}()

var ecAdrCmplFlags = complete.Flags{
	"--ecadr": PredictECAddresses,
	"-e":      PredictECAddresses,
}

var issueCmplCmd = complete.Command{
	Flags: mergeFlags(apiCmplFlags, tokenCmplFlags, ecAdrCmplFlags),
	Sub:   complete.Commands{},
}

func validateECAdrFlag(cmd *cobra.Command, _ []string) error {
	if !cmd.Flags().Changed("ecadr") {
		return fmt.Errorf("--ecadr is required")
	}

	// All subsequent errors are not issues with correct use of flags, so
	// avoid printing Usage() by calling os.Exit(1) instead of returning.

	// Get the private Es Address if an EC address was given.
	var zero factom.EsAddress
	var err error
	if ecEsAdr.Es == zero {
		ecEsAdr.Es, err = ecEsAdr.EC.GetEsAddress(FactomClient)
		if err != nil {
			if err, ok := err.(jrpc.Error); ok {
				fmt.Println(err.Data, ecEsAdr.EC)
			} else {
				fmt.Println(err)
			}
			os.Exit(1)
		}
	}
	if force {
		// Skip balance check.
		return nil
	}
	ecBalance, err = ecEsAdr.EC.GetBalance(FactomClient)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return nil
}

type ECEsAddress struct {
	EC factom.ECAddress
	Es factom.EsAddress
}

func (e *ECEsAddress) Set(adrStr string) error {
	if err := e.EC.Set(adrStr); err != nil {
		if err := e.Es.Set(adrStr); err != nil {
			return err
		}
		e.EC = e.Es.ECAddress()
	}
	return nil
}

func (e ECEsAddress) String() string {
	return e.EC.String()
}

func (ECEsAddress) Type() string {
	return "EC|Es"
}
