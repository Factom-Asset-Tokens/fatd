package state

import (
	"fmt"
	"time"

	_ "bitbucket.org/canonical-ledgers/fatd/db"
	_ "bitbucket.org/canonical-ledgers/fatd/factom"
	_ "bitbucket.org/canonical-ledgers/fatd/flag"
)

var (
	returnError chan error
	stop        chan error
)

const (
	scanInterval = 1 * time.Minute
)

func Start() chan error {
	setupLogger()

	returnError = make(chan error, 1)
	stop = make(chan error)

	go engine()

	return returnError
}

func Stop() error {
	if stop == nil {
		return fmt.Errorf("%#", "Already not running")
	}
	close(stop)
	stop = nil
	return nil
}

func errorStop(err error) {
	log.Debug("errorStop: %v", err)
	returnError <- err
}

func engine() {
	scanTick := time.Tick(scanInterval)
	for {
		select {
		case <-scanTick:
			err := scanNewBlocks()
			if err != nil {
				errorStop(fmt.Errorf("scanNewBlocks(): %v", err))
				return
			}
		case <-stop:
			log.Debug("stopped")
			return
		}
	}
}

func scanNewBlocks() error {
	return nil
}
