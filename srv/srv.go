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

package srv

import (
	"net/http"

	jrpc "github.com/AdamSLevy/jsonrpc2/v11"
	"github.com/Factom-Asset-Tokens/fatd/flag"
	_log "github.com/Factom-Asset-Tokens/fatd/log"
	"github.com/rs/cors"
)

var (
	log _log.Log
	srv http.Server

	APIVersion              = "1"
	FatdVersionHeaderKey    = http.CanonicalHeaderKey("Fatd-Version")
	FatdAPIVersionHeaderKey = http.CanonicalHeaderKey("Fatd-Api-Version")
)

func Start() {
	log = _log.New("srv")
	jrpc.DebugMethodFunc = true
	jrpcHandler := jrpc.HTTPRequestHandler(jrpcMethods)
	var handler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		header := w.Header()
		header.Add(FatdVersionHeaderKey, flag.Revision)
		header.Add(FatdAPIVersionHeaderKey, APIVersion)
		jrpcHandler(w, r)
	}

	// Set up server
	srvMux := http.NewServeMux()
	srvMux.Handle("/", handler)
	srvMux.Handle("/v1", handler)

	cors := cors.New(cors.Options{AllowedOrigins: []string{"*"}})

	srv = http.Server{Handler: cors.Handler(srvMux)}
	srv.Addr = flag.APIAddress
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Errorf("srv.ListenAndServe(): %v", err)
		}
	}()
}

func Stop() error {
	srv.Shutdown(nil)
	return nil
}
