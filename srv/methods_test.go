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

package srv

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"testing"
	"time"

	jrpc "github.com/AdamSLevy/jsonrpc2/v11"
	"github.com/Factom-Asset-Tokens/factom"
	"github.com/Factom-Asset-Tokens/fatd/flag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//// We make copies because the original is modified during the method call.
//var tokenParamsRes = cpResponse(TokenParamsRes)
//var tokenNotFoundRes = cpResponse(TokenNotFoundRes)
//var transactionNotFoundRes = cpResponse(TransactionNotFoundRes)
//var getTransactionParamsRes = cpResponse(GetTransactionParamsRes)
//var getTransactionsParamsRes = cpResponse(GetTransactionsParamsRes)
//var getBalanceParamsRes = cpResponse(GetBalanceParamsRes)
var getBalanceValidRes = jrpc.NewResponse(float64(0))

var tokenID = "invalid"

type Test struct {
	Params      interface{}
	Description string
	Result      interface{}
	Error       interface{}
}

var getIssuanceTests = []Test{{
	Description: "nil params",
	Error:       TokenParamsError,
}, {
	Params:      TokenParams{},
	Description: "empty params",
	Error:       TokenParamsError,
}, {
	Params: struct {
		TokenParams
		NewField string
	}{TokenParams: TokenParams{ChainID: factom.NewBytes32(nil)}, NewField: "hello"},
	Description: "unknown field",
	Error:       jrpc.NewInvalidParamsError(`json: unknown field "NewField"`),
}, {
	Params:      TokenParams{ChainID: factom.NewBytes32(nil), TokenID: &tokenID},
	Description: "chain id and token id",
	Error:       TokenParamsError,
}, {
	Params: TokenParams{ChainID: factom.NewBytes32(nil),
		IssuerChainID: factom.NewBytes32(nil)},
	Description: "chain id and issuer chain id",
	Error:       TokenParamsError,
}, {
	Params: TokenParams{ChainID: factom.NewBytes32(nil),
		IssuerChainID: factom.NewBytes32(nil), TokenID: &tokenID},
	Description: "chain id and token id and issuer chain id",
	Error:       TokenParamsError,
}, {
	Params: TokenParams{IssuerChainID: factom.NewBytes32(nil),
		TokenID: &tokenID},
	Description: "token id and issuer chain id",
	Error:       TokenNotFoundError,
}, {
	Params:      TokenParams{ChainID: factom.NewBytes32(nil)},
	Description: "chain id",
	Error:       TokenNotFoundError,
},
}

var getTransactionTests = []Test{{
	Params: TokenParams{ChainID: factom.NewBytes32(nil),
		IssuerChainID: factom.NewBytes32(nil)},
	Description: "no hash",
	Error:       GetTransactionParamsError,
}, {
	Params: GetTransactionParams{
		TokenParams: TokenParams{ChainID: factom.NewBytes32(nil)},
		Hash:        factom.NewBytes32(nil)},
	Description: "tx not found",
	Error:       TransactionNotFoundError,
},
}

var getTransactionsTests = []Test{{
	Params: GetTransactionsParams{
		TokenParams: TokenParams{ChainID: factom.NewBytes32(nil)},
		Hash:        factom.NewBytes32(nil), Start: new(uint)},
	Description: "hash and start",
	Error:       GetTransactionsParamsError,
}, {
	Params: GetTransactionsParams{
		TokenParams: TokenParams{ChainID: factom.NewBytes32(nil)},
		Hash:        factom.NewBytes32(nil)},
	Description: "tx not found, with hash",
	Error:       TransactionNotFoundError,
}, {
	Params: GetTransactionsParams{
		TokenParams: TokenParams{ChainID: factom.NewBytes32(nil)}, Limit: new(uint)},
	Description: "zero limit",
	Error:       GetTransactionsParamsError,
}, {
	Params: GetTransactionsParams{
		TokenParams: TokenParams{ChainID: factom.NewBytes32(nil)}},
	Description: "tx not found",
	Error:       TransactionNotFoundError,
},
}

var getBalanceTests = []Test{{
	Params: GetBalanceParams{
		TokenParams: TokenParams{ChainID: factom.NewBytes32(nil)}},
	Description: "no address",
	Error:       GetBalanceParamsError,
}, {
	Params: GetBalanceParams{
		Address: &factom.Address{}},
	Description: "no chain",
	Error:       GetBalanceParamsError,
}, {
	Params: GetBalanceParams{TokenParams: TokenParams{ChainID: factom.NewBytes32(nil)},
		Address: &factom.Address{}},
	Description: "valid",
	Result:      0,
},
}

