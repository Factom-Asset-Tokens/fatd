package factom

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"

	_log "github.com/Factom-Asset-Tokens/fatd/log"

	jrpc "github.com/AdamSLevy/jsonrpc2/v4"
)

var log _log.Log

func Init() {
	log = _log.New("factom")
}

func request(method string, params interface{}, result interface{}) error {
	id := rand.Uint32()%200 + 500
	reqBytes, err := json.Marshal(jrpc.NewRequest(method, id, params))
	if err != nil {
		return fmt.Errorf("json.Marshal(jrpc.NewRequest(%#v, %v, %#v): %v",
			method, id, params, err)
	}
	endpoint := "http://" + RpcConfig.FactomdServer + "/v2"
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(reqBytes))
	if err != nil {
		return fmt.Errorf("http.NewRequest(%#v, %#v, %#v): %v",
			http.MethodPost, endpoint, reqBytes, err)
	}
	req.Header.Add("Content-Type", "application/json")

	c := http.Client{Timeout: RpcConfig.FactomdTimeout}
	res, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("http.Client%+v.Do(%#v): %v",
			c, req, err)
	}

	resBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("ioutil.ReadAll(http.Response.Body): %v", err)
	}

	resJrpc := jrpc.NewResponse(result)
	if err := json.Unmarshal(resBytes, resJrpc); err != nil {
		return fmt.Errorf("json.Unmarshal(, ): %v", err)
	}
	return nil
}
