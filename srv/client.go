package srv

import (
	"fmt"
	"time"

	jrpc "github.com/AdamSLevy/jsonrpc2/v11"
)

// Client makes RPC requests to fatd's APIs. Client embeds a jsonrpc2.Client,
// and thus also the http.Client.  Use jsonrpc2.Client's BasicAuth settings to
// set up BasicAuth and http.Client's transport settings to configure TLS.
type Client struct {
	FatdServer string
	jrpc.Client
}

// Defaults for the factomd and factom-walletd endpoints.
const (
	FatdDefault = "http://localhost:8078"
)

// NewClient returns a pointer to a Client initialized with the default
// localhost endpoints for factomd and factom-walletd, and 15 second timeouts
// for each of the http.Clients.
func NewClient() *Client {
	c := &Client{FatdServer: FatdDefault}
	c.Timeout = 15 * time.Second
	return c
}

// Request makes a request to fatd's v1 API.
func (c *Client) Request(method string, params, result interface{}) error {
	url := c.FatdServer + "/v1"
	if c.DebugRequest {
		fmt.Println("fatd:", url)
	}
	return c.Client.Request(url, method, params, result)
}
