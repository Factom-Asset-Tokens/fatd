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

	jrpc "github.com/AdamSLevy/jsonrpc2/v7"
	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/flag"
	"github.com/stretchr/testify/assert"
)

func cpResponse(res jrpc.Response) jrpc.Response {
	return jrpc.NewErrorResponse(res.Code, res.Message, res.Data)
}

// We make copies because the original is modified during the method call.
var tokenParamsRes = cpResponse(TokenParamsRes)
var tokenNotFoundRes = cpResponse(TokenNotFoundRes)
var transactionNotFoundRes = cpResponse(TransactionNotFoundRes)
var getTransactionParamsRes = cpResponse(GetTransactionParamsRes)
var getTransactionsParamsRes = cpResponse(GetTransactionsParamsRes)
var getBalanceParamsRes = cpResponse(GetBalanceParamsRes)
var getBalanceValidRes = jrpc.NewResponse(float64(0))

var tokenID = "invalid"

type Test struct {
	Params           interface{}
	Description      string
	ExpectedResponse jrpc.Response
}

var getIssuanceTests = []Test{{
	Params:           nil,
	Description:      "nil params",
	ExpectedResponse: tokenParamsRes,
}, {
	Params:           TokenParams{},
	Description:      "empty params",
	ExpectedResponse: tokenParamsRes,
}, {
	Params: struct {
		TokenParams
		NewField string
	}{TokenParams: TokenParams{ChainID: factom.NewBytes32(nil)}, NewField: "hello"},
	Description: "unknown field",
	ExpectedResponse: jrpc.NewInvalidParamsErrorResponse(
		`json: unknown field "NewField"`),
}, {
	Params:           TokenParams{ChainID: factom.NewBytes32(nil), TokenID: &tokenID},
	Description:      "chain id and token id",
	ExpectedResponse: tokenParamsRes,
}, {
	Params: TokenParams{ChainID: factom.NewBytes32(nil),
		IssuerChainID: factom.NewBytes32(nil)},
	Description:      "chain id and issuer chain id",
	ExpectedResponse: tokenParamsRes,
}, {
	Params: TokenParams{ChainID: factom.NewBytes32(nil),
		IssuerChainID: factom.NewBytes32(nil), TokenID: &tokenID},
	Description:      "chain id and token id and issuer chain id",
	ExpectedResponse: tokenParamsRes,
}, {
	Params: TokenParams{IssuerChainID: factom.NewBytes32(nil),
		TokenID: &tokenID},
	Description:      "token id and issuer chain id",
	ExpectedResponse: tokenNotFoundRes,
}, {
	Params:           TokenParams{ChainID: factom.NewBytes32(nil)},
	Description:      "chain id",
	ExpectedResponse: tokenNotFoundRes,
},
}

var getTransactionTests = []Test{{
	Params: TokenParams{ChainID: factom.NewBytes32(nil),
		IssuerChainID: factom.NewBytes32(nil)},
	Description:      "no hash",
	ExpectedResponse: getTransactionParamsRes,
}, {
	Params: GetTransactionParams{
		TokenParams: TokenParams{ChainID: factom.NewBytes32(nil)},
		Hash:        factom.NewBytes32(nil)},
	Description:      "tx not found",
	ExpectedResponse: transactionNotFoundRes,
},
}

var getTransactionsTests = []Test{{
	Params: GetTransactionsParams{
		TokenParams: TokenParams{ChainID: factom.NewBytes32(nil)},
		Hash:        factom.NewBytes32(nil), Start: new(uint)},
	Description:      "hash and start",
	ExpectedResponse: getTransactionsParamsRes,
}, {
	Params: GetTransactionsParams{
		TokenParams: TokenParams{ChainID: factom.NewBytes32(nil)},
		Hash:        factom.NewBytes32(nil)},
	Description:      "tx not found",
	ExpectedResponse: transactionNotFoundRes,
}, {
	Params: GetTransactionsParams{
		TokenParams: TokenParams{ChainID: factom.NewBytes32(nil)}, Limit: new(uint)},
	Description:      "zero limit",
	ExpectedResponse: getTransactionsParamsRes,
}, {
	Params: GetTransactionsParams{
		TokenParams: TokenParams{ChainID: factom.NewBytes32(nil)}},
	Description:      "tx not found",
	ExpectedResponse: transactionNotFoundRes,
},
}

var getBalanceTests = []Test{{
	Params: GetBalanceParams{
		TokenParams: TokenParams{ChainID: factom.NewBytes32(nil)}},
	Description:      "no address",
	ExpectedResponse: getBalanceParamsRes,
}, {
	Params: GetBalanceParams{
		Address: &factom.Address{}},
	Description:      "no chain",
	ExpectedResponse: getBalanceParamsRes,
}, {
	Params: GetBalanceParams{TokenParams: TokenParams{ChainID: factom.NewBytes32(nil)},
		Address: &factom.Address{}},
	Description:      "valid",
	ExpectedResponse: getBalanceValidRes,
},
}

var methodTests = map[string][]Test{
	"get-issuance":     getIssuanceTests,
	"get-transaction":  getTransactionTests,
	"get-transactions": getTransactionsTests,
	"get-balance":      getBalanceTests,
}

func TestMethods(t *testing.T) {
	flag.APIAddress = "localhost:18888"
	Start()
	for method, tests := range methodTests {
		t.Run(method, func(t *testing.T) {
			assert := assert.New(t)
			for _, test := range tests {
				res, err := request(method, test.Params, nil)
				assert.NoError(err)
				assert.NotNil(res.ID)
				res.ID = nil
				assert.Equal(test.ExpectedResponse.Error, res.Error,
					test.Description)
				assert.Equal(test.ExpectedResponse.Result, res.Result,
					test.Description)
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
