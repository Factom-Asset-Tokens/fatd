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

package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Factom-Asset-Tokens/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat"
	"github.com/Factom-Asset-Tokens/fatd/srv"

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
		Use: `
issue --ecadr <EC | Es> --sk1 <sk1-key>
        --identity <issuer-identity-chain-id> --tokenid <token-id>
        --type <"FAT-0" | "FAT-1"> --supply <supply> [--metadata <JSON>]`[1:],
		Short: "Issue a new token chain",
		Long: `
Issue a new FAT-0 or FAT-1 token chain.

Issuing a new FAT token chain involves submitting two Factom entries.

First, the Token Chain must be created with the correct Name IDs on the Factom
Blockchain. So both --tokenid and --identity are required and use of --chainid
is not allowed. Submitting the Chain Creation Entry will be skipped if it
already exists.

Second, the Token Initialization Entry must be added to the Token Chain. The
Token Initialization Entry must be signed by the SK1 key corresponding to the
ID1 key declared in the --identity chain. Both --type and --supply are
required. The --supply must be positive or -1 for an unlimited supply of
tokens.

Note that publishing a Token Initialization Entry is an immutable operation.
The protocol does not permit altering the Token Initialization Entry in any
way.

Sanity Checks
        Prior to composing the Chain Creation or Token Initialization Entry, a
        number of calls to fatd and factomd are made to ensure that the token
        can be issued. These checks are skipped if --force is used.

        - Skip Chain Creation Entry if already submitted.
        - The token has not already been issued.
        - The --identity chain exists.
        - The --sk1 key corresponds to the --identity's id1 key.
        - The --ecadr has enough ECs to pay for all entries.

Identity Chain
        FAT token chains may only be issued by an entity controlling the
        sk1/id1 key established by the Identity Chain pointed to by the FAT
        token chain. An Identity Chain and the associated keys can be created
        using the factom-identity-cli.

        https://github.com/PaulBernier/factom-identity-cli

Entry Credits
        Creating entries on the Factom blockchain costs Entry Credits. The full
        Token Issuance process normally costs 12 ECs. You must specify a funded
        Entry Credit address with --ecadr, which may be either a private Es
        address, or a pubilc EC address that can be fetched from
        factom-walletd.
`[1:],
		PreRunE: validateIssueFlags,
		Run:     issue,
		Args:    cobra.ExactArgs(0),
	}
	rootCmd.AddCommand(cmd)
	rootCmplCmd.Sub["issue"] = issueCmplCmd
	rootCmplCmd.Sub["help"].Sub["issue"] = complete.Command{Sub: complete.Commands{}}

	flags := cmd.Flags()
	flags.AddFlagSet(composeFlags)
	flags.VarPF((*Type)(&Issuance.Type), "type", "",
		"Token standard to use").DefValue = ""
	flags.VarPF(&sk1, "sk1", "", "Secret Identity Key 1 to sign entry").DefValue = ""
	flags.Int64Var(&Issuance.Supply, "supply", 0,
		"Max Token supply, use -1 for unlimited")
	flags.StringVar(&Issuance.Symbol, "symbol", "", "Optional abbreviated token symbol")
	flags.VarPF((*RawMessage)(&Issuance.Metadata), "metadata", "m",
		"JSON metadata to include in tx")

	generateCmplFlags(cmd, issueCmplCmd.Flags)
	// Don't complete these global flags as they are ignored by this
	// command.
	for _, flg := range []string{"-C", "--chainid"} {
		delete(issueCmplCmd.Flags, flg)
	}
	usage := cmd.UsageFunc()
	cmd.SetUsageFunc(func(cmd *cobra.Command) error {
		cmd.Flags().MarkHidden("chainid")
		return usage(cmd)
	})
	return cmd
}()

var ecAdrCmplFlags = complete.Flags{
	"--ecadr": PredictECAddresses,
	"-e":      PredictECAddresses,
}

var issueCmplCmd = complete.Command{
	Flags: mergeFlags(apiCmplFlags, tokenCmplFlags, ecAdrCmplFlags,
		complete.Flags{"--type": complete.PredictSet(fat.TypeFAT0.String(),
			fat.TypeFAT1.String())}),
}

var (
	missingChainHeadErr      = jrpc.Error{Code: -32009, Message: "Missing Chain Head"}
	newChainInProcessListErr = jrpc.Error{Message: "new chain in process list"}

	first       factom.Entry
	chainExists bool

	Issuance fat.Issuance
	sk1      factom.SK1Key
)

