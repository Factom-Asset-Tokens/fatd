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
	scanInterval = 2 * time.Second
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
	for height := db.GetSavedHeight(); height < currentHeight; height++ {
		log.Debugf("Scanning block %v for deposits.", height)
		dblock, err := factom.DBlockByHeight(height)
		if err != nil {
			return fmt.Errorf("factom.DBlockByHeight(%v): %v", height, err)
		}

		for _, eb := range dblock.EBlocks {
			if _, ok := ignoredChains[eb.ChainID]; ok {
				continue
			}
			if err := processEBlock(&eb); err != nil {
				return err
			}
		}

		if err := db.SaveHeight(height); err != nil {
			return fmt.Errorf("db.SaveHeight(%v): %v", height, err)
		}
	}

	return nil
}

var (
	ignoredChains = map[factom.Bytes32]bool{
		factom.Bytes32{31: 0x0a}: true,
		factom.Bytes32{31: 0x0c}: true,
		factom.Bytes32{31: 0x0f}: true,
	}
	trackedChains map[factom.Bytes32]bool
)

// Assumption: Chain is not yet ignored
func processEBlock(eb *factom.EBlock) error {
	// Check whether this is a new chain.
	if err := factom.GetEntryBlock(eb); err != nil {
		return fmt.Errorf("factom.GetEntryBlock(%#v): %v", eb, err)
	}
	if !eb.IsNewChain() {
		// Check whether we are already tracking this chain.
		if _, ok := trackedChains[eb.ChainID]; ok {
			// process chain
			return nil
		}
		// Otherwise we ignore this existing chain.
		ignoredChains[eb.ChainID] = true
		return nil
	}
	// New Chain!
	log.Debugf("EBlock%+v", eb)

	// Get first entry of chain.
	if err := factom.GetEntry(&eb.Entries[0]); err != nil {
		return fmt.Errorf("factom.GetEntry(%#v): %v", eb.Entries[0], err)
	}
	log.Debugf("Entry%+v", eb.Entries[0])

	// Check if ExtIDs of first entry match a FAT pattern
	if false {
		// If so track the chain for future entries.
		trackedChains[eb.ChainID] = true
		// Process any remaining entries
	}
	return nil
}
