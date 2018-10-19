package srv

import (
	"net/http"

	jrpc "github.com/AdamSLevy/jsonrpc2/v4"

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
	jrpc.RegisterMethod("get-transactions", getTransactions)
	jrpc.RegisterMethod("get-balance", getBalance)
	jrpc.RegisterMethod("get-stats", getStats)

	// Set up server
	srvMux := http.NewServeMux()
	srvMux.Handle("/", jrpc.HTTPRequestHandler)
	srvMux.Handle("/v1", Cors(jrpc.HTTPRequestHandler))
	srv.Handler = srvMux
	srv.Addr = flag.APIAddress

	// Launch server
	go listen()

	return nil
}

func Cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		next.ServeHTTP(w, r)
	})
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
