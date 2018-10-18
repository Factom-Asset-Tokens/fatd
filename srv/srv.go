package srv

import (
	"net/http"

	jrpc "github.com/AdamSLevy/jsonrpc2/v3"

	"github.com/Factom-Asset-Tokens/fatd/flag"
	_log "github.com/Factom-Asset-Tokens/fatd/log"
)

var (
	log _log.Log
	srv http.Server
)

func Start() error {
	log = _log.New("srv")

	// Register JSON RPC Methods:
	jrpc.RegisterMethod("version", version)

	//Token methods (Mock data for now)
	jrpc.RegisterMethod("get-issuance", getIssuance)
	jrpc.RegisterMethod("get-transaction", getTransaction)
	jrpc.RegisterMethod("get-balance", getBalance)
	jrpc.RegisterMethod("get-stats", getStats)

	// Set up server
	srvMux := http.NewServeMux()
	srvMux.Handle("/", jrpc.HTTPRequestHandler)
	srvMux.Handle("/v1", jrpc.HTTPRequestHandler)
	srv.Handler = srvMux
	srv.Addr = flag.APIAddress

	// Launch server
	go listen()

	return nil
}

func Stop() error {
	srv.Shutdown(nil)
	return nil
}

func listen() {
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Errorf("srv.ListenAndServe(): %v", err)
	}
}
