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
	"strings"

	jrpc "github.com/AdamSLevy/jsonrpc2/v11"
	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat"
	"github.com/Factom-Asset-Tokens/fatd/srv"
	"github.com/posener/complete"
	"github.com/spf13/cobra"
)

var (
	Issuance fat.Issuance
	sk1      factom.SK1Key
)

// issueTokenCmd represents the token command
var issueTokenCmd = func() *cobra.Command {
	cmd := &cobra.Command{
		DisableFlagsInUseLine: true,
		Use: `
token --ecadr <EC | Es> --chainid <chain-id> --sk1 <sk1-key>
        --type <"FAT-0" | "FAT-1"> --supply <supply> [--metadata <JSON>]`[1:],
		Short: "Initialize a new FAT token chain",
		Long: `
Compose or submit the Initialization Entry for a new FAT token.

Submitting the Token Initialization Entry is the second and final step to issue
a new FAT token. You must wait until the Factom chain has been created before
this command will succeed. This can take up to 10 minutes. Attempting to run
the command prematurely will fail unless --force and --curl are used.

See 'fat-cli issue chain --help' for information about the first step.

Sanity Checks
        Prior to composing the Token Initialization Entry, a number of calls to
        fatd and factomd are made to ensure that the chain can be created.
        These checks are skipped if --force is used.
        - The token has not already been issued.
        - The --identity chain exists.
        - The --sk1 key corresponds to the --identity's id1 key.
        - The --ecadr has enough ECs to pay for entry creation.
`[1:],
		Args:    cobra.ExactArgs(0),
		PreRunE: validateIssueTokenFlags,
		Run:     issueToken,
	}
	issueCmd.AddCommand(cmd)
	issueCmplCmd.Sub["token"] = issueTokenCmplCmd
	rootCmplCmd.Sub["help"].Sub["issue"].Sub["token"] = complete.Command{}

	flags := cmd.Flags()
	flags.VarPF((*Type)(&Issuance.Type), "type", "",
		"Token standard to use").DefValue = ""
	flags.VarPF(&sk1, "sk1", "", "Secret Identity Key 1 to sign entry").DefValue = ""
	flags.Int64Var(&Issuance.Supply, "supply", 0,
		"Max Token supply, use -1 for unlimited")
	flags.StringVar(&Issuance.Symbol, "symbol", "", "Optional abbreviated token symbol")
	flags.VarPF((*RawMessage)(&Issuance.Metadata), "metadata", "m",
		"JSON metadata to include in tx")

	generateCmplFlags(cmd, issueTokenCmplCmd.Flags)
	return cmd
}()

var issueTokenCmplCmd = complete.Command{
	Flags: mergeFlags(apiCmplFlags, tokenCmplFlags, ecAdrCmplFlags,
		complete.Flags{"--type": complete.PredictSet(fat.TypeFAT0.String(),
			fat.TypeFAT1.String())}),
}

func validateIssueTokenFlags(cmd *cobra.Command, args []string) error {
	if err := validateChainIDFlags(cmd, args); err != nil {
		return err
	}
	flags := cmd.Flags()
	for _, flg := range []string{"type", "supply", "sk1"} {
		if !flags.Changed(flg) {
			return fmt.Errorf("--" + flg + " is required")
		}
	}

	if Issuance.Supply == 0 {
		return fmt.Errorf("--supply may not be 0, use -1 for unlimited supply")
	}

	vrbLog.Println("Preparing and signing Token Initialization Entry...")
	Issuance.ChainID = paramsToken.ChainID
	Issuance.Sign(sk1)
	if err := Issuance.MarshalEntry(); err != nil {
		errLog.Fatal(err)
	}
	cost, err := Issuance.Cost()
	if err != nil {
		errLog.Fatal(err)
	}

	if !force {
		vrbLog.Println("Checking token chain status...", paramsToken.ChainID)
		params := srv.ParamsToken{ChainID: paramsToken.ChainID}
		var stats srv.ResultGetStats
		err := FATClient.Request("get-stats", params, &stats)
		if err == nil {
			errLog.Fatal("Token is already initialized!")
		}
		rpcErr, _ := err.(jrpc.Error)
		if rpcErr != *srv.ErrorTokenNotFound {
			errLog.Fatal(err)
		}

		vrbLog.Println("Checking chain existence...", paramsToken.ChainID)
		eb := factom.EBlock{ChainID: paramsToken.ChainID}
		if err := eb.GetChainHead(FactomClient); err != nil {
			rpcErr, _ := err.(jrpc.Error)
			if rpcErr == newChainInProcessListErr {
				errLog.Fatalf("New chain %v is in process list. "+
					"Wait ~10 mins.\n", eb.ChainID)
			}
			if rpcErr == missingChainHeadErr {
				errLog.Fatalf("Chain %v does not exist. "+
					"First run `fat-cli issue chain`\n",
					eb.ChainID)
			}
			errLog.Fatal(err)
		}

		verifySK1Key(&sk1, stats.IssuerChainID)

		verifyECBalance(&ecEsAdr.EC, cost)
		vrbLog.Printf("Token Initialization Entry cost: %v EC", cost)
	}
	return nil
}

func verifySK1Key(sk1 *factom.SK1Key, idChainID *factom.Bytes32) {
	vrbLog.Printf("Fetching Identity Chain... %v", idChainID)
	var identity factom.Identity
	identity.ChainID = idChainID
	if err := identity.Get(FactomClient); err != nil {
		rpcErr, _ := err.(jrpc.Error)
		if rpcErr == newChainInProcessListErr {
			errLog.Fatalf("New identity chain %v is in process list. "+
				"Wait ~10 mins.\n", idChainID)
		}
		if rpcErr == missingChainHeadErr {
			errLog.Fatalf("Identity Chain %v does not exist.\n",
				idChainID)
		}
		errLog.Fatal(err)
	}
	vrbLog.Println("Verifying SK1 Key... ")
	if identity.ID1 != sk1.ID1Key() {
		errLog.Fatal("--sk1 is not the secret key corresponding to " +
			"the ID1Key declared in the Identity Chain.")
	}
}

func issueToken(_ *cobra.Command, _ []string) {
	if curl {
		if err := printCurl(Issuance.Entry.Entry, ecEsAdr.Es); err != nil {
			errLog.Fatal(err)
		}
		return
	}

	vrbLog.Println(
		"Submitting the Token Initialization Entry to the Factom blockchain...")
	txID, err := Issuance.ComposeCreate(FactomClient, ecEsAdr.Es)
	if err != nil {
		errLog.Fatal(err)
	}
	fmt.Printf("Token Initialization Entry Created: %v\n", Issuance.Hash)
	fmt.Printf("Chain ID: %v\n", Issuance.ChainID)
	fmt.Printf("Factom Tx ID:  %v\n", txID)
	return
}

type Type fat.Type

func (t *Type) Set(typeStr string) error {
	typeStr = strings.ToUpper(typeStr)
	switch typeStr {
	case "FAT0":
		typeStr = "FAT-0"
	case "FAT1":
		typeStr = "FAT-1"
	}
	return (*fat.Type)(t).Set(typeStr)
}

func (t Type) String() string {
	return fat.Type(t).String()
}
func (t Type) Type() string {
	return `<"FAT-0" | "FAT-1">`
}

type RawMessage json.RawMessage

func (r *RawMessage) Set(data string) error {
	if !json.Valid([]byte(data)) {
		return fmt.Errorf("invalid JSON")
	}
	*r = RawMessage(data)
	return nil
}

func (r RawMessage) String() string {
	return string(r)
}

func (RawMessage) Type() string {
	return "JSON"
}
