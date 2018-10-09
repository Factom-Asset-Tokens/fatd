package main

import (
	"os"
	"os/signal"

	"bitbucket.org/canonical-ledgers/fatd/db"
	"bitbucket.org/canonical-ledgers/fatd/flag"
	"bitbucket.org/canonical-ledgers/fatd/srv"
	"bitbucket.org/canonical-ledgers/fatd/state"

	"github.com/sirupsen/logrus"
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

	log := getLogger()

	if err := db.Open(); err != nil {
		log.Fatalf("db.Open(): %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Errorf("db.Close(): %v", err)
			ret = 1
		}
	}()
	log.Info("DB connected.")

	stateErrCh := state.Start()
	defer func() {
		if err := state.Stop(); err != nil {
			log.Errorf("state.Stop(): %v", err)
			ret = 1
		}
	}()
	log.Info("State machine started.")

	if err := srv.Start(); err != nil {
		log.Fatalf("srv.Start(): %v", err)
	}
	defer func() {
		if err := srv.Stop(); err != nil {
			log.Errorf("srv.Stop(): %v", err)
			ret = 1
		}
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

func getLogger() *logrus.Entry {
	log := logrus.New()
	log.Formatter = &logrus.TextFormatter{ForceColors: true,
		DisableTimestamp:       true,
		DisableLevelTruncation: true}
	if flag.LogDebug {
		log.SetLevel(logrus.DebugLevel)
	}
	return log.WithField("pkg", "main")

}
