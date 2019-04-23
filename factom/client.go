package factom

import (
	"fmt"
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
	WalletdServer string
}

// Defaults for the factomd and factom-walletd endpoints.
const (
	FactomdDefault = "http://localhost:8088"
	WalletdDefault = "http://localhost:8089"
)

// NewClient returns a pointer to a Client initialized with the default
// localhost endpoints for factomd and factom-walletd, and 15 second timeouts
// for each of the http.Clients.
func NewClient() *Client {
	c := &Client{FactomdServer: FactomdDefault, WalletdServer: WalletdDefault}
	c.Factomd.Timeout = 20 * time.Second
	c.Walletd.Timeout = 10 * time.Second
	return c
}

// FactomdRequest makes a request to factomd's v2 API.
func (c *Client) FactomdRequest(method string, params, result interface{}) error {
	url := c.FactomdServer + "/v2"
	if c.Factomd.DebugRequest {
		fmt.Println("factomd:", url)
	}
	return c.Factomd.Request(url, method, params, result)
}

// WalletdRequest makes a request to factom-walletd's v2 API.
func (c *Client) WalletdRequest(method string, params, result interface{}) error {
	url := c.WalletdServer + "/v2"
	if c.Walletd.DebugRequest {
		fmt.Println("factom-walletd:", url)
	}
	return c.Walletd.Request(url, method, params, result)
}
