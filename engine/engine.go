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
	log _log.Log
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

	log = _log.New("engine")

	if err := state.Load(); err != nil {
		log.Error(err)
		return
	}

	// Set up sync and factom heights...
	setSyncHeight(state.SavedHeight)
	updateFactomHeight()
	if syncHeight > factomHeight {
		// Probably indicates a bad StartScanHeight or we aren't
		// connected to Factom MainNet.
		log.Errorf("Saved height (%v) > Factom height (%v)",
			state.SavedHeight, factomHeight)
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

	log.Infof("Syncing from block %v to %v...", syncHeight+1, factomHeight)
	var synced bool
	scanTicker := time.NewTicker(scanInterval)
	for {
		if !synced && syncHeight == factomHeight {
			synced = true
			log.Infof("Synced to block %v.", syncHeight)
		}

		// Process all new DBlocks sequentially...
		for h := syncHeight + 1; h <= factomHeight; h++ {
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

			if flag.LogDebug && h%100 == 0 {
				log.Debugf("Synced to block %v...", h)
			}
		}

		if synced {
			// Wait until the next scan tick or we're told to stop.
			select {
			case <-scanTicker.C:
			case <-stop:
				return
			}
		}

		updateFactomHeight()
	}
}

var (
	syncHeight, factomHeight uint32
	heightMtx                = &sync.RWMutex{}
)

// GetSyncStatus is a threadsafe way to get the sync height and current Factom
// Blockchain height.
func GetSyncStatus() (sync, current uint32) {
	heightMtx.RLock()
	defer heightMtx.RUnlock()
	return syncHeight, factomHeight
}

func setSyncHeight(sync uint32) {
	heightMtx.Lock()
	defer heightMtx.Unlock()
	syncHeight = sync
}
func updateFactomHeight() {
	// Get the current Factom Blockchain height.
	var heights factom.Heights
	err := heights.Get(c)
	if err != nil {
		log.Errorf("factom.Heights.Get(c): %v", err)
		return
	}
	heightMtx.Lock()
	defer heightMtx.Unlock()
	factomHeight = heights.Entry
}
