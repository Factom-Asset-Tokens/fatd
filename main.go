package main

import (
	"os"
	"os/signal"

	"bitbucket.org/canonical-ledgers/fatd/db"
	"bitbucket.org/canonical-ledgers/fatd/factom"
	"bitbucket.org/canonical-ledgers/fatd/flag"
	"bitbucket.org/canonical-ledgers/fatd/log"
	"bitbucket.org/canonical-ledgers/fatd/srv"
	"bitbucket.org/canonical-ledgers/fatd/state"
)

func main() { os.Exit(_main()) }
func _main() (ret int) {
	flag.Parse()
	// Attempt to run the completion program.
	if flag.Completion.Complete() {
		// The completion program ran, so just return.
		return 0
	}
	flag.Validate()

	log := log.New("main")
	factom.Init()

	if err := db.Open(); err != nil {
		log.Errorf("db.Open(): %v", err)
		return 1
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Errorf("db.Close(): %v", err)
			ret = 1
			return
		}
		log.Info("DB connection closed.")
	}()
	log.Info("DB connection opened.")

	stateErrCh := state.Start()
	defer func() {
		if err := state.Stop(); err != nil {
			log.Errorf("state.Stop(): %v", err)
			ret = 1
			return
		}
		log.Info("State machine stopped.")
	}()
	log.Info("State machine started.")

	if err := srv.Start(); err != nil {
		log.Errorf("srv.Start(): %v", err)
		return 1
	}
	defer func() {
		if err := srv.Stop(); err != nil {
			log.Errorf("srv.Stop(): %v", err)
			ret = 1
			return
		}
		log.Info("JSON RPC API server stopped.")
	}()
	log.Info("JSON RPC API server started.")

	log.Info("Factom Asset Token Daemon started.")
	defer log.Info("Factom Asset Token Daemon stopped.")

	// Set up interrupts channel.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	select {
	case <-sig:
		log.Infof("SIGINT: Shutting down now.")
	case err := <-stateErrCh:
		log.Errorf("state: %v", err)
	}

	return
}
