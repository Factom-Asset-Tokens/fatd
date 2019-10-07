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
	"fmt"
	"strconv"
	"strings"

	"github.com/Factom-Asset-Tokens/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat"
	"github.com/Factom-Asset-Tokens/fatd/fat/fat0"

	"github.com/posener/complete"
	"github.com/spf13/cobra"
)

var fat0Tx fat0.Transaction

// transactFAT0Cmd represents the FAT0 command
var transactFAT0Cmd = func() *cobra.Command {
	cmd := &cobra.Command{
		Use: `
fat0 --ecadr <EC | Es> --chainid <chain-id> [--metadata JSON]
        --input <FA | Fs>:<amount> [--input <FA | Fs>:<amount>]...
        --output <FA | Fs>:<amount> [--output <FA | Fs>:<amount>]...

  fat-cli transact fat0 --ecadr <EC | Es> --chainid <chain-id> [--metadata JSON]
        --sk1 <sk1-key>
        --output <FA | Fs>:<amount> [--output <FA | Fs>:<amount>]...
`[1:],
		Aliases: []string{"fat-0", "FAT0", "FAT-0"},
		Short:   "Send or distribute FAT-0 tokens",
		Long: `
Send or distribute FAT-0 tokens.

Generate, sign, and submit a FAT-0 transaction entry for the given --chainid.

Inputs and Outputs
        Both --input and --output expect an FA or Fs address, followed by ":",
        and then an <amount>.

        The <amount> must be a positive number.

        For example,
                FA3SjebEevRe964p4tQ6eieEvzi7puv9JWF3S3Wgw2v3WGKueL3R:150
                Fs2mGpZiHMwiEfe7kBD5ZYpXJsaxb3gUX258PJsAcNJ8GxFy8pBt:150

        For normal transactions, the sum of all of the --input <amount>s must
        equal the sum of the --output <amount>s.

See 'fat-cli transact --help' for more information about transactions.
`[1:],
		Run: func(_ *cobra.Command, _ []string) {},
	}
	transactCmd.AddCommand(cmd)
	transactCmplCmd.Sub["fat0"] = transactFAT0CmplCmd
	rootCmplCmd.Sub["help"].Sub["transact"].Sub["fat0"] = complete.Command{}

	flags := cmd.Flags()
	flags.VarPF((*AddressAmountMap)(&fat0Tx.Inputs), "input", "i", "").DefValue = ""
	flags.VarPF((*AddressAmountMap)(&fat0Tx.Outputs), "output", "o", "").DefValue = ""

	generateCmplFlags(cmd, transactFAT0CmplCmd.Flags)
	return cmd
}()

var PredictFAAddressesColon = PredictAppend(PredictFAAddresses, ":")

var transactFAT0CmplCmd = complete.Command{
	Flags: mergeFlags(apiCmplFlags, tokenCmplFlags,
		ecAdrCmplFlags, complete.Flags{
			"--input":  PredictFAAddressesColon,
			"-i":       PredictFAAddressesColon,
			"--output": PredictFAAddressesColon,
			"-o":       PredictFAAddressesColon,
		}),
}

var privateAddress = map[factom.FAAddress]factom.FsAddress{}
var addressValueStrMap = map[factom.FAAddress]string{}

type AddressAmountMap fat0.AddressAmountMap

func (m *AddressAmountMap) Set(adrAmtStr string) error {
	if *m == nil {
		*m = make(AddressAmountMap)
	}
	return m.set(adrAmtStr)
}
func (m AddressAmountMap) set(data string) error {
	// Split address from amount.
	strs := strings.Split(data, ":")
	if len(strs) != 2 {
		return fmt.Errorf("invalid format")
	}
	adrStr := strs[0]
	amountStr := strs[1]

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
		return fmt.Errorf("duplicate address")
	}

	// Parse amount
	amount, err := parsePositiveInt(amountStr)
	if err != nil {
		return fmt.Errorf("invalid amount: %w", err)
	}
	m[fa] = amount
	addressValueStrMap[fa] = amountStr

	return nil
}
func (m AddressAmountMap) String() string {
	return fmt.Sprintf("%v", fat0.AddressAmountMap(m))
}
func (AddressAmountMap) Type() string {
	return "<FA | Fs>:<amount>"
}

func parsePositiveInt(intStr string) (uint64, error) {
	if len(intStr) == 0 {
		return 0, fmt.Errorf("empty")
	}
	// Parse amount
	amount, err := strconv.ParseInt(intStr, 10, 64)
	if err != nil {
		return 0, err
	}
	if amount == 0 {
		return 0, fmt.Errorf("zero")
	}
	if amount < 0 {
		return 0, fmt.Errorf("negative")
	}
	return uint64(amount), nil
}
