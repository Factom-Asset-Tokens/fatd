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
	"fmt"
	"strings"

	"github.com/Factom-Asset-Tokens/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat"
	"github.com/Factom-Asset-Tokens/fatd/fat/fat1"

	"github.com/posener/complete"
	"github.com/spf13/cobra"
)

var fat1Tx fat1.Transaction

// transactFAT1Cmd represents the FAT1 command
var transactFAT1Cmd = func() *cobra.Command {
	cmd := &cobra.Command{
		Use: `
fat1 --ecadr <EC | Es> --chainid <chain-id> [--metadata JSON]
        --input <FA | Fs>:<nf-token-ids>, [--input <FA | Fs>:<nf-token-ids>]...
        --output <FA | Fs>:<nf-token-ids> [--output <FA | Fs>:<nf-token-ids>]...

  fat-cli transact fat1 --ecadr <EC | Es> --chainid <chain-id> [--metadata JSON]
        --sk1 <sk1-key>
        --output <FA | Fs>:<nf-token-ids> [--output <FA | Fs>:<nf-token-ids>]...
`[1:],
		Aliases: []string{"fat-1", "FAT1", "FAT-1"},
		Short:   "Send or distribute FAT-1 tokens",
		Long: `
Send or distribute FAT-1 tokens.

Generate, sign, and submit a FAT-1 transaction entry for the given --chainid.

Inputs and Outputs
        Both --input and --output expect an FA or Fs address, followed by ":",
        and then <nf-token-ids>.

        The <nf-token-ids> is a set NF Token IDs written as a comma separated
        list of IDs and ID ranges written as <min>-<max> (e.g. 1-100). The list
        must appear between "[" and "]". There may not be any duplicate NF
        Token IDs within a set, regardless of whether they are specified within
        a range or individually. The set does not need to be sorted.

        For example,
                FA3SjebEevRe964p4tQ6eieEvzi7puv9JWF3S3Wgw2v3WGKueL3R:[5,1,3,40-500,13]
                Fs2mGpZiHMwiEfe7kBD5ZYpXJsaxb3gUX258PJsAcNJ8GxFy8pBt:[1,3,5,13,40-500]

        For normal transactions, every NF Token ID used in an --input, must
        also be used in some --output.

See 'fat-cli transact --help' for more information about transactions.
`[1:],
		Run: func(_ *cobra.Command, _ []string) {},
	}
	transactCmd.AddCommand(cmd)
	transactCmplCmd.Sub["fat1"] = transactFAT1CmplCmd
	rootCmplCmd.Sub["help"].Sub["transact"].Sub["fat1"] = complete.Command{}

	flags := cmd.Flags()
	flags.VarPF((*AddressNFTokensMap)(&fat1Tx.Inputs), "input", "i", "").DefValue = ""
	flags.VarPF((*AddressNFTokensMap)(&fat1Tx.Outputs), "output", "o", "").DefValue = ""

	generateCmplFlags(cmd, transactFAT1CmplCmd.Flags)
	return cmd
}()

var PredictFAAddressesColonOpenBracket = PredictAppend(PredictFAAddresses, ":[")

var transactFAT1CmplCmd = complete.Command{
	Flags: mergeFlags(apiCmplFlags, tokenCmplFlags,
		ecAdrCmplFlags, complete.Flags{
			"--input":  PredictFAAddressesColonOpenBracket,
			"-i":       PredictFAAddressesColonOpenBracket,
			"--output": PredictFAAddressesColonOpenBracket,
			"-o":       PredictFAAddressesColonOpenBracket,
		}),
}

type AddressNFTokensMap fat1.AddressNFTokensMap

