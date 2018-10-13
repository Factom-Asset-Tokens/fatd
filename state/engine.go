package state

import (
	"fmt"
	"sync"
	"time"

	"bitbucket.org/canonical-ledgers/fatd/db"
	"bitbucket.org/canonical-ledgers/fatd/factom"
	"bitbucket.org/canonical-ledgers/fatd/fat0"
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
	for height := db.GetSavedHeight() + 1; height <= currentHeight; height++ {
		log.Debugf("Scanning block %v for deposits.", height)
		dblock := &factom.DBlock{Height: height}
		if err := dblock.Get(); err != nil {
			return fmt.Errorf("factom.DBlockByHeight(%v): %v", height, err)
		}

		wg := &sync.WaitGroup{}
		ignored := 0
		for i, _ := range dblock.EBlocks {
			eb := &dblock.EBlocks[i]
			if chains.Get(&eb.ChainID).Ignored() {
				ignored++
				continue
			}
			wg.Add(1)
			go processEBlock(eb, wg)
		}
		log.Debugf("Ignored %v in block %v", ignored, height)
		wg.Wait()

		if err := db.SaveHeight(height); err != nil {
			return fmt.Errorf("db.SaveHeight(%v): %v", height, err)
		}
	}

	return nil
}

// Assumption: Chain is not yet ignored
func processEBlock(eb *factom.EBlock, wg *sync.WaitGroup) {
	defer wg.Done()
	// Check whether this is a new chain.
	if err := eb.Get(); err != nil {
		errorStop(fmt.Errorf("factom.GetEntryBlock(%#v): %v", eb, err))
		return
	}
	if !eb.IsNewChain() {
		// Check whether we are already tracking this chain.
		if status := chains.Get(&eb.ChainID); status.Tracked() {
			// process chain
			if status.Issued() {
				if err := processTransactionEntries(eb.Entries); err != nil {
					errorStop(err)
				}
				return
			}
			if err := processIssuanceEntries(eb.Entries); err != nil {
				errorStop(err)
			}
			return
		}
		// Otherwise we ignore this existing chain.
		chains.Ignore(&eb.ChainID)
		return
	}
	// New Chain!
	log.Debugf("EBlock%+v", eb)

	// Get first entry of chain.
	if err := eb.Entries[0].Get(); err != nil {
		errorStop(fmt.Errorf("Entry%#v.Get: %v", eb.Entries[0], err))
		return
	}
	log.Debugf("Entry%+v", eb.Entries[0])

	// Check if ExtIDs of first entry match a FAT pattern
	e := fat0.Entry{Entry: &eb.Entries[0]}
	if !e.ValidExtID() {
		// Otherwise we ignore this new chain.
		chains.Ignore(&eb.ChainID)
		return
	}
	// If ExtIDs match track the chain for future entries.
	chains.Track(&eb.ChainID)
	// Process any remaining for issuance. Discard first entry.
	if err := processIssuanceEntries(eb.Entries[1:]); err != nil {
		errorStop(err)
		return
	}
}

func processIssuanceEntries(es []factom.Entry) error {
	//for i, _ := range es {
	//	_ := &es[i]
	//}
	return nil
}

func processTransactionEntries(es []factom.Entry) error {
	//for i, _ := range es {
	//	_ := &es[i]
	//}
	return nil
}
