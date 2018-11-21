package state

import (
	"fmt"
	"sync"
	"time"

	"github.com/Factom-Asset-Tokens/fatd/db"
	"github.com/Factom-Asset-Tokens/fatd/factom"
	_log "github.com/Factom-Asset-Tokens/fatd/log"
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

func Start() (chan error, error) {
	log = _log.New("state")

	if err := scanNewBlocks(); err != nil {
		return nil, fmt.Errorf("scanNewBlocks(): %v", err)
	}

	returnError = make(chan error, 1)
	stop = make(chan error)

	go engine()

	return returnError, nil
}

func Stop() error {
	if stop == nil {
		return fmt.Errorf("%#v", "Already not running")
	}
	close(stop)
	return nil
}

func errorStop(err error) {
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
	currentHeight := uint64(heights.EntryHeight)
	// Scan blocks from the last saved FBlockHeight up to but not including
	// the leader height
	for height := db.GetSavedHeight() + 1; height <= currentHeight; height++ {
		log.Debugf("Scanning block %v for FAT entries.", height)
		dblock := factom.DBlock{Height: height}
		if err := dblock.Get(); err != nil {
			return fmt.Errorf("DBlock%+v.Get(): %v", dblock, err)
		}

		wg := &sync.WaitGroup{}
		for i, _ := range dblock.EBlocks {
			wg.Add(1)
			go processEBlock(wg, dblock.EBlocks[i])
		}
		wg.Wait()

		if err := db.SaveHeight(height); err != nil {
			return fmt.Errorf("db.SaveHeight(%v): %v", height, err)
		}
	}

	return nil
}
