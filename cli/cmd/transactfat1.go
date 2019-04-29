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
	"strings"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat"
	"github.com/Factom-Asset-Tokens/fatd/fat/fat1"
	"github.com/posener/complete"
	"github.com/spf13/cobra"
)

var fat1Tx fat1.Transaction

// transactFAT1Cmd represents the FAT1 command
var transactFAT1Cmd = func() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "fat1",
		Aliases: []string{"fat-1", "FAT1", "FAT-1"},
		Short:   "Send or distribute FAT-1 tokens",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
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
				return fmt.Errorf("invalid address: %v", err)
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
			return fmt.Errorf("invalid NFTokens: %v", err)
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
		return 0, fmt.Errorf("invalid NFTokenID: %v", err)
	}
	return fat1.NFTokenID(tknID), nil
}
