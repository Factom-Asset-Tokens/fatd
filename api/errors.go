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

package api

import jsonrpc2 "github.com/AdamSLevy/jsonrpc2/v14"

var (
	ErrorTokenNotFound = jsonrpc2.NewError(-32800, "Token Not Found",
		"token may be invalid, or not yet issued or tracked")
	ErrorTransactionNotFound = jsonrpc2.NewError(-32803, "Transaction Not Found",
		"no matching tx-id was found")
	ErrorInvalidTransaction = jsonrpc2.NewError(-32804, "Invalid Transaction", nil)
	ErrorTokenSyncing       = jsonrpc2.NewError(-32805, "Token Syncing",
		"token is in the process of syncing")
	ErrorNoEC = jsonrpc2.NewError(-32806, "No Entry Credits",
		"not configured with entry credits")
	ErrorPendingDisabled = jsonrpc2.NewError(-32807, "Pending Transactions Disabled",
		"fatd is not tracking pending transactions")
)
