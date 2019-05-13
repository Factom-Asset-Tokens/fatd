// MIT License
//
// Copyright 2018 Canonical Ledgers, LLC
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
// IN THE SOFTWARE.

package engine

import (
	"fmt"
	"sync"
	"time"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/flag"
	_log "github.com/Factom-Asset-Tokens/fatd/log"
	"github.com/Factom-Asset-Tokens/fatd/state"
)

var (
	returnError chan error
	stop        chan error
	log         _log.Log
	scanTicker  *time.Ticker
	c           = flag.FactomClient
)

const (
	scanInterval = 30 * time.Second
)

func Start() (chan error, error) {
	if err := state.Load(); err != nil {
		return nil, err
	}

	setSyncHeight(state.SavedHeight)

	log = _log.New("engine")
	returnError = make(chan error, 1)
	stop = make(chan error)

	scanTicker = time.NewTicker(scanInterval)
	go engine()

	return returnError, nil
}

func Stop() error {
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
		if err := scanNewBlocks(); err != nil {
			go errorStop(fmt.Errorf("scanNewBlocks(): %v", err))
		}
		select {
		case <-scanTicker.C:
			continue
		case <-stop:
			return
		}
	}
}

var (
	synced                    bool
	syncHeight, currentHeight uint64
	heightMtx                 = &sync.RWMutex{}
)

func GetSyncStatus() (sync, current uint64) {
	heightMtx.RLock()
	defer heightMtx.RUnlock()
	return syncHeight, currentHeight
}

func setSyncHeight(sync uint64) {
	heightMtx.Lock()
	defer heightMtx.Unlock()
	syncHeight = sync
}
func setCurrentHeight(current uint64) {
	heightMtx.Lock()
	defer heightMtx.Unlock()
	currentHeight = current
}

func scanNewBlocks() error {
	// Get the current leader's block height
	var heights factom.Heights
	err := heights.Get(c)
	if err != nil {
		return fmt.Errorf("factom.Heights.Get(c): %v", err)
	}
	setCurrentHeight(uint64(heights.Entry))
	if !synced && currentHeight > state.SavedHeight {
		log.Infof("Syncing from block %v to %v...",
			state.SavedHeight, currentHeight)
	}
	// Scan blocks from the last saved block height up to but not including
	// the leader height
	for height := state.SavedHeight + 1; height <= currentHeight; height++ {
		log.Debugf("Scanning block %v for FAT entries.", height)
		dblock := factom.DBlock{Height: height}
		if err := dblock.Get(c); err != nil {
			return fmt.Errorf("%#v.Get(c): %v", dblock, err)
		}

		wg := &sync.WaitGroup{}
		chainIDs := make(map[factom.Bytes32]struct{}, len(dblock.EBlocks))
		processErrors := make(map[factom.Bytes32]error, len(dblock.EBlocks))
		processErrorsMutex := &sync.Mutex{}
		for _, eb := range dblock.EBlocks {
			// Because chains are processed concurrently, there
			// must never be a duplicate ChainID. Since the DBlock
			// is external data we must validate it. Factomd should
			// never return a DBlock with duplicate Chain IDs in
			// its EBlocks. If this happens it indicates a serious
			// issue with the factomd API endpoint we are talking
			// to.
			_, ok := chainIDs[*eb.ChainID]
			if ok {
				return fmt.Errorf("duplicate ChainID in DBlock.EBlocks")
			}
			chainIDs[*eb.ChainID] = struct{}{}

			// Skip ignored chains or EBlocks for heights earlier
			// than this chain's state.
			chain := state.Chains.Get(eb.ChainID)
			if chain.IsIgnored() || dblock.Height <= chain.Metadata.Height {
				continue
			}

			eb := eb
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := chain.Process(eb); err != nil {
					processErrorsMutex.Lock()
					defer processErrorsMutex.Unlock()
					processErrors[*chain.ID] = err
				}
			}()
		}
		wg.Wait()
		if len(processErrors) > 0 {
			for chainID, err := range processErrors {
				return fmt.Errorf("ChainID(%v): %v", chainID, err)
			}
		}
		setSyncHeight(height)
		if err := state.SaveHeight(height); err != nil {
			return err
		}
	}
	if !synced {
		log.Infof("Synced.")
		synced = true
	}

	return nil
}
