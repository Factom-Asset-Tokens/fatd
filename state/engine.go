package state

import (
	"fmt"
	"time"

	"bitbucket.org/canonical-ledgers/fatd/db"
	"bitbucket.org/canonical-ledgers/fatd/factom"
	_log "bitbucket.org/canonical-ledgers/fatd/log"
)

var (
	returnError chan error
	stop        chan error
	log         _log.Log
	scanTicker  = time.NewTicker(scanInterval)
)

const (
	scanInterval = 1 * time.Minute
)

func Start() chan error {
	log = _log.New("state")

	returnError = make(chan error, 1)
	stop = make(chan error)

	go engine()

	return returnError
}

func Stop() error {
	if stop == nil {
		return fmt.Errorf("%#v", "Already not running")
	}
	close(stop)
	stop = nil
	return nil
}

func errorStop(err error) {
	log.Debug("errorStop: %v", err)
	returnError <- err
	scanTicker.Stop()
}

func engine() {
	for {
		select {
		case <-scanTicker.C:
			err := scanNewBlocks()
			if err != nil {
				errorStop(fmt.Errorf("scanNewBlocks(): %v", err))
			}
		case <-stop:
			scanTicker.Stop()
			return
		}
	}
}

func scanNewBlocks() error {
	// Get the current leader's block height
	heights, err := factom.GetHeights()
	if err != nil {
		return fmt.Errorf("factom.GetHeights(): %v", err)
	}
	currentHeight := heights.EntryHeight
	// Scan blocks from the last saved FBlockHeight up to but not including
	// the leader height
	for height := db.GetSavedHeight(); height < currentHeight-1; height++ {
		log.Debugf("Scanning block %v for deposits.", height)
		if err := db.SaveHeight(height); err != nil {
			return fmt.Errorf("db.SaveHeight(%v): %v", height, err)
		}
	}

	return nil
}
