package state

import (
	"fmt"
	"time"

	"bitbucket.org/canonical-ledgers/fatd/db"
	"bitbucket.org/canonical-ledgers/fatd/factom"
	_log "bitbucket.org/canonical-ledgers/fatd/log"
)

var (
	returnError chan error
	stop        chan error
	log         _log.Log
)

const (
	scanInterval = 1 * time.Minute
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
		return fmt.Errorf("%#", "Already not running")
	}
	close(stop)
	stop = nil
	return nil
}

func errorStop(err error) {
	log.Debug("errorStop: %v", err)
	returnError <- err
}

func engine() {
	scanTick := time.Tick(scanInterval)
	for {
		select {
		case <-scanTick:
			err := scanNewBlocks()
			if err != nil {
				errorStop(fmt.Errorf("scanNewBlocks(): %v", err))
				return
			}
		case <-stop:
			log.Debug("stopped")
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
	for height := db.GetSavedHeight(); height < currentHeight; height++ {
		log.Debugf("Scanning block %v for deposits.", height)
		//// Get the transactions from this block
		//fctTransactions, err := getFCTTransactionsByHeight(height)
		//if err != nil {
		//	return fmt.Errorf("getFCTTransactionsByHeight(%v): %v", height, err)
		//}
		//// Scan the block's FCT transactions for deposits
		//if err := saveTransactions(
		//	scanFCTTransactionsForDeposits(fctTransactions)); err != nil {
		//	return fmt.Errorf("saveTransactions(txs) : %v", err)
		//}
		if err := db.SaveHeight(height); err != nil {
			return fmt.Errorf("db.SaveHeight(%v): %v", height, err)
		}
	}

	return nil
}
