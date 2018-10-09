package db

import (
	_ "github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

func Init() error {
	setupLogger()
	return nil
}
