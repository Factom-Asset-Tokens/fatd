package srv

import (
	"net/http"

	jrpc "github.com/AdamSLevy/jsonrpc2/v3"

	"bitbucket.org/canonical-ledgers/fatd/flag"
	_log "bitbucket.org/canonical-ledgers/fatd/log"
)

var (
	log _log.Log
	srv http.Server
)

func Start() error {
	log = _log.New("srv")

	// Register JSON RPC Methods
	jrpc.RegisterMethod("version", version)

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
