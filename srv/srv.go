package srv

import (
	"net/http"

	jrpc "github.com/AdamSLevy/jsonrpc2/v7"
	"github.com/Factom-Asset-Tokens/fatd/flag"
	_log "github.com/Factom-Asset-Tokens/fatd/log"
	"github.com/rs/cors"
)

var (
	log = _log.New("srv")
	srv = func() http.Server {
		jrpcHandler := jrpc.HTTPRequestHandler(jrpcMethods)
		// Set up server
		srvMux := http.NewServeMux()
		srvMux.Handle("/", jrpcHandler)
		srvMux.Handle("/v1", jrpcHandler)

		cors := cors.New(cors.Options{AllowedOrigins: []string{"*"}})

		return http.Server{Handler: cors.Handler(srvMux)}
	}()
)

func Start() {
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
