package srv

import (
	"net/http"

	jrpc "github.com/AdamSLevy/jsonrpc2/v2"

	"bitbucket.org/canonical-ledgers/fatd/flag"
	_log "bitbucket.org/canonical-ledgers/fatd/log"
)

var (
	log _log.Log
	srv = http.Server{Handler: jrpc.HTTPRequestHandler}
)

func Start() error {
	log = _log.New("srv")

	// Register Methods
	jrpc.RegisterMethod("version", version)

	srv.Addr = flag.APIAddress

	go listen()

	return nil
}

func Stop() error {
	srv.Shutdown(nil)
	return nil
}

func listen() {
	err := srv.ListenAndServe()
	if err != http.ErrServerClosed {
		log.Errorf("srv.ListenAndServe(): %v", err)
	}
}
