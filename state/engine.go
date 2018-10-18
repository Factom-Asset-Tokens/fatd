package state

import (
	"fmt"
	"sync"
	"time"

	"github.com/Factom-Asset-Tokens/fatd/db"
	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat0"
	_log "github.com/Factom-Asset-Tokens/fatd/log"
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
	currentHeight := uint64(heights.EntryHeight)
	// Scan blocks from the last saved FBlockHeight up to but not including
	// the leader height
	for height := db.GetSavedHeight() + 1; height <= currentHeight; height++ {
		log.Debugf("Scanning block %v for deposits.", height)
		dblock := &factom.DBlock{Height: height}
		if err := dblock.Get(); err != nil {
			return fmt.Errorf("DBlock%+v.Get(): %v", dblock, err)
		}

		wg := &sync.WaitGroup{}
		for i, _ := range dblock.EBlocks {
			wg.Add(1)
			go processEBlock(wg, &dblock.EBlocks[i])
		}
		wg.Wait()

		if err := db.SaveHeight(height); err != nil {
			return fmt.Errorf("db.SaveHeight(%v): %v", height, err)
		}
	}

	return nil
}

// Assumption: eb is not nil and has valid ChainID and KeyMR.
func processEBlock(wg *sync.WaitGroup, eb *factom.EBlock) {
	defer wg.Done()

	// Get the saved data for this chain.
	chain := chains.Get(eb.ChainID)

	// Skip ignored chains.
	if chain.Ignored() {
		return
	}

	// Load this Entry Block.
	if err := eb.Get(); err != nil {
		errorStop(fmt.Errorf("factom.GetEntryBlock(%#v): %v", eb, err))
		return
	}

	// Check whether this is a new chain.
	if !eb.First() {
		// Check whether this chain is tracked.
		if chain.Tracked() {
			if err := processEntries(chain, eb.Entries); err != nil {
				errorStop(err)
			}
			return
		}
		// Since the chian is not new and isn't already tracked, then
		// ignore this chain going forward.
		chains.Ignore(eb.ChainID)
		return
	}

	// Load first entry of new chain.
	if err := eb.Entries[0].Get(); err != nil {
		errorStop(fmt.Errorf("Entry%#v.Get: %v", eb.Entries[0], err))
		return
	}
	nameIDs := eb.Entries[0].ExtIDs

	// Filter out chains with NameIDs that don't match the fat0 pattern.
	if !fat0.ValidTokenNameIDs(nameIDs) {
		chains.Ignore(eb.ChainID)
		return
	}

	// Track this chain going forward.
	chains.Track(eb.ChainID, &fat0.Identity{ChainID: factom.NewBytes32(nameIDs[3])})

	// The first entry cannot be a valid Issuance entry, so discard it and
	// process the rest.
	if err := processEntries(chain, eb.Entries[1:]); err != nil {
		errorStop(err)
		return
	}
}

func processEntries(chain *Chain, es []factom.Entry) error {
	if !chain.Issuance.Populated() {
		return processIssuance(chain, es)
	}
	return processTransactions(chain, es)
}

// In general the following checks are ordered from cheapest to most expensive
// in terms of computation and memory.
func processIssuance(chain *Chain, es []factom.Entry) error {
	if len(es) == 0 {
		return nil
	}
	// The Identity may not have existed when this chain was first tracked.
	// Attempt to retrieve it.
	if err := chain.Identity.Get(); err != nil {
		return err
	}
	// If the Identity isn't yet populated then Issuance entries can't be
	// validated.
	if !chain.Identity.Populated() {
		return nil
	}
	// If these entries were created in a lower block height than the
	// Identity entry, then none of them can be a valid Issuance entry.
	if es[0].Height < chain.Identity.Height {
		return nil
	}

	for i, _ := range es {
		e := &es[i]
		// If this entry was created before the Identity entry then it
		// can't be valid.
		if e.Timestamp.Before(chain.Issuance.Timestamp.Time) {
			continue
		}
		// Get the data for the entry.
		if err := e.Get(); err != nil {
			return fmt.Errorf("factom.Entry%#v.Get(): %v", e, err)
		}
		issuance := &fat0.Issuance{Entry: fat0.Entry{Entry: e}}
		if !issuance.ValidExtIDs() {
			continue
		}
		if issuance.RCDHash() != *chain.Identity.IDKey {
			continue
		}
		if issuance.Unmarshal() != nil {
			continue
		}
		if !issuance.ValidData() {
			continue
		}
		if !issuance.VerifySignature() {
			continue
		}

		chains.Issue(issuance.ChainID, issuance)

		// Process remaining entries as transactions
		return processTransactions(chain, es[i+1:])
	}
	return nil
}

func processTransactions(chain *Chain, es []factom.Entry) error {
	for i, _ := range es {
		e := &es[i]
		if err := e.Get(); err != nil {
			return fmt.Errorf("factom.Entry%#v.Get(): %v", e, err)
		}
		transaction := &fat0.Transaction{Entry: fat0.Entry{Entry: e}}
		if transaction.Unmarshal() != nil {
			continue
		}
		if !transaction.ValidData() {
			continue
		}
		if !transaction.ValidExtIDs() {
			continue
		}
		if transaction.Coinbase {
			if transaction.RCDHash() != *chain.Identity.IDKey {
				continue
			}
		} else {
			if !transaction.VerifyRCDHashes() {
				continue
			}
		}
		if !transaction.VerifySignatures() {
			continue
		}
		if !chain.UniqueSignatures(transaction) {
			continue
		}
		if !chain.SufficientBalances(transaction) {
			continue
		}
		chain.Apply(transaction)
	}
	return nil
}
