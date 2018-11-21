// Package db provides database related functions. The intention is to hide
// external libraries and lower level database operations from the state
// package.
//
// We are using gorm and sqlite.
//
// Each FAT token has a separate database file.
// A FAT token has entries
package db
