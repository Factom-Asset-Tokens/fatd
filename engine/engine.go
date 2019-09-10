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
	"os"
	"sync"
	"time"

	"github.com/nightlyone/lockfile"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/flag"
	_log "github.com/Factom-Asset-Tokens/fatd/log"
)

var (
	log      _log.Log
	c        = flag.FactomClient
	lockFile lockfile.Lockfile
)

const (
	scanInterval = 30 * time.Second
)

// Start launches the main engine goroutine, which loads state and starts the
// worker goroutines. If stop is closed or if an error occurs, the engine will
// finish processing the current DBlock, cleanup and close state, all
// goroutines will exit, and done will be closed. If the done channel is closed
// before the stop channel is closed, an error occurred.
func Start(stop <-chan struct{}) (done <-chan struct{}) {
	log = _log.New("pkg", "engine")

	// Try to create the main and pending database directories, in case
	// they don't already exist.
	if err := createDir(flag.DBPath); err != nil {
		log.Error(err)
		return
	}

	// Try to create a lockfile
	lockFilePath := flag.DBPath + "db.lock"
	var err error
	lockFile, err = lockfile.New(lockFilePath)
	if err != nil {
		log.Error(err)
		return
	}
	if err = lockFile.TryLock(); err != nil {
		log.Errorf("Database in use by other process. %v", err)
		return
	}
	defer func() {
		if done == nil {
			if err := lockFile.Unlock(); err != nil {
				log.Errorf("lockFile.Unlock(): %v", err)
			}
		}
	}()

	// Verify Factom Blockchain NetworkID...
	if err := updateFactomHeight(); err != nil {
		log.Error(err)
		return
	}
	var dblock factom.DBlock
	dblock.Header.Height = factomHeight
	if err := dblock.Get(c); err != nil {
		log.Errorf("dblock.Get(): %v", err)
		return
	}
	if dblock.Header.NetworkID != flag.NetworkID {
		log.Errorf("invalid Factom Blockchain NetworkID: %v, expected: %v",
			dblock.Header.NetworkID, flag.NetworkID)
		return
	}

	// Load and sync all existing and whitelisted chains.
	log.Infof("Loading chain databases from %v...", flag.DBPath)
	if flag.SkipDBValidation {
		log.Warn("Skipping database validation...")
	}
	syncHeight, err = loadChains()
	if err != nil {
		log.Error(err)
		return
	}
	defer func() {
		if done == nil {
			Chains.Close()
		}
	}()

	if flag.IgnoreNewChains() {
		// We can assume that all chains are synced to their
		// chainheads, so we can start at the current height if we are
		// ignoring new chains.
		syncHeight = factomHeight
		if len(Chains.trackedIDs) == 0 {
			log.Error("no chains to track")
			return
		}
	} else if flag.StartScanHeight > -1 { // If -startscanheight was set...
		if flag.StartScanHeight > int32(factomHeight) {
			log.Errorf("-startscanheight %v > Factom height (%v)",
				flag.StartScanHeight, factomHeight)
			return
		}
		if !flag.IgnoreNewChains() &&
			flag.StartScanHeight > int32(syncHeight)+1 {
			log.Warnf("-startscanheight %v skips over %v blocks from the last saved last saved block height which will result in missing any new FAT Chains created in those blocks.",
				flag.StartScanHeight,
				flag.StartScanHeight-int32(syncHeight)-1)
		}
		// We start syncing at syncHeight+1, so subtract one. This
		// overflows for 0 but it's OK as long as we don't rely on the
		// value until the first scan loop.
		syncHeight = uint32(flag.StartScanHeight - 1)
	} else if syncHeight == 0 { // else if the syncHeight has not been set...
		switch flag.NetworkID {
		case factom.MainnetID():
			const mainnetStart = 163180
			syncHeight = mainnetStart // Set for mainnet
		case factom.TestnetID():
			const testnetStart = 60783
			syncHeight = testnetStart // Set for testnet
		default:
			var zero uint32       // Avoid constant overflow compile error.
			syncHeight = zero - 1 // Start scan at 0.
		}
	}

	_done := make(chan struct{})
	go engine(stop, _done)
	return _done
}

const numWorkers = 8

func engine(stop <-chan struct{}, done chan struct{}) {
	defer func() {
		Chains.Close()
		if err := lockFile.Unlock(); err != nil {
			log.Errorf("lockFile.Unlock(): %v", err)
		}
		close(done)
	}()

	// Launch workers
	eblocks := make(chan factom.EBlock)
	var once sync.Once
	stopWorkers := func() { once.Do(func() { close(eblocks) }) }
	defer stopWorkers()

	var dblock factom.DBlock
	wg := &sync.WaitGroup{}
	launchWorkers(numWorkers, func() {
		for eb := range eblocks { // Read until close(eblocks)
			if err := Process(dblock.KeyMR, eb); err != nil {
				log.Errorf("ChainID(%v): %v", eb.ChainID, err)
				stopWorkers() // Tell workers and engine() to exit.
			}
			wg.Done()
		}
	})

	if !flag.IgnoreNewChains() && syncHeight < factomHeight {
		log.Infof("Searching for new FAT chains from block %v to %v...",
			syncHeight+1, factomHeight)
	}
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
			dblock = factom.DBlock{}
			dblock.Header.Height = h
			if err := dblock.Get(c); err != nil {
				log.Errorf("%#v.Get(c): %v", dblock, err)
				return
			}

			// Queue all EBlocks for processing.
			wg.Add(len(dblock.EBlocks))
			for _, eb := range dblock.EBlocks {
				eblocks <- eb
			}
			wg.Wait() // Wait for all EBlocks to be processed.

			// Check for process errors...
			select {
			case <-eblocks:
				// One or more of the workers had an error and
				// closed the eblocks channel.
				// We cannot consider this DBlock completed, so
				// we do not update sync height for all chains.
				return
			default:
			}

			// DBlock completed.
			setSyncHeight(h)
			if err := Chains.setSync(h, dblock.KeyMR); err != nil {
				return
			}

			// Check that we haven't been told to stop.
			select {
			case <-stop:
				return
			default:
			}

			if flag.LogDebug && h%100 == 0 {
				log.Debugf("Synced to block Height: %v KeyMR: %v",
					h, dblock.KeyMR)
			}
		}

		if synced {
			var pe factom.PendingEntries
			if flag.DisablePending {
				goto WAIT
			}
			// Get and apply any pending entries
			if err := pe.Get(c); err != nil {
				log.Error(err)
				return
			}
			for i, j := 0, 0; i < len(pe); i = j {
				e := pe[i]
				if e.ChainID == nil {
					// No more revealed entries
					break
				}
				// Grab remaining entries with this chain ID.
				for j = i + 1; j < len(pe); j++ {
					chainID := pe[j].ChainID
					if chainID == nil || *chainID != *e.ChainID {
						break
					}
				}
				if err := ProcessPending(pe[i:j]...); err != nil {
					log.Error(err)
					return
				}
			}
		WAIT:
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

func launchWorkers(num int, job func()) {
	for i := 0; i < num; i++ {
		go job()
	}
}

func createDir(path string) error {
	if err := os.Mkdir(path, 0755); err != nil {
		if !os.IsExist(err) {
			return fmt.Errorf("os.Mkdir(%#v): %v", path, err)
		}
	}
	return nil
}
