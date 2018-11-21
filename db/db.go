package db

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

type Metadata struct {
	gorm.Model

	Height uint64 `gorm:"default:161460"`
}

var (
	gDB    *gorm.DB
	tokens map[factom.Bytes32]*gorm.DB
	log    _log.Log

	metadata Metadata
)

const (
	dbDriver = "sqlite3"
)

func Open() error {
	log = _log.New("db")
	// Create the database directory, if it does not exist.
	if err := os.Mkdir(flag.DBPath, 0755); err != nil && !os.IsExist(err) {
		return fmt.Errorf("os.Mkdir(%#v)", flag.DBPath)
	}
	// Open the main fatd database file.
	dbFile := flag.DBPath + "/fatd.sqlite3"
	var err error
	if gDB, err = gorm.Open(dbDriver, dbFile); err != nil {
		return fmt.Errorf("db.Open(%#v, %#v): %v", dbDriver, dbFile, err)
	}
	gDB.SetLogger(log)
	// Run migrations
	if err = gDB.AutoMigrate(&metadata).Error; err != nil {
		return fmt.Errorf("gDB.AutoMigrate(&metadata): %v", err)
	}
	// Load metadata
	if err := gDB.FirstOrCreate(&metadata).Error; err != nil {
		return err
	}

	if flag.StartScanHeight >= 0 {
		SaveHeight(uint64(flag.StartScanHeight))
	}

	// Scan through all files within the database directory. Throw an error
	// for any unrecognised or invalidly named files.
	files, err := ioutil.ReadDir(flag.DBPath)
	if err != nil {
		return fmt.Errorf("ioutil.ReadDir(%#v): %v", flag.DBPath, err)
	}
	tokens = make(map[factom.Bytes32]*gorm.DB)
	for _, f := range files {
		fname := f.Name()
		if fname == "fatd.sqlite3" {
			continue
		}
		dbFile := flag.DBPath + "/" + fname
		if len(fname) == 64+8 && fname[64:64+8] == ".sqlite3" {
			var db *gorm.DB
			if db, err = gorm.Open(dbDriver, fname); err != nil {
				return fmt.Errorf("db.Open(%#v, %#v): %v",
					dbDriver, dbFile, err)
			}
			var chainID factom.Bytes32
			if err := json.Unmarshal([]byte(fmt.Sprintf("%#v", fname[0:64])),
				&chainID); err != nil {
				return fmt.Errorf("invalid file name: %#v", dbFile)
			}
			tokens[chainID] = db
			db.SetLogger(log)
			var md Metadata
			// Run migrations
			if err = db.AutoMigrate(&md).Error; err != nil {
				return fmt.Errorf("db{%v}.AutoMigrate(&md): %v",
					chainID, err)
			}
			// Load metadata
			if err := db.FirstOrCreate(&md).Error; err != nil {
				return err
			}
			// Pick the minimum height to start scanning from.
			if 0 < md.Height && md.Height < GetSavedHeight() {
				if err := SaveHeight(md.Height); err != nil {
					return err
				}
			}
		} else {
			return fmt.Errorf("invalid file name: %#v", dbFile)
		}
	}

	return nil
}

func Close() error {
	if gDB == nil {
		return fmt.Errorf("%v", "DB is not open.")
	}
	return gDB.Close()
}

func GetSavedHeight() uint64 {
	return metadata.Height
}

func SaveHeight(height uint64) error {
	metadata.Height = height
	if err := gDB.Model(&metadata).Update("height", height).Error; err != nil {
		return fmt.Errorf("gDB.Model(&metadata).Update(%#v, %v): %v",
			"height", height, err)
	}
	return nil
}

func GetChainState(c *factom.Bytes32) {

}
