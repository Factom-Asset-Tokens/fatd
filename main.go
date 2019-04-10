package main

import (
	"os"
	"os/signal"

	"github.com/Factom-Asset-Tokens/fatd/engine"
	"github.com/Factom-Asset-Tokens/fatd/flag"
	"github.com/Factom-Asset-Tokens/fatd/log"
	"github.com/Factom-Asset-Tokens/fatd/srv"
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
	log.Info("Fatd Version: ", flag.Revision)

	engineErrCh, err := engine.Start()
	if err != nil {
		log.Errorf("engine.Start(): %v", err)
		return 1
	}
	defer func() {
		if err := engine.Stop(); err != nil {
			log.Errorf("engine.Stop(): %v", err)
			ret = 1
			return
		}
		log.Info("State engine stopped.")
	}()
	log.Info("State engine started.")

	srv.Start()
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
	case err := <-engineErrCh:
		log.Errorf("engine: %v", err)
	}

	return
}
