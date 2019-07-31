package db

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/flag"
	_log "github.com/Factom-Asset-Tokens/fatd/log"
)

var (
	log _log.Log
)

const (
	dbDriver        = "sqlite3"
	dbFileExtension = ".sqlite3"
	dbFileNameLen   = len(factom.Bytes32{})*2 + len(dbFileExtension)

	PoolSize = 10
)

func OpenAll() (chains []*Chain, err error) {
	log = _log.New("pkg", "db")
	// Try to create the database directory in case it doesn't already
	// exist.
	if err := os.Mkdir(flag.DBPath, 0755); err != nil {
		if !os.IsExist(err) {
			return nil, fmt.Errorf("os.Mkdir(%#v): %v", flag.DBPath, err)
		}
		log.Debug("Using existing database directory...")
	}

	defer func() {
		if err != nil {
			for _, chain := range chains {
				chain.Close()
			}
			chains = nil
		}
	}()

	// Scan through all files within the database directory. Ignore invalid
	// file names.
	files, err := ioutil.ReadDir(flag.DBPath)
	if err != nil {
		return nil, fmt.Errorf("ioutil.ReadDir(%q): %v", flag.DBPath, err)
	}
	chains = make([]*Chain, 0, len(files))
	for _, f := range files {
		fname := f.Name()
		chainID, err := fnameToChainID(fname)
		if err != nil {
			log.Debug(err)
			continue
		}
		log.Debugf("Loading chain: %v", chainID)
		chain, err := Open(fname)
		if err != nil {
			return nil, err
		}
		if *chainID != *chain.ID {
			return nil, fmt.Errorf("chain id does not match filename")
		}
		chains = append(chains, chain)
	}
	return chains, nil
}
func fnameToChainID(fname string) (*factom.Bytes32, error) {
	invalidFNameErr := fmt.Errorf("invalid filename: %v", fname)
	if len(fname) != dbFileNameLen ||
		fname[dbFileNameLen-len(dbFileExtension):dbFileNameLen] !=
			dbFileExtension {
		return nil, invalidFNameErr
	}
	chainID := factom.NewBytes32FromString(fname[0:64])
	if chainID == nil {
		return nil, invalidFNameErr
	}
	return chainID, nil
}
