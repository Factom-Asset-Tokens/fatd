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

	"github.com/Factom-Asset-Tokens/factom"
	"github.com/Factom-Asset-Tokens/fatd/internal/flag"
	_log "github.com/Factom-Asset-Tokens/fatd/internal/log"
	"golang.org/x/sync/errgroup"
)

var log _log.Log

// Start launches the main engine goroutine, which loads state and starts the
// worker goroutines. If stop is closed or if an error occurs, the engine will
// finish processing the current DBlock, cleanup and close state, all
// goroutines will exit, and done will be closed. If the done channel is closed
// before the stop channel is closed, an error occurred.
func Start(ctx context.Context, c *factom.Client) (done <-chan struct{}) {
	log = _log.New("pkg", "engine")

	// Verify Factom Blockchain NetworkID...
	log.Debug("Checking Factom DBlock height...")
	if err := updateFactomHeight(ctx, c); err != nil {
		if ctx.Err() == nil {
			log.Error(err)
		}
		return
	}
	log.Debug("Checking Factom NetworkID...")
	var dblock factom.DBlock
	dblock.Height = factomHeight
	if err := dblock.Get(ctx, c); err != nil {
		if ctx.Err() == nil {
			log.Errorf("dblock.Get(): %v", err)
		}
		return
	}
	if dblock.NetworkID != flag.NetworkID {
		log.Errorf("invalid Factom Blockchain NetworkID: %v, expected: %v",
			dblock.NetworkID, flag.NetworkID)
		return
	}

	// Add NetworkID subdirectory.
	flag.DBPath += fmt.Sprintf("%s%c",
		strings.ReplaceAll(flag.NetworkID.String(), " ", ""), os.PathSeparator)

	log.Debug("Loading state...")
	state, ctx, err := openState(ctx, c,
		flag.DBPath,
		flag.NetworkID,
		flag.Whitelist, flag.Blacklist,
		flag.SkipDBValidation, flag.RepairDB)
	if err != nil {
		log.Error(err)
		return
	}
	// Always close all chain databases if Start fails.
	defer func() {
		if done == nil {
			state.Close()
		}
	}()

	syncHeight = state.GetSync()

	if flag.IgnoreNewChains() {
		// We can assume that all chains are synced to their
		// chainheads, so we can start at the current height if we are
		// ignoring new chains.
		syncHeight = factomHeight
		if len(state.TrackedIDs()) == 0 {
			log.Error("no chains to track")
			return
		}
	} else if flag.StartScanHeight > -1 { // If -startscanheight was set...
		if flag.StartScanHeight > int32(factomHeight) {
			log.Warnf("-startscanheight %v > Factom height (%v), factomd may be syncing...",
				flag.StartScanHeight, factomHeight)
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
	_state = state // Set global _state
	go engine(ctx, c, _done, state)
	return _done
}

func engine(ctx context.Context, c *factom.Client,
	done chan struct{}, state State) {

	// Always close state and done on exit.
	defer func() {
		state.Close()
		log.Infof("Synced to block height %v.", syncHeight)
		close(done)
	}()

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
		runtime.GC()
		if !synced && syncHeight == factomHeight {
			synced = true
			log.Infof("DBlock scan complete to block %v.", syncHeight)
		}

		// Process all new DBlocks sequentially.
		for h := syncHeight + 1; h <= factomHeight; h++ {
			if err := ApplyDBlock(ctx, c, h, state); err != nil {
				if ctx.Err() == nil {
					log.Errorf("ApplyDBlock(): %v", err)
				}
				return
			}
		}

		if !flag.DisablePending || !synced {
			if err := ApplyPendingEntries(ctx, c, state); err != nil {
				if ctx.Err() == nil {
					log.Errorf("ApplyPendingEntries(): %v", err)
				}
				return
			}
		}

		if synced {
			// Wait until the next scan tick or we're told to stop.
			select {
			case <-scanTicker.C:
			case <-ctx.Done():
				return
			}
		}

		// Check the Factom blockchain height but log and retry if this
		// request fails.
		if err := updateFactomHeight(ctx, c); err != nil {
			log.Error(err)
			if flag.FactomScanRetries > -1 &&
				retries >= flag.FactomScanRetries {
				return
			}
			retries++
			log.Infof("Retrying in %v... (%v)",
				flag.FactomScanInterval, retries)
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

func updateFactomHeight(ctx context.Context, c *factom.Client) error {
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

func goN(ctx context.Context, n int,
	f func(context.Context) func() error) *errgroup.Group {

	g, ctx := errgroup.WithContext(ctx)
	for i := 0; i < n; i++ {
		g.Go(f(ctx))
	}

	return g
}

func ApplyDBlock(ctx context.Context, c *factom.Client, h uint32, state State) error {
	//log.Debugf("Syncing DBlock %v...", h)
	// Load next DBlock.
	dblock := factom.DBlock{Height: h}
	if err := dblock.Get(ctx, c); err != nil {
		return fmt.Errorf("%#v.Get(): %w", dblock, err)
	}

	n := runtime.NumCPU()
	if len(dblock.EBlocks) < n {
		n = len(dblock.EBlocks)
	}

	eblocks := make(chan factom.EBlock, n)
	g := goN(ctx, n, func(ctx context.Context) func() error {
		return func() error {
			for {
				select {
				case eb, ok := <-eblocks:
					if !ok {
						return nil
					}
					if err := state.ApplyEBlock(ctx,
						dblock.KeyMR, eb); err != nil {
						return fmt.Errorf("ChainID(%v): %w",
							eb.ChainID, err)
					}
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		}
	})
	for _, eb := range dblock.EBlocks {
		select {
		case eblocks <- eb:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	close(eblocks)
	if err := g.Wait(); err != nil {
		return fmt.Errorf("errgroup.Wait(): %w", err)
	}

	// DBlock completed so update the sync height for all
	// chains.
	setSyncHeight(h)
	if err := state.SetSync(ctx, h, dblock.KeyMR); err != nil {
		return fmt.Errorf("state.SetSync(): %w", err)
	}

	// Check that we haven't been told to stop.
	return ctx.Err()
}

func ApplyPendingEntries(ctx context.Context, c *factom.Client, state State) error {
	var pe factom.PendingEntries
	// Get and apply any pending entries.
	if err := pe.Get(ctx, c); err != nil {
		return fmt.Errorf("factom.PendingEntries.Get(): %w", err)
	}

	n := runtime.NumCPU()
	if len(pe) < n {
		n = len(pe)
	}

	entries := make(chan []factom.Entry, n)
	g := goN(ctx, runtime.NumCPU(), func(ctx context.Context) func() error {
		return func() error {
			for {
				select {
				case entries, ok := <-entries:
					if !ok {
						return nil
					}
					if err := state.ApplyPendingEntries(ctx,
						entries); err != nil {
						return fmt.Errorf("ChainID(%v): %w",
							entries[0].ChainID, err)
					}
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		}
	})

	for first := 0; first < len(pe); {
		if err := ctx.Err(); err != nil {
			return err
		}
		firstE := pe[first]

		// Unrevealed entries have no ChainID and are at the
		// end of the slice.
		if firstE.ChainID == nil {
			// No more revealed entries.
			break
		}

		// Grab any subsequent entries with this ChainID.
		var end int
		for end = first + 1; end < len(pe); end++ {
			chainID := pe[end].ChainID
			if chainID == nil || *chainID != *firstE.ChainID {
				break
			}
		}

		select {
		case entries <- pe[first:end]:
		case <-ctx.Done():
		}
		first = end
	}
	close(entries)

	return g.Wait()
}
