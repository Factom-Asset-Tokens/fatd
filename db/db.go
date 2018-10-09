package db

import (
	"fmt"

	"bitbucket.org/canonical-ledgers/fatd/flag"
	_log "bitbucket.org/canonical-ledgers/fatd/log"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

var (
	db  *gorm.DB
	log _log.Log
)

const (
	dbDriver = "sqlite3"
)

func Open() error {
	log = _log.New("db")
	var err error
	if db, err = gorm.Open(dbDriver, flag.DBFile); err != nil {
		return fmt.Errorf("db.Open(%#v, %#v): %v", dbDriver, flag.DBFile, err)
	}

	// Run migrations
	if err = db.AutoMigrate(&metadata{}).Error; err != nil {
		return fmt.Errorf("db.AutoMigrate(&metadata{}): %v", err)
	}

	return nil
}

func Close() error {
	if db == nil {
		return fmt.Errorf("%v", "DB is not open.")
	}
	return db.Close()
}

func GetSavedHeight() int64 {
	return 0
}

func SaveHeight(height int64) error {
	return nil
}
