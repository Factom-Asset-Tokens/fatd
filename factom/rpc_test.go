package factom_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	. "github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/stretchr/testify/assert"
)

type badParams int

func (b badParams) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("bad params")
}

func (b *badParams) UnmarshalJSON(_ []byte) error {
	return fmt.Errorf("bad params")
}

func newClient() *Client {
	return &Client{Client: http.Client{Timeout: 5 * time.Second},
		FactomdServer: "localhost:8088", WalletServer: "localhost:8089"}
}

func TestRequest(t *testing.T) {
	c := newClient()
	var b badParams
	assert := assert.New(t)
	assert.EqualError(c.WalletRequest("test", &b, nil),
		"json: error calling MarshalJSON for type jsonrpc2.Request: json: error calling MarshalJSON for type *factom_test.badParams: bad params")

	assert.EqualError(c.FactomdRequest("test", &b, nil),
		"json: error calling MarshalJSON for type jsonrpc2.Request: json: error calling MarshalJSON for type *factom_test.badParams: bad params")

	c.FactomdServer = "@#$%^"
	assert.EqualError(c.FactomdRequest("test", nil, nil),
		`parse http://@#$%^/v2: invalid URL escape "%^/"`)

	c.FactomdServer = "localhost"
	assert.EqualError(c.FactomdRequest("test", nil, nil),
		"Post http://localhost/v2: dial tcp [::1]:80: connect: connection refused")

	c.FactomdServer = "example.com/404please"
	assert.EqualError(c.FactomdRequest("test", nil, nil), "http: 404 Not Found")

	badServeURL := "localhost:10000"
	go http.ListenAndServe(badServeURL, http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Length", "1")
		}))
	c.FactomdServer = badServeURL
	assert.EqualError(c.FactomdRequest("properties", nil, &b),
		"ioutil.ReadAll(http.Response.Body): unexpected EOF")

	c.FactomdServer = "courtesy-node.factom.com"
	assert.EqualError(c.FactomdRequest("properties", nil, &b), "json.Unmarshal({\"jsonrpc\":\"2.0\",\"id\":594,\"result\":{\"factomdversion\":\"6.0.0\",\"factomdapiversion\":\"2.0\"}}): bad params")

	var result map[string]string
	assert.NoError(c.FactomdRequest("properties", nil, &result))
	version, ok := result["factomdversion"]
	assert.True(ok)
	assert.NotEmpty(version, "factomd version")
	version, ok = result["factomdapiversion"]
	assert.True(ok)
	assert.Equal(version, "2.0", "factomd api version")
}
