package engine

import (
	"fmt"
	"sync"
	"time"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	_log "github.com/Factom-Asset-Tokens/fatd/log"
	"github.com/Factom-Asset-Tokens/fatd/state"
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
	if err := state.Load(); err != nil {
		return nil, err
	}

	log = _log.New("engine")
	returnError = make(chan error, 1)
	stop = make(chan error)

	go engine()

	return returnError, nil
}

func Stop() error {
	if stop == nil {
		return fmt.Errorf("Already not running")
	}
	close(stop)
	state.Close()
	return nil
}

func errorStop(err error) {
	returnError <- err
	scanTicker.Stop()
}

func engine() {
	for {
		err := scanNewBlocks()
		if err != nil {
			errorStop(fmt.Errorf("scanNewBlocks(): %v", err))
		}
		select {
		case <-scanTicker.C:
			continue
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
	// Scan blocks from the last saved block height up to but not including
	// the leader height
	for height := state.GetSavedHeight() + 1; height <= currentHeight; height++ {
		log.Debugf("Scanning block %v for FAT entries.", height)
		dblock := factom.DBlock{Height: height}
		if err := dblock.Get(); err != nil {
			return fmt.Errorf("%#v.Get(): %v", dblock, err)
		}

		wg := &sync.WaitGroup{}
		chainIDs := make(map[factom.Bytes32]struct{}, len(dblock.EBlocks))
		for _, eb := range dblock.EBlocks {
			// There must never be a duplicate ChainID since chains
			// are processed concurrently. Since this is external
			// data we must validate it. This should never occur
			// and indicates a serious issue with the factomd API
			// endpoint we are talking to.
			_, ok := chainIDs[*eb.ChainID]
			if ok {
				return fmt.Errorf("duplicate ChainID in DBlock.EBlocks")
			}
			chainIDs[*eb.ChainID] = struct{}{}
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := state.ProcessEBlock(eb); err != nil {
					errorStop(err)
				}
			}()
		}
		wg.Wait()

		if err := state.SaveHeight(height); err != nil {
			return fmt.Errorf("state.SaveHeight(%v): %v", height, err)
		}
	}

	return nil
}
