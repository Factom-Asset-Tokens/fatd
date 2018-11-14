package factom

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"

	jrpc "github.com/AdamSLevy/jsonrpc2/v4"
)

// request makes a JSON RPC request with the given method and params, and then
// parses the response with the given result type. request only returns
// networking and unmarshaling errors. JSON RPC Errors are not returned. It is
// up to the caller to determine whether their result is properly populated
// after request returns. Since data will need to be marshaled into result, the
// result type should be passed as a pointer.
func request(method string, params interface{}, result interface{}) error {
	// Generate a random ID for this request.
	id := rand.Uint32()%200 + 500

	// Marshal the JSON RPC Request.
	reqBytes, err := json.Marshal(jrpc.NewRequest(method, id, params))
	if err != nil {
		return err
	}

	// Make the HTTP request.
	endpoint := "http://" + RpcConfig.FactomdServer + "/v2"
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(reqBytes))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	c := http.Client{Timeout: RpcConfig.FactomdTimeout}
	res, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusBadRequest {
		return fmt.Errorf("http: %v", res.Status)
	}

	// Read the HTTP response.
	resBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("ioutil.ReadAll(http.Response.Body): %v", err)
	}

	// Unmarshal the HTTP response into a JSON RPC response.
	resJrpc := jrpc.NewResponse(result)
	if err := json.Unmarshal(resBytes, resJrpc); err != nil {
		return fmt.Errorf("json.Unmarshal(): %v", err)
	}
	return nil
}