var getStatsTests = []Test{{
	Description: "no params",
	Error:       TokenParamsError,
}, {
	Params:      TokenParams{ChainID: factom.NewBytes32(nil)},
	Description: "valid",
	Result: struct {
		Supply                   int `json:"supply"`
		CirculatingSupply        int `json:"circulating-supply"`
		Transactions             int `json:"transactions"`
		IssuanceTimestamp        int `json:"issuance-timestamp"`
		LastTransactionTimestamp int `json:"last-transaction-timestamp"`
	}{},
}}

var NFTokenID = "test"

var getNFTokenTests = []Test{{
	Params: GetNFTokenParams{
		TokenParams: TokenParams{ChainID: factom.NewBytes32(nil)}},
	Description: "no nf token param",
	Error:       GetNFTokenParamsError,
}, {
	Params: GetNFTokenParams{NonFungibleTokenID: &NFTokenID,
		TokenParams: TokenParams{ChainID: factom.NewBytes32(nil)}},
	Description: "valid",
	Error:       TokenNotFoundError,
}}

var sendTransactionTests = []Test{{
	Description: "no params",
	Error:       SendTransactionParamsError,
}, {
	Params: SendTransactionParams{Content: factom.Bytes{0x00},
		ExtIDs:      []factom.Bytes{{0x00}},
		TokenParams: TokenParams{ChainID: factom.NewBytes32(nil)}},
	Description: "invalid token",
	Error:       TokenNotFoundError,
}}

var getDaemonTokensTests = []Test{{
	Description: "no params",
	Result: []struct {
		TokenID  string          `json:"token-id"`
		IssuerID *factom.Bytes32 `json:"issuer-id"`
		ChainID  *factom.Bytes32 `json:"chain-id"`
	}{{}},
}, {
	Params:      TokenParams{ChainID: factom.NewBytes32(nil)},
	Description: "valid",
	Error:       NoParamsError,
}}

var getDaemonPropertiesTests = []Test{{
	Params:      []int{0},
	Description: "invalid params",
	Error:       NoParamsError,
}, {
	Description: "no params",
	Result: struct {
		FatdVersion string `json:"fatd-version"`
		APIVersion  string `json:"api-version"`
	}{FatdVersion: "0.0.0", APIVersion: "v0"},
}}

var methodTests = map[string][]Test{
	"get-issuance":          getIssuanceTests,
	"get-transaction":       getTransactionTests,
	"get-transactions":      getTransactionsTests,
	"get-balance":           getBalanceTests,
	"get-stats":             getStatsTests,
	"get-nf-token":          getNFTokenTests,
	"send-transaction":      sendTransactionTests,
	"get-daemon-tokens":     getDaemonTokensTests,
	"get-daemon-properties": getDaemonPropertiesTests,
}

func TestMethods(t *testing.T) {
	flag.APIAddress = "localhost:18888"
	Start()
	for method, tests := range methodTests {
		t.Run(method, func(t *testing.T) {
			for _, test := range tests {
				t.Run(test.Description, func(t *testing.T) {
					assert := assert.New(t)
					require := require.New(t)
					res, err := request(method,
						test.Params, &json.RawMessage{})
					assert.NoError(err)
					assert.NotNil(res.ID)
					if test.Result != nil {
						data, err := json.Marshal(test.Result)
						require.NoError(err)
						result := res.Result.(*json.RawMessage)
						require.NotEmpty(result)
						assert.JSONEq(string(data), string(*result),
							"Result")
					} else {
						require.NotNil(res.Error)
						assert.Equal(test.Error, *res.Error, "Error")
					}
				})
			}
		})
	}
	Stop()
}

func request(method string, params interface{}, result interface{}) (jrpc.Response, error) {
	// Generate a random ID for this request.
	id := rand.Uint32()%200 + 500

	// Marshal the JSON RPC Request.
	reqBytes, err := json.Marshal(jrpc.NewRequest(method, id, params))
	if err != nil {
		return jrpc.Response{}, err
	}

	// Make the HTTP request.
	endpoint := "http://" + flag.APIAddress
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(reqBytes))
	if err != nil {
		return jrpc.Response{}, err
	}
	req.Header.Add("Content-Type", "application/json")
	c := http.Client{Timeout: 2 * time.Second}
	res, err := c.Do(req)
	if err != nil {
		return jrpc.Response{}, err
	}
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusBadRequest {
		return jrpc.Response{}, fmt.Errorf("http: %v", res.Status)
	}

	// Read the HTTP response.
	resBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return jrpc.Response{}, fmt.Errorf("ioutil.ReadAll(http.Response.Body): %v", err)
	}

	// Unmarshal the HTTP response into a JSON RPC response.
	resJrpc := jrpc.NewResponse(result)
	if err := json.Unmarshal(resBytes, &resJrpc); err != nil {
		return jrpc.Response{}, fmt.Errorf("json.Unmarshal(): %v", err)
	}
	return resJrpc, nil
}
