package state

import (
	"fmt"
	"sync"
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
	for height := db.GetSavedHeight() + 1; height <= currentHeight; height++ {
		log.Debugf("Scanning block %v for deposits.", height)
		dblock, err := factom.DBlockByHeight(height)
		if err != nil {
			return fmt.Errorf("factom.DBlockByHeight(%v): %v", height, err)
		}

		wg := &sync.WaitGroup{}
		ignored := 0
		for _, eb := range dblock.EBlocks {
			if ignore.get(&eb.ChainID) {
				ignored++
				continue
			}
			wg.Add(1)
			go processEBlock(&eb, wg)
		}
		log.Debugf("Ignored %v in block %v", ignored, height)
		wg.Wait()

		if err := db.SaveHeight(height); err != nil {
			return fmt.Errorf("db.SaveHeight(%v): %v", height, err)
		}
	}

	return nil
}

type chainMap struct {
	m map[factom.Bytes32]bool
	sync.RWMutex
}

func (c chainMap) set(b *factom.Bytes32) {
	defer c.Unlock()
	c.Lock()
	log.Debugf("Adding chain to ignore list")
	c.m[*b] = true
}

func (c chainMap) get(b *factom.Bytes32) bool {
	defer c.RUnlock()
	c.RLock()
	_, ok := c.m[*b]
	return ok
}

var (
	ignore = chainMap{m: map[factom.Bytes32]bool{
		factom.Bytes32{31: 0x0a}: true,
		factom.Bytes32{31: 0x0c}: true,
		factom.Bytes32{31: 0x0f}: true,
	}}

	track = chainMap{m: map[factom.Bytes32]bool{}}
)

// Assumption: Chain is not yet ignored
func processEBlock(eb *factom.EBlock, wg *sync.WaitGroup) {
	defer wg.Done()
	// Check whether this is a new chain.
	if err := factom.GetEntryBlock(eb); err != nil {
		errorStop(fmt.Errorf("factom.GetEntryBlock(%#v): %v", eb, err))
		return
	}
	if !eb.IsNewChain() {
		// Check whether we are already tracking this chain.
		if track.get(&eb.ChainID) {
			// process chain
			return
		}
		// Otherwise we ignore this existing chain.
		ignore.set(&eb.ChainID)
		return
	}
	// New Chain!
	log.Debugf("EBlock%+v", eb)

	// Get first entry of chain.
	if err := factom.GetEntry(&eb.Entries[0]); err != nil {
		errorStop(fmt.Errorf("factom.GetEntry(%#v): %v", eb.Entries[0], err))
		return
	}
	log.Debugf("Entry%+v", eb.Entries[0])

	// Check if ExtIDs of first entry match a FAT pattern
	ExtIDsDoNotMatch := true
	if ExtIDsDoNotMatch {
		// Otherwise we ignore this new chain.
		ignore.set(&eb.ChainID)
		return
	}
	// If ExtIDs match track the chain for future entries.
	track.set(&eb.ChainID)
	// Process any remaining entries
}
