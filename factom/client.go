package factom

import (
	"net/http"
	"time"

	jrpc "github.com/AdamSLevy/jsonrpc2/v11"
)

type Client struct {
	jrpc.Client
	FactomdServer string
	WalletServer  string
}

func NewClient() *Client {
	return &Client{Client: jrpc.Client{Client: http.Client{Timeout: 5 * time.Second}},
		FactomdServer: "localhost:8088", WalletServer: "localhost:8089"}
}

func (c *Client) FactomdRequest(method string, params, result interface{}) error {
	url := "http://" + c.FactomdServer + "/v2"
	return c.Request(url, method, params, result)
}
func (c *Client) WalletRequest(method string, params, result interface{}) error {
	url := "http://" + c.WalletServer + "/v2"
	return c.Request(url, method, params, result)
}