func validateIssueFlags(cmd *cobra.Command, args []string) error {
	if err := validateECAdrFlag(cmd, args); err != nil {
		return err
	}
	flags := cmd.Flags()
	if flags.Changed("chainid") {
		return fmt.Errorf("--chainid is not permitted, use --tokenid and --identity")
	}
	if !flags.Changed("tokenid") || !flags.Changed("identity") {
		return fmt.Errorf("--tokenid and --identity are required")
	}
	initChainID()

	for _, flg := range []string{"type", "supply", "sk1"} {
		if !flags.Changed(flg) {
			return fmt.Errorf("--" + flg + " is required")
		}
	}

	if Issuance.Supply == 0 {
		return fmt.Errorf("--supply may not be 0, use -1 for unlimited supply")
	}

	vrbLog.Println("Preparing Chain Creation Entry...")
	first.ExtIDs = NameIDs
	chainCost, err := first.Cost()
	if err != nil {
		errLog.Fatal(err)
	}

	vrbLog.Println("Preparing and signing Token Initialization Entry...")
	Issuance.ChainID = paramsToken.ChainID
	if err := Issuance.MarshalEntry(); err != nil {
		errLog.Fatal(err)
	}
	Issuance.Sign(sk1)
	initCost, err := Issuance.Cost()
	if err != nil {
		errLog.Fatal(err)
	}

	if !force {
		vrbLog.Println("Checking chain existence...")
		eb := factom.EBlock{ChainID: paramsToken.ChainID}
		if err := eb.GetChainHead(FactomClient); err != nil {
			rpcErr, _ := err.(jrpc.Error)
			if rpcErr != missingChainHeadErr &&
				rpcErr != newChainInProcessListErr {
				// If err was anything other than the missingChainHeadErr...
				errLog.Fatal(err)
			}
		} else {
			chainCost = 0
			chainExists = true
			vrbLog.Printf("Chain already exists.")
		}

		vrbLog.Println("Checking token chain status...")
		params := srv.ParamsToken{ChainID: paramsToken.ChainID}
		var stats srv.ResultGetStats
		if err := FATClient.Request("get-stats", params, &stats); err != nil {
			rpcErr, _ := err.(jrpc.Error)
			if rpcErr != *srv.ErrorTokenNotFound {
				errLog.Fatal(err)
			}
		} else {
			errLog.Fatal("Token is already initialized!")
		}

		verifySK1Key(&sk1, paramsToken.IssuerChainID)
		verifyECBalance(&ecEsAdr.EC, chainCost+initCost)
	}

	if !chainExists {
		vrbLog.Printf("New chain creation cost: %v EC", chainCost)
	}
	vrbLog.Printf("Token Initialization Entry cost: %v EC", initCost)
	vrbLog.Println()
	return nil
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

func verifySK1Key(sk1 *factom.SK1Key, idChainID *factom.Bytes32) {
	vrbLog.Printf("Fetching Identity Chain...")
	var identity factom.Identity
	identity.ChainID = idChainID
	if err := identity.Get(FactomClient); err != nil {
		rpcErr, _ := err.(jrpc.Error)
		if rpcErr == newChainInProcessListErr {
			errLog.Fatalf("New identity chain %v is in process list. "+
				"Wait ~10 mins.\n", idChainID)
		}
		if rpcErr == missingChainHeadErr {
			errLog.Fatalf("Identity Chain does not exist: %v", idChainID)
		}
		errLog.Fatal(err)
	}
	vrbLog.Println("Verifying SK1 Key... ")
	if identity.ID1 != sk1.ID1Key() {
		errLog.Fatal("--sk1 is not the secret key corresponding to " +
			"the ID1Key declared in the Identity Chain.")
	}
}

func issue(cmd *cobra.Command, args []string) {
	if !chainExists {
		issueChain(cmd, args)
	}
	issueToken(cmd, args)
}
func issueChain(_ *cobra.Command, _ []string) {
	if curl {
		if err := printCurl(first, ecEsAdr.Es); err != nil {
			errLog.Fatal(err)
		}
		return
	}

	vrbLog.Println("Submitting the Chain Creation Entry to the Factom blockchain...")
	txID, err := first.ComposeCreate(FactomClient, ecEsAdr.Es)
	if err != nil {
		errLog.Fatal(err)
	}
	fmt.Println("Chain Creation Entry Submitted")
	fmt.Println("Chain ID:    ", first.ChainID)
	fmt.Println("Entry Hash:  ", first.Hash)
	fmt.Println("Factom Tx ID:", txID)
	fmt.Println()
	return
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
	fmt.Println("Token Initialization Entry Submitted")
	fmt.Println("Entry Hash:  ", Issuance.Hash)
	fmt.Println("Factom Tx ID:", txID)
	return
}

func printCurl(entry factom.Entry, es factom.EsAddress) error {
	newChain := (entry.ChainID == nil)
	vrbLog.Println("Composing entry...")
	commit, reveal, _, err := entry.Compose(es)
	if err != nil {
		return err
	}

	commitMethod := "commit"
	revealMethod := "reveal"
	if newChain {
		commitMethod += "-chain"
		revealMethod += "-chain"
	} else {
		commitMethod += "-entry"
		revealMethod += "-entry"
	}

	vrbLog.Println("Curl commands:")
	fmt.Printf(`curl -X POST --data-binary '{"jsonrpc":"2.0","id":0,"method":%q,"params":{"message":%q}}' -H 'content-type:text/plain;' %v`,
		commitMethod, factom.Bytes(commit), FactomClient.FactomdServer)
	fmt.Println()

	fmt.Printf(`curl -X POST --data-binary '{"jsonrpc":"2.0","id": 0,"method":%q,"params":{"entry":%q}}' -H 'content-type:text/plain;' %v`,
		revealMethod, factom.Bytes(reveal), FactomClient.FactomdServer)
	fmt.Println()
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
	return "<EC | Es>"
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
