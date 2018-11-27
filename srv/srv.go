package srv

import (
	"net/http"

	"github.com/Factom-Asset-Tokens/fatd/flag"
	_log "github.com/Factom-Asset-Tokens/fatd/log"
	"github.com/rs/cors"
)

var (
	log _log.Log
	srv = func() http.Server {
		// Set up server
		srvMux := http.NewServeMux()
		srvMux.Handle("/", jrpcHandler)
		srvMux.Handle("/v1", jrpcHandler)

		cors := cors.New(cors.Options{AllowedOrigins: []string{"*"}})

		return http.Server{Handler: cors.Handler(srvMux)}
	}()
)

func Start() error {
	log = _log.New("srv")

	// Launch server
	srv.Addr = flag.APIAddress
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
