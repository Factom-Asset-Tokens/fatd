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

func OpenAll() (cps []ConnPool, err error) {
	log = _log.New("db")
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
			for _, cp := range cps {
				cp.Close()
			}
			cps = nil
		}
	}()

	// Scan through all files within the database directory. Ignore invalid
	// file names.
	files, err := ioutil.ReadDir(flag.DBPath)
	if err != nil {
		return cps, fmt.Errorf("ioutil.ReadDir(%q): %v", flag.DBPath, err)
	}
	cps = make([]ConnPool, 0, len(files))
	for _, f := range files {
		fname := f.Name()
		chainID, err := fnameToChainID(fname)
		if err != nil {
			log.Debug(err)
			continue
		}
		log.Debugf("Loading chain: %v", chainID)
		cp, err := open(fname, chainDBSchema)
		if err != nil {
			return cps, err
		}
		cp.ChainID = chainID
		cps = append(cps, cp)
	}
	return cps, nil
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
