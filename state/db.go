package state

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/flag"
	_log "github.com/Factom-Asset-Tokens/fatd/log"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

// Load state from all existing databases
func Load() error {
	// Try to create the database directory in case it doesn't already
	// exist.
	if err := os.Mkdir(flag.DBPath, 0755); err != nil && !os.IsExist(err) {
		return fmt.Errorf("os.Mkdir(%#v)", flag.DBPath)
	}

	// Scan through all files within the database directory. Ignore invalid
	// file names.
	files, err := ioutil.ReadDir(flag.DBPath)
	if err != nil {
		return fmt.Errorf("ioutil.ReadDir(%#v): %v", flag.DBPath, err)
	}
	for _, f := range files {
		fname := f.Name()
		chain := &Chain{ChainStatus: ChainStatusTracked}

		if chain.ID = fnameToChainID(fname); chain.ID == nil {
			continue
		}
		var err error
		if chain.DB, err = open(fname); err != nil {
			return err
		}
		if err := chain.loadMetadata(); err != nil {
			return err
		}
		if err := chain.loadIssuance(); err != nil {
			return err
		}
		chains.Set(chain)
	}

	return nil
}

func fnameToChainID(fname string) *factom.Bytes32 {
	if len(fname) != dbFileNameLen ||
		fname[dbFileNameLen-len(dbFileExtension):dbFileNameLen] != dbFileExtension {
		return nil
	}
	var chainID factom.Bytes32
	if err := json.Unmarshal(
		[]byte(fmt.Sprintf("%#v", fname[0:64])), &chainID); err != nil {
		return nil
	}
	return &chainID
}

var (
	log = _log.New("db")
)

const (
	dbDriver        = "sqlite3"
	dbFileExtension = ".sqlite3"
	dbFileNameLen   = len(factom.Bytes32{})*2 + len(dbFileExtension)
)

// open a database
func open(fname string) (*gorm.DB, error) {
	fpath := flag.DBPath + "/" + fname
	db, err := gorm.Open(dbDriver, fpath)
	if err != nil {
		return nil, err
	}
	// Ensure the db gets closed if there are any issues.
	defer func() {
		if err != nil {
			db.Close()
		}
	}()
	db.SetLogger(log)
	if err = autoMigrate(db); err != nil {
		return nil, err
	}
	return db, nil
}
func autoMigrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&entry{}).Error; err != nil {
		return fmt.Errorf("db.AutoMigrate(&Entry{}): %v", err)
	}
	if err := db.AutoMigrate(&address{}).Error; err != nil {
		return fmt.Errorf("db.AutoMigrate(&Address{}): %v", err)
	}
	if err := db.AutoMigrate(&metadata{}).Error; err != nil {
		return fmt.Errorf("db.AutoMigrate(&Metadata{}): %v", err)
	}
	return nil
}

func Close() {
	defer chains.Unlock()
	chains.Lock()
	for _, chain := range chains.m {
		if chain.DB == nil {
			continue
		}
		if err := chain.Close(); err != nil {
			log.Errorf(err.Error())
		}
	}
}

func GetSavedHeight() uint64 {
	return 0
}

func SaveHeight(height uint64) error {
	return nil
}
