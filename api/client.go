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

package api

import (
	"context"
	"fmt"
	"time"

	jrpc "github.com/AdamSLevy/jsonrpc2/v14"
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
func (c *Client) Request(ctx context.Context,
	method string, params, result interface{}) error {

	if c.DebugRequest {
		fmt.Println("fatd:", c.FatdServer)
	}
	return c.Client.Request(ctx, c.FatdServer, method, params, result)
}
