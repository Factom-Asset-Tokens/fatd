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

// Package engine manages syncing with the Factom blockchain and updating
// state. Start launches a number of goroutines: one to query for DBlocks
// sequentially, and a number of workers to concurrently process EBlocks within
// a DBlock and update state. If any runtime errors occur, engine finishes
// processing the current set of EBlocks and then exits. See Start for more
// details.
package engine

import (
	"sync"
	"time"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/flag"
	_log "github.com/Factom-Asset-Tokens/fatd/log"
	"github.com/Factom-Asset-Tokens/fatd/state"
)

var (
	log = _log.New("engine")
	c   = flag.FactomClient
)

const (
	scanInterval = 15 * time.Second
)

// Start launches the main engine goroutine, which loads state and starts the
// worker goroutines. If stop is closed or if an error occurs, the engine will
// finish processing the current DBlock, cleanup and close state, all
// goroutines will exit, and done will be closed. If the done channel is closed
// before the stop channel is closed, an error occurred.
func Start(stop <-chan struct{}) (done <-chan struct{}) {
	_done := make(chan struct{})
	go engine(stop, _done)
	return _done
}

func engine(stop <-chan struct{}, done chan struct{}) {
	// Ensure done is always closed exactly once.
	var once sync.Once
	exit := func() { once.Do(func() { close(done) }) }
	defer exit()

	if err := state.Load(); err != nil {
		log.Error(err)
		return
	}

	wg := &sync.WaitGroup{}
	eblocks := make(chan factom.EBlock)
	// Ensure all workers exit and state is closed when we exit.
	defer close(eblocks)
	defer state.Close()

	// Launch workers
	const numWorkers = 8
	for i := 0; i < numWorkers; i++ {
		go func() {
			for eb := range eblocks { // Read until close(eblocks)
				if err := state.Process(eb); err != nil {
					log.Errorf("ChainID(%v): %v", eb.ChainID, err)
					exit() // Tell engine() to exit.
				}
				wg.Done()
			}
		}()
	}

	// Scan for new DBlocks...
	setSyncHeight(state.SavedHeight)
	var synced, syncStatusPrinted bool
	scanTicker := time.NewTicker(scanInterval)
	for {
		// Get the current Factom Blockchain height.
		var heights factom.Heights
		err := heights.Get(c)
		if err != nil {
			log.Errorf("factom.Heights.Get(c): %v", err)
			return
		}
		setCurrentHeight(heights.Entry)

		// Print sync status...
		if !synced {
			if !syncStatusPrinted && syncHeight < currentHeight {
				syncStatusPrinted = true
				log.Infof("Syncing from block %v to %v...",
					state.SavedHeight+1, currentHeight)
			} else if currentHeight == state.SavedHeight {
				synced = true
				log.Infof("Synced to block %v.", currentHeight)
			} else {
				// Probably indicates a bad StartScanHeight or
				// we aren't connected to Factom MainNet.
				log.Errorf("Saved height (%v) > Factom height (%v)",
					state.SavedHeight, currentHeight)
				return
			}
		}

		// Process all new DBlocks sequentially...
		for h := syncHeight + 1; h <= currentHeight; h++ {
			log.Debugf("Scanning block %v for FAT entries.", h)
			// Get DBlock.
			var dblock factom.DBlock
			dblock.Header.Height = h
			if err := dblock.Get(c); err != nil {
				log.Errorf("%#v.Get(c): %v", dblock, err)
				return
			}

			// Queue all EBlocks for processing and wait.
			wg.Add(len(dblock.EBlocks))
			for _, eb := range dblock.EBlocks {
				eblocks <- eb
			}
			wg.Wait() // Wait for all EBlocks to be processed.

			// Check for process errors...
			select {
			case <-done:
				// We cannot consider this DBlock completed.
				return
			default:
			}

			// DBlock completed.
			setSyncHeight(h)
			if err := state.SaveHeight(h); err != nil {
				log.Errorf("state.SaveHeight(%v): %v", h, err)
				return
			}

			// Check that we haven't been told to stop.
			select {
			case <-stop:
				return
			default:
			}
		}

		if !synced {
			// Don't wait for the scan tick since the blockchain
			// may have advanced since we first checked.
			continue
		}

		// Wait until the next scan tick or we're told to stop.
		select {
		case <-scanTicker.C:
		case <-stop:
			return
		}
	}
}

var (
	syncHeight, currentHeight uint32
	heightMtx                 = &sync.RWMutex{}
)

// GetSyncStatus is a threadsafe way to get the sync height and current Factom
// Blockchain height.
func GetSyncStatus() (sync, current uint32) {
	heightMtx.RLock()
	defer heightMtx.RUnlock()
	return syncHeight, currentHeight
}

func setSyncHeight(sync uint32) {
	heightMtx.Lock()
	defer heightMtx.Unlock()
	syncHeight = sync
}
func setCurrentHeight(current uint32) {
	heightMtx.Lock()
	defer heightMtx.Unlock()
	currentHeight = current
}
