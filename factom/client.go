// MIT License
//
// Copyright 2018 Canonical Ledgers, LLC
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
// IN THE SOFTWARE.

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
