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
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/nightlyone/lockfile"

	"github.com/Factom-Asset-Tokens/factom"
	"github.com/Factom-Asset-Tokens/fatd/internal/flag"
	_log "github.com/Factom-Asset-Tokens/fatd/internal/log"
)

var (
	log      _log.Log
	c        = flag.FactomClient
	lockFile lockfile.Lockfile
)

func runIfNotDone(ctx context.Context, f func()) {
	select {
	case <-ctx.Done():
	default:
		f()
	}
}

const (
	scanInterval = 30 * time.Second
)

// Start launches the main engine goroutine, which loads state and starts the
// worker goroutines. If stop is closed or if an error occurs, the engine will
// finish processing the current DBlock, cleanup and close state, all
// goroutines will exit, and done will be closed. If the done channel is closed
// before the stop channel is closed, an error occurred.
func Start(ctx context.Context) (done <-chan struct{}) {
	log = _log.New("pkg", "engine")

	// Try to create the database directory.
	if err := os.Mkdir(flag.DBPath, 0755); err != nil {
		if !os.IsExist(err) {
			log.Errorf("os.Mkdir(%q): %v", flag.DBPath, err)
			return nil
		}
	}
	// Add NetworkID subdirectory.
	flag.DBPath += fmt.Sprintf("%s%c",
		strings.ReplaceAll(flag.NetworkID.String(), " ", ""), os.PathSeparator)
	if err := os.Mkdir(flag.DBPath, 0755); err != nil {
		if !os.IsExist(err) {
			log.Errorf("os.Mkdir(%q): %v", flag.DBPath, err)
			return nil
		}
	}

	// Try to create a lockfile
	lockFilePath := flag.DBPath + "db.lock"
	var err error
	lockFile, err = lockfile.New(lockFilePath)
	if err != nil {
		log.Errorf("lockfile.New(%q): %v", lockFilePath, err)
		return
	}
	if err = lockFile.TryLock(); err != nil {
		log.Errorf("lockFile.TryLock(): %v", err)
		return
	}
	// Always clean up the lockfile if Start fails.
	defer func() {
		if done == nil {
			if err := lockFile.Unlock(); err != nil {
				log.Errorf("lockFile.Unlock(): %v", err)
			}
		}
	}()

	// Verify Factom Blockchain NetworkID...
	if err := updateFactomHeight(ctx); err != nil {
		runIfNotDone(ctx, func() {
			log.Error(err)
		})
		return
	}
	var dblock factom.DBlock
	dblock.Height = factomHeight
	if err := dblock.Get(ctx, c); err != nil {
		runIfNotDone(ctx, func() {
			log.Errorf("dblock.Get(): %v", err)
		})
		return
	}
	if dblock.NetworkID != flag.NetworkID {
		log.Errorf("invalid Factom Blockchain NetworkID: %v, expected: %v",
			dblock.NetworkID, flag.NetworkID)
		return
	}

	// Load and sync all existing and whitelisted chains.
	log.Infof("Loading chain databases from %v...", flag.DBPath)
	if flag.SkipDBValidation {
		log.Warn("Skipping database validation...")
	}
	syncHeight, err = loadChains(ctx)
	if err != nil {
		runIfNotDone(ctx, func() {
			log.Error(err)
		})
		return
	}
	// Always close all chain databases if Start fails.
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
	go engine(ctx, _done)
	return _done
}

