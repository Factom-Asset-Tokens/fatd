package db

import (
	"context"

	"crawshaw.io/sqlite"
	"crawshaw.io/sqlite/sqlitex"
	"github.com/Factom-Asset-Tokens/fatd/internal/db/contract"
)

const ApplicationIDContractsDB = ApplicationID | 0x5C

func OpenConnPoolContract(ctx context.Context, dbPath string) (
	*sqlite.Conn, *sqlitex.Pool, error) {
	return OpenConnPool(ctx, dbPath,
		ApplicationIDContractsDB, contractDBSchema, contractDBMigrations)
}

const contractDBSchema = contract.CreateTable

var contractDBVersion = len(chainDBMigrations)

var contractDBMigrations = []func(*sqlite.Conn) error{}
