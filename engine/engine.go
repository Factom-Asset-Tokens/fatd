package engine

import (
	"fmt"
	"sync"
	"time"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	_log "github.com/Factom-Asset-Tokens/fatd/log"
	"github.com/Factom-Asset-Tokens/fatd/state"
)

var (
	returnError chan error
	stop        chan error
	log         _log.Log
	scanTicker  = time.NewTicker(scanInterval)
)

const (
	scanInterval = 30 * time.Second
)

func Start() (chan error, error) {
	if err := state.Load(); err != nil {
		return nil, err
	}

	log = _log.New("engine")
	returnError = make(chan error, 1)
	stop = make(chan error)

	go engine()

	return returnError, nil
}

func Stop() error {
	if stop == nil {
		return fmt.Errorf("Already not running")
	}
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
			errorStop(fmt.Errorf("scanNewBlocks(): %v", err))
		}
		select {
		case <-scanTicker.C:
			continue
		case <-stop:
			scanTicker.Stop()
			return
		}
	}
}

var synced bool

func scanNewBlocks() error {
	// Get the current leader's block height
	heights, err := factom.GetHeights()
	if err != nil {
		return fmt.Errorf("factom.GetHeights(): %v", err)
	}
	currentHeight := uint64(heights.EntryHeight)
	if !synced && currentHeight > state.SavedHeight {
		log.Infof("Syncing from block %v to %v...",
			state.SavedHeight, currentHeight)
	}
	// Scan blocks from the last saved block height up to but not including
	// the leader height
	for height := state.SavedHeight + 1; height <= currentHeight; height++ {
		log.Debugf("Scanning block %v for FAT entries.", height)
		dblock := factom.DBlock{Height: height}
		if err := dblock.Get(); err != nil {
			return fmt.Errorf("%#v.Get(): %v", dblock, err)
		}
		if !dblock.IsPopulated() {
			return fmt.Errorf("DBlock%+v.IsPopulated(): false")
		}

		wg := &sync.WaitGroup{}
		chainIDs := make(map[factom.Bytes32]struct{}, len(dblock.EBlocks))
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
					errorStop(err)
				}
			}()
		}
		wg.Wait()
		select {
		case <-stop:
			return nil
		default:
		}
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
