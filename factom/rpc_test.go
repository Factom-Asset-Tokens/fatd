package factom

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type badParams int

func (b badParams) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("bad params")
}

func (b *badParams) UnmarshalJSON(_ []byte) error {
	return fmt.Errorf("bad params")
}

func TestRequest(t *testing.T) {
	var b badParams
	assert := assert.New(t)
	assert.EqualError(request("test", &b, nil),
		"json: error calling MarshalJSON for type *factom.badParams: bad params")

	RpcConfig.FactomdServer = "@#$%^"
	assert.EqualError(request("test", nil, nil),
		`parse http://@#$%^/v2: invalid URL escape "%^/"`)

	RpcConfig.FactomdServer = "localhost"
	assert.EqualError(request("test", nil, nil),
		"Post http://localhost/v2: dial tcp [::1]:80: connect: connection refused")

	RpcConfig.FactomdServer = "example.com/404please"
	assert.EqualError(request("test", nil, nil), "http: 404 Not Found")

	badServeURL := "localhost:10000"
	go http.ListenAndServe(badServeURL, http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Length", "1")
		}))
	RpcConfig.FactomdServer = badServeURL
	assert.EqualError(request("properties", nil, &b),
		"ioutil.ReadAll(http.Response.Body): unexpected EOF")

	RpcConfig.FactomdServer = "courtesy-node.factom.com"
	assert.EqualError(request("properties", nil, &b), "json.Unmarshal(): bad params")

	var result map[string]string
	assert.NoError(request("properties", nil, &result))
	version, ok := result["factomdversion"]
	assert.True(ok)
	assert.NotEmpty(version, "factomd version")
	version, ok = result["factomdapiversion"]
	assert.True(ok)
	assert.Equal(version, "2.0", "factomd api version")
}
