package factom

import (
	"time"

	jrpc "github.com/AdamSLevy/jsonrpc2/v11"
)

// Client makes RPC requests to factomd's and factom-walletd's APIs.  Client
// embeds two jsonrpc2.Clients, and thus also two http.Client, one for requests
// to factomd and one for requests to factom-walletd.  Use jsonrpc2.Client's
// BasicAuth settings to set up BasicAuth and http.Client's transport settings
// to configure TLS.
type Client struct {
	Factomd       jrpc.Client
	FactomdServer string
	Walletd       jrpc.Client
	WalletServer  string
}

// NewClient returns a pointer to a Client initialized with the default
// localhost endpoints for factomd and factom-walletd, and 15 second timeouts
// for each of the http.Clients.
func NewClient() *Client {
	c := &Client{FactomdServer: "localhost:8088", WalletServer: "localhost:8089"}
	c.Factomd.Timeout = 15 * time.Second
	c.Walletd.Timeout = 15 * time.Second
	return c
}

// FactomdRequest makes a request to factomd's v2 API.
func (c *Client) FactomdRequest(method string, params, result interface{}) error {
	url := c.FactomdServer + "/v2"
	return c.Factomd.Request(url, method, params, result)
}

// WalletdRequest makes a request to factom-walletd's v2 API.
func (c *Client) WalletdRequest(method string, params, result interface{}) error {
	url := c.WalletServer + "/v2"
	return c.Walletd.Request(url, method, params, result)
}
