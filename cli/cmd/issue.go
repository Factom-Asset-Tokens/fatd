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

	jrpc "github.com/AdamSLevy/jsonrpc2/v11"
	"github.com/posener/complete"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
)

var (
	ecEsAdr ECEsAddress
	force   bool
	curl    bool
)

var composeFlags = func() *flag.FlagSet {
	flags := flag.NewFlagSet("", flag.ContinueOnError)
	flags.VarPF(&ecEsAdr, "ecadr", "e",
		"EC or Es address to pay for entries").DefValue = ""
	flags.BoolVar(&force, "force", false,
		"Skip sanity checks for balances, chain status, and sk1 key")
	flags.BoolVar(&curl, "curl", false,
		"Do not submit Factom entry; print curl commands")
	return flags
}()

// issueCmd represents the issue command
var issueCmd = func() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "issue",
		Short: "Issue a new token chain",
		Long: `
Issue a new FAT-0 or FAT-1 token chain.

Issuing a new FAT token chain is a two step process. First the Token Chain must
be created on the Factom Blockchain. Both --tokenid and --identity are
required. Use of --chainid is not allowed for this step.

fat-cli issue chain --ecadr <EC | Es> --identity <issuer-identity-chain-id>
        --tokenid <token-id>

Second, the Token Initialization Entry must be added to the Token Chain. Since
Factom chain creation takes a full Factom block before entries can be added,
the process may take up to 10 minutes.

fat-cli issue token --ecadr <EC | Es> --chainid <token-chain-id>
        --sk1 <sk1-key> --type <FAT-0|FAT-1> --supply <max-supply>

Entry Credits
        Creating entries on the Factom blockchain costs Entry Credits. The full
        Token Issuance process normally costs 12 ECs. You must specify a funded
        Entry Credit address with --ecadr, which may be either a private Es
        address, or a pubilc EC address that can be fetched from
        factom-walletd.

Identity Chain
        FAT token chains may only be issued by an entity controlling the
        sk1/id1 key established by the Identity Chain pointed to by the FAT
        token chain. An Identity Chain and the associated keys can be created
        using the factom-identity-cli.

        https://github.com/PaulBernier/factom-identity-cli
`[1:],
		PersistentPreRunE: validateECAdrFlag,
	}
	rootCmd.AddCommand(cmd)
	rootCmplCmd.Sub["issue"] = issueCmplCmd
	rootCmplCmd.Sub["help"].Sub["issue"] = complete.Command{Sub: complete.Commands{}}

	cmd.PersistentFlags().AddFlagSet(composeFlags)

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
	// avoid printing Usage() by calling errLog.Fatal instead of returning.

	// Get the private Es Address if an EC address was given.
	var zero factom.EsAddress
	var err error
	if ecEsAdr.Es == zero {
		vrbLog.Println("Fetching secret address...", ecEsAdr.EC)
		ecEsAdr.Es, err = ecEsAdr.EC.GetEsAddress(FactomClient)
		if err != nil {
			if err, ok := err.(jrpc.Error); ok {
				errLog.Fatal(err.Data, ecEsAdr.EC)
			}
			errLog.Fatal(err)
		}
	}
	return nil
}

func verifyECBalance(ec *factom.ECAddress, cost int8) {
	vrbLog.Println("Checking EC balance... ")
	ecBalance, err := ec.GetBalance(FactomClient)
	if err != nil {
		errLog.Fatal(err)
	}
	if uint64(cost) > ecBalance {
		errLog.Fatalf("Insufficient EC balance %v: needs at least %v",
			ecBalance, cost)
	}
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
	return "<EC | Es>"
}
