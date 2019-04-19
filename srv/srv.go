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
)

const (
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
