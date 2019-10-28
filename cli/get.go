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
	"github.com/posener/complete"
	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = func() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "balance|chains|transactions",
		Long: `
Get balance, transaction, or issuance data about an existing FAT Chain.

The fatd API is used to lookup information about FAT chains. Thus fat-cli can
only return data about chains that the instance of fatd is tracking. The fatd
API must be trusted to ensure the security and validity of returned data.
`[1:],
	}
	rootCmd.AddCommand(cmd)
	rootCmplCmd.Sub["get"] = getCmplCmd
	rootCmplCmd.Sub["help"].Sub["get"] = complete.Command{Sub: complete.Commands{}}
	generateCmplFlags(cmd, getCmplCmd.Flags)
	return cmd
}()

var getCmplCmd = complete.Command{
	Flags: mergeFlags(apiCmplFlags, tokenCmplFlags),
	Sub:   complete.Commands{},
}
