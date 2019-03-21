package factom

import (
	jrpc "github.com/AdamSLevy/jsonrpc2/v11"
)

type Client struct {
	jrpc.Client
	FactomdServer string
	WalletServer  string
}

func (c *Client) FactomdRequest(method string, params, result interface{}) error {
	url := "http://" + c.FactomdServer + "/v2"
	return c.Request(url, method, params, result)
}
func (c *Client) WalletRequest(method string, params, result interface{}) error {
	url := "http://" + c.WalletServer + "/v2"
	return c.Request(url, method, params, result)
}