func engine(ctx context.Context, done chan struct{}) {
	// Always close all chains and remove lockfile on exit.
	defer func() {
		Chains.Close()
		if err := lockFile.Unlock(); err != nil {
			log.Errorf("lockFile.Unlock(): %v", err)
		}
		log.Infof("Synced to block height %v.", syncHeight)

		close(done)
	}()

	// eblocks is used to send new EBlocks to the workers for processing.
	eblocks := make(chan factom.EBlock)

	// eblocksWG is used to signal that all EBlocks for the current DBlock
	// are done being processed. This is reused each DBlock.
	var eblocksWG sync.WaitGroup

	// stopWorkers may be called multiple times by any worker or this
	// goroutine, but eblocks will only ever be closed once.
	var once sync.Once
	stopWorkers := func() { once.Do(func() { close(eblocks) }) }

	// Always stop all workers on exit.
	defer stopWorkers()

	// dblock is declared here and reused so that the workers below can
	// form a closure around it.
	var dblock factom.DBlock

	// Launch workers to process new EBlocks.
	numWorkers := runtime.NumCPU()
	for i := 0; i < numWorkers; i++ {
		go func() {
			for eb := range eblocks {
				if err := Process(ctx, dblock.KeyMR, eb); err != nil {
					runIfNotDone(ctx, func() {
						log.Errorf("ChainID(%v): %v",
							eb.ChainID, err)
					})
					// Tell workers and engine() to exit.
					stopWorkers()
				}
				eblocksWG.Done()
			}
		}()
	}

	if !flag.IgnoreNewChains() && syncHeight < factomHeight {
		log.Infof("Searching for new FAT chains from block %v to %v...",
			syncHeight+1, factomHeight)
	}

	// synced tracks whether we have completed our first sync.
	var synced bool

	// retries tracks the number of times we have had to retry querying for
	// the latest factom height.
	var retries int64

	// scanTicker kicks off a new scan.
	scanTicker := time.NewTicker(flag.FactomScanInterval)

	// Factom Blockchain Scan Loop
	for {
		if !synced && syncHeight == factomHeight {
			synced = true
			log.Infof("Synced to block %v.", syncHeight)
		}

		// Process all new DBlocks sequentially.
		for h := syncHeight + 1; h <= factomHeight; h++ {
			// Get DBlock.
			dblock = factom.DBlock{}
			dblock.Height = h
			if err := dblock.Get(ctx, c); err != nil {
				runIfNotDone(ctx, func() {
					log.Errorf("%#v.Get(): %v", dblock, err)
				})
				return
			}

			// Queue all EBlocks for processing.
			eblocksWG.Add(len(dblock.EBlocks))
			for _, eb := range dblock.EBlocks {
				eblocks <- eb
			}

			// Wait for all EBlocks to be processed.
			eblocksWG.Wait()

			// Check if any of the workers closed the eblocks
			// channel to indicate a Process() error.
			select {
			case <-eblocks:
				// One or more of the workers had an error and
				// closed the eblocks channel.
				// Since we cannot consider this DBlock
				// completed, we do not update sync height for
				// any chains.
				return
			default:
			}

			// DBlock completed so update the sync height for all
			// chains.
			setSyncHeight(h)
			if err := Chains.setSync(h, dblock.KeyMR); err != nil {
				runIfNotDone(ctx, func() {
					log.Errorf("Chains.setSync(): %v", err)
				})
				return
			}

			// Check that we haven't been told to stop.
			select {
			case <-ctx.Done():
				return
			default:
			}
		}

		// If we aren't yet synced, we want to immediately re-check the
		// Factom Height as the Blockchain may have advanced in the
		// time since we started the sync.
		if synced {
			if !flag.DisablePending {
				// Get and apply any pending entries.
				var pe factom.PendingEntries
				if err := pe.Get(ctx, c); err != nil {
					runIfNotDone(ctx, func() {
						log.Errorf(
							"factom.PendingEntries.Get(): %v",
							err)
					})
					return
				}

				for i, j := 0, 0; i < len(pe); i = j {
					e := pe[i]
					// Unrevealed entries have no ChainID
					// and are at the end of the slice.
					if e.ChainID == nil {
						// No more revealed entries.
						break
					}
					// Grab any subsequent entries with
					// this ChainID.
					for j = i + 1; j < len(pe); j++ {
						chainID := pe[j].ChainID
						if chainID == nil ||
							*chainID != *e.ChainID {
							break
						}
					}
					// Process all pending entries for this
					// chain.
					if err := ProcessPending(
						ctx, pe[i:j]...); err != nil {
						runIfNotDone(ctx, func() {
							log.Errorf("ChainID(%v): %v",
								e.ChainID, err)
						})
						return
					}
				}
			}

			// Wait until the next scan tick or we're told to stop.
			select {
			case <-scanTicker.C:
			case <-ctx.Done():
				return
			}
		}

		// Check the Factom blockchain height but log and retry if this
		// request fails.
		if err := updateFactomHeight(ctx); err != nil {
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

func updateFactomHeight(ctx context.Context) error {
	// Get the current Factom Blockchain height.
	var heights factom.Heights
	err := heights.Get(ctx, c)
	if err != nil {
		return fmt.Errorf("factom.Heights.Get(): %v", err)
	}
	heightMtx.Lock()
	defer heightMtx.Unlock()
	factomHeight = heights.Entry
	return nil
}