func (m *AddressNFTokensMap) Set(adrAmtStr string) error {
	if *m == nil {
		*m = make(AddressNFTokensMap)
	}
	return m.set(adrAmtStr)
}
func (m AddressNFTokensMap) set(data string) error {
	// Split address from amount.
	strs := strings.Split(data, ":")
	if len(strs) != 2 {
		return fmt.Errorf("invalid format")
	}
	adrStr := strs[0]
	tknIDsStr := strs[1]

	// Parse address, which could be FA or Fs or the keyword "coinbase" or
	// "burn"
	var fa factom.FAAddress
	var fs factom.FsAddress
	switch adrStr {
	case "coinbase", "burn":
		fa = fat.Coinbase()
	default:
		// Attempt to parse as FAAddress first
		if err := fa.Set(adrStr); err != nil {
			// Not FA, try FsAddress...
			if err := fs.Set(adrStr); err != nil {
				return fmt.Errorf("invalid address: %w", err)
			}
			fa = fs.FAAddress()
			if fa != fat.Coinbase() {
				// Save private addresses for future use.
				privateAddress[fa] = fs
			}
		}
	}
	if _, ok := m[fa]; ok {
		return fmt.Errorf("duplicate address: %v", fa)
	}

	// Parse NFTokens
	var tkns NFTokens
	if err := tkns.Set(tknIDsStr); err != nil {
		return err
	}

	m[fa] = fat1.NFTokens(tkns)
	addressValueStrMap[fa] = tknIDsStr
	return nil
}
func (m AddressNFTokensMap) String() string {
	return fmt.Sprintf("%v", fat1.AddressNFTokensMap(m))
}
func (AddressNFTokensMap) Type() string {
	return "<FA | Fs>:[<id>,<min>-<max>]"
}

type NFTokens fat1.NFTokens

func (tkns *NFTokens) Set(adrAmtStr string) error {
	if *tkns == nil {
		*tkns = make(NFTokens)
	}
	return tkns.set(adrAmtStr)
}
func (tkns NFTokens) set(data string) error {
	if len(data) < 2 || data[0] != '[' || data[len(data)-1] != ']' {
		return fmt.Errorf("invalid NFTokenIDs format")
	}
	data = data[1 : len(data)-1] // Trim '[' and ']'

	// Split NFTokenIDs or NFTokenIDRanges on ','
	tknIDStrs := strings.Split(data, ",")
	for _, tknIDStr := range tknIDStrs {
		var tknIDs fat1.NFTokensSetter
		tknRangeStrs := strings.Split(tknIDStr, "-")
		switch len(tknRangeStrs) {
		case 1:
			// Parse single NFToken
			tknID, err := parseNFTokenID(tknIDStr)
			if err != nil {
				return err
			}
			tknIDs = tknID
		case 2:
			minMax := make([]fat1.NFTokenID, 2)
			for i, tknIDStr := range tknRangeStrs {
				if len(tknIDStr) == 0 {
					return fmt.Errorf("invalid NFTokenIDRange format: %v",
						tknIDStr)
				}
				tknID, err := parseNFTokenID(tknIDStr)
				if err != nil {
					return err
				}
				minMax[i] = tknID
			}
			if minMax[0] > minMax[1] {
				return fmt.Errorf("invalid NFTokenIDRange: %v > %v",
					minMax[0], minMax[1])
			}
			tknIDs = fat1.NewNFTokenIDRange(minMax...)
		default:
			return fmt.Errorf("invalid NFTokenIDRange format: %v", tknIDStr)
		}
		// Set all NFTokenIDs to the NFTokens map.
		if err := fat1.NFTokens(tkns).Set(tknIDs); err != nil {
			return fmt.Errorf("invalid NFTokens: %w", err)
		}
	}
	return nil
}
func (tkns NFTokens) String() string {
	return fmt.Sprintf("%v", fat1.NFTokens(tkns))
}

func parseNFTokenID(tknIDStr string) (fat1.NFTokenID, error) {
	tknID, err := parsePositiveInt(tknIDStr)
	if err != nil {
		return 0, fmt.Errorf("invalid NFTokenID: %w", err)
	}
	return fat1.NFTokenID(tknID), nil
}
