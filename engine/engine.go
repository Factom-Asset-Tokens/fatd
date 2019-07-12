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
	"fmt"
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
	if err := updateFactomHeight(); err != nil {
		log.Error(err)
		return
	}

	// Guard against syncing against a network with an earlier blockheight.
	if syncHeight > factomHeight {
		log.Errorf("Saved height (%v) > Factom height (%v)",
			syncHeight, factomHeight)
		return
	}
	if flag.StartScanHeight > -1 { // If -startscanheight was set...
		if flag.StartScanHeight > int32(factomHeight) {
			log.Errorf("-startscanheight %v > Factom height (%v)",
				flag.StartScanHeight, factomHeight)
			return
		}
		// Warn if we are skipping blocks.
		if flag.StartScanHeight > int32(syncHeight)+1 {
			log.Warnf("-startscanheight %v skips over %v blocks from the last saved last saved block height which will very likely result in a corrupted database.",
				flag.StartScanHeight,
				flag.StartScanHeight-int32(syncHeight)-1)
		}
		// We start syncing at syncHeight+1, so subtract one. This
		// overflows for 0 but it's OK as long as we don't rely on the
		// value until the first scan loop.
		setSyncHeight(uint32(flag.StartScanHeight - 1))
	} else if syncHeight == 0 { // else if the syncHeight has not been set...
		const mainnetStart = 163180
		const testnetStart = 60000
		// This is a hacky, unreliable way to determine what network we
		// are on. This needs to be replaced with using the actually
		// Network ID.
		if factomHeight > mainnetStart {
			setSyncHeight(mainnetStart) // Set for mainnet
		} else if factomHeight > testnetStart {
			setSyncHeight(testnetStart) // Set for testnet
		} else {
			var zero uint32         // Avoid constant overflow compile error.
			setSyncHeight(zero - 1) // Start scan at 0.
		}
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
	var retries int64
	scanTicker := time.NewTicker(scanInterval)
	for {
		if !synced && syncHeight == factomHeight {
			synced = true
			log.Debugf("Synced to block %v...", syncHeight)
			log.Infof("Synced.")
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

		if err := updateFactomHeight(); err != nil {
			log.Error(err)
			if flag.FactomScanRetries > -1 &&
				retries >= flag.FactomScanRetries {
				return
			}
			retries++
			log.Infof("Retrying in %v... (%v)", scanInterval, retries)
		} else {
			retries = 0
		}
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
func updateFactomHeight() error {
	// Get the current Factom Blockchain height.
	var heights factom.Heights
	err := heights.Get(c)
	if err != nil {
		return fmt.Errorf("factom.Heights.Get(c): %v", err)
	}
	heightMtx.Lock()
	defer heightMtx.Unlock()
	factomHeight = heights.Entry
	return nil
}
