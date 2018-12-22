package state

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat0"
	"github.com/Factom-Asset-Tokens/fatd/flag"
	_log "github.com/Factom-Asset-Tokens/fatd/log"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

var (
	SavedHeight uint64 = 163180
	log         _log.Log
)

// Load state from all existing databases
func Load() error {
	log = _log.New("state")
	// Try to create the database directory in case it doesn't already
	// exist.
	if err := os.Mkdir(flag.DBPath, 0755); err != nil && !os.IsExist(err) {
		return fmt.Errorf("os.Mkdir(%#v)", flag.DBPath)
	}

	minHeight := uint64(math.MaxUint64)

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
		chains.Set(*chain)
		log.Debugf("loaded chain: %+v", chain)
		if chain.metadata.Height < minHeight {
			minHeight = chain.metadata.Height
		}
	}

	if minHeight < math.MaxUint64 {
		SavedHeight = minHeight
	}
	if flag.StartScanHeight > -1 {
		SavedHeight = uint64(flag.StartScanHeight)
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

// setupDB a database for a given token chain.
func (chain *Chain) setupDB() error {
	fname := fmt.Sprintf("%v%v", chain.ID, dbFileExtension)
	var err error
	if chain.DB, err = open(fname); err != nil {
		return err
	}
	// Ensure the db gets closed if there are any issues.
	defer func() {
		if err != nil {
			chain.Close()
		}
	}()
	if err := chain.Create(&chain.metadata).Error; err != nil {
		return err
	}
	return nil
}

func (chain *Chain) loadMetadata() error {
	var metadataTableCount int
	if err := chain.DB.Model(&metadata{}).Count(&metadataTableCount).Error; err != nil {
		return err
	}
	if metadataTableCount != 1 {
		return fmt.Errorf(`table "metadata" must have exactly one row`)
	}
	if err := chain.First(&chain.metadata).Error; err != nil {
		return err
	}
	if !fat0.ValidTokenNameIDs(fat0.NameIDs(chain.Token, chain.Issuer)) ||
		*chain.ID != fat0.ChainID(chain.Token, chain.Issuer) {
		return fmt.Errorf("corrupted metadata table for chain %v", chain.ID)
	}
	return nil
}

func (chain *Chain) loadIssuance() error {
	e := entry{}
	if err := chain.First(&e).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		return err
	}
	if !e.IsValid() {
		return fmt.Errorf("corrupted entry hash")
	}
	chain.Issuance = fat0.NewIssuance(e.Entry())
	if err := chain.Issuance.UnmarshalEntry(); err != nil {
		return err
	}
	chain.ChainStatus = ChainStatusIssued
	return nil
}

func (chain *Chain) saveIssuance() error {
	if chain.IsIssued() {
		return fmt.Errorf("already issued")
	}
	var entriesTableCount int
	if err := chain.DB.Model(&entry{}).Count(&entriesTableCount).Error; err != nil {
		return err
	}
	if entriesTableCount != 0 {
		return fmt.Errorf(`table "entries" must be empty prior to issuance`)
	}

	if err := chain.createEntry(chain.Issuance.Entry.Entry); err != nil {
		return err
	}
	chain.ChainStatus = ChainStatusIssued
	return nil
}
func (chain *Chain) saveMetadata() error {
	if err := chain.Save(&chain.metadata).Error; err != nil {
		return err
	}
	return nil
}
func (chain *Chain) createEntry(fe factom.Entry) error {
	if err := chain.Create(newEntry(fe)).Error; err != nil {
		return err
	}
	return nil
}

func (chain *Chain) saveHeight(height uint64) error {
	chain.metadata.Height = height
	if err := chain.saveMetadata(); err != nil {
		return err
	}
	return nil
}
