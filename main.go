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
func _main() int {
	flag.Parse()
	// Attempt to run the completion program.
	if flag.Completion.Complete() {
		// The completion program ran, so just return.
		return 0
	}
	flag.Validate()

	log := getLogger()

	if err := db.Init(); err != nil {
		log.Fatal(err)
	}
	log.Info("DB connected.")

	stateErrCh := state.Start()
	log.Info("State machine started.")

	if err := srv.Start(); err != nil {
		log.Fatal(err)
	}
	log.Info("JSON RPC API server started.")

	log.Info("Factom Asset Token Daemon started.")
	defer log.Info("Factom Asset Token Daemon stopped.")

	// Set up interrupts channel.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	ret := 0
	select {
	case <-sig:
		log.Infof("SIGINT: Shutting down now.")
	case err := <-stateErrCh:
		log.Errorf("state: %v", err)
	}
	if err := state.Stop(); err != nil {
		log.Errorf("state.Stop(): %v", err)
		ret = 1
	}

	return ret
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
