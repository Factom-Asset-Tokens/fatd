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

package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/Factom-Asset-Tokens/fatd/engine"
	"github.com/Factom-Asset-Tokens/fatd/flag"
	"github.com/Factom-Asset-Tokens/fatd/log"
	"github.com/Factom-Asset-Tokens/fatd/srv"
)

func main() { os.Exit(_main()) }
func _main() (ret int) {
	// Completion uses some flags, so parse them first thing.
	flag.Parse()
	if flag.Completion.Complete() {
		// Invoked for the purposes of completion, so don't actually
		// run the daemon.
		return 0
	}
	flag.Validate()

	// Set up interrupts channel. We don't want to be interrupted during
	// initialization. If the signal is sent we will handle it later.
	ctx, cancel := context.WithCancel(context.Background())
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	go func() {
		<-sigint
		cancel()
	}()

	log := log.New("pkg", "main")
	log.Info("Fatd Version: ", flag.Revision)
	defer log.Info("Factom Asset Token Daemon stopped.")

	// Engine
	engineDone := engine.Start(ctx)
	if engineDone == nil {
		return 1
	}
	defer func() {
		<-engineDone // Wait for engine to stop.
		log.Info("State engine stopped.")
	}()
	log.Info("State engine started.")

	// Server
	srvDone := srv.Start(ctx)
	if srvDone == nil {
		return 1
	}
	defer func() {
		<-srvDone // Wait for server to stop.
		log.Info("JSON RPC API server stopped.")
	}()
	log.Info("JSON RPC API server started.")

	log.Info("Factom Asset Token Daemon started.")

	// Stop handling signals once we return.
	defer func() { signal.Reset(); close(sigint) }()

	select {
	case <-ctx.Done():
		log.Infof("SIGINT: Shutting down...")
		return 0
	case <-engineDone: // Closed if engine exits prematurely.
	case <-srvDone: // Closed if server exits prematurely.
	}
	return 1
}
