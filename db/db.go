package db

import (
	"fmt"

	"github.com/Factom-Asset-Tokens/fatd/flag"
	_log "github.com/Factom-Asset-Tokens/fatd/log"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

type Metadata struct {
	gorm.Model

	Height int64 `gorm:"default:161460"`
}

var (
	gDB *gorm.DB
	log _log.Log

	metadata Metadata
)

const (
	dbDriver = "sqlite3"
)

func Open() error {
	log = _log.New("db")
	var err error
	if gDB, err = gorm.Open(dbDriver, flag.DBFile); err != nil {
		return fmt.Errorf("db.Open(%#v, %#v): %v", dbDriver, flag.DBFile, err)
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
		SaveHeight(flag.StartScanHeight)
	}

	return nil
}

func Close() error {
	if gDB == nil {
		return fmt.Errorf("%v", "DB is not open.")
	}
	return gDB.Close()
}

func GetSavedHeight() int64 {
	return metadata.Height
}

func SaveHeight(height int64) error {
	metadata.Height = height
	if err := gDB.Model(&metadata).Update("height", height).Error; err != nil {
		return fmt.Errorf("gDB.Model(&metadata).Update(%#v, %v): %v",
			"height", height, err)
	}
	return nil
}
