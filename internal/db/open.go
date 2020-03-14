// MIT License
//
// Copyright 2018 Canonical Ledgers, LLC
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
// IN THE SOFTWARE.

package db

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"

	"crawshaw.io/sqlite"
	"crawshaw.io/sqlite/sqlitex"
	"github.com/Factom-Asset-Tokens/factom"
	_log "github.com/Factom-Asset-Tokens/fatd/internal/log"
)

const (
	dbDriver        = "sqlite3"
	dbFileExtension = ".sqlite3"
	dbFileNameLen   = len(factom.Bytes32{})*2 + len(dbFileExtension)

	// TODO: expose this as flag
	PoolSize = 2
	// If the PoolSize is large, we can hit a limit on the number of open
	// SQLite connections. The limit seems to be somewhere around 500 total
	// connections but whether those are pooled or not seems to affect the
	// number.
	// 2 allows for roughly 170 simultaneously tracked chains.
	// 10 allows for roughly 46 simultaneously tracked chains.
	// Many improvements could be made to this.
)

var poolSize = runtime.NumCPU()

var (
	log _log.Log
)

func OpenAllFATChains(ctx context.Context, dbPath string) (chains []FATChain, err error) {
	log = _log.New("pkg", "db")
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
	files, err := ioutil.ReadDir(dbPath)
	if err != nil {
		return nil, fmt.Errorf("ioutil.ReadDir(%q): %w", dbPath, err)
	}
	chains = make([]FATChain, 0, len(files))
	for _, f := range files {
		fname := f.Name()
		chainID, err := fnameToChainID(fname)
		if err != nil {
			continue
		}
		chain, err := OpenFATChain(ctx, dbPath, fname)
		if err != nil {
			return nil, err
		}
		chains = append(chains, chain)
		if *chainID != *chain.ID {
			return nil, fmt.Errorf(
				"filename %v does not match database Chain ID %v",
				fname, chain.ID)
		}
	}
	return chains, nil
}
func fnameToChainID(fname string) (*factom.Bytes32, error) {
	invalidFName := fmt.Errorf("invalid filename: %v", fname)
	if len(fname) != dbFileNameLen ||
		fname[dbFileNameLen-len(dbFileExtension):dbFileNameLen] !=
			dbFileExtension {
		return nil, invalidFName
	}
	chainID := factom.NewBytes32(fname[0:64])
	if chainID.IsZero() {
		return nil, invalidFName
	}
	return &chainID, nil
}

const baseFlags = sqlite.SQLITE_OPEN_WAL |
	sqlite.SQLITE_OPEN_URI |
	sqlite.SQLITE_OPEN_NOMUTEX

// OpenConnPool opens a Conn to the sqlite3 database at dbURI and performs a
// number of checks and operations to ensure that the Conn is ready for use.
// The Interrupt on the Conn is set to ctx.Done().
//
// If the Conn is successfully initialized a read-only Pool will be opened on
// the same database.
//
// The caller is responsible for closing conn and pool if err is not nil.
func OpenConnPool(ctx context.Context, dbURI string,
	appID int32, initSchema string, migrations []func(*sqlite.Conn) error) (
	conn *sqlite.Conn, pool *sqlitex.Pool, err error) {

	flags := baseFlags | sqlite.SQLITE_OPEN_READWRITE | sqlite.SQLITE_OPEN_CREATE

	// Open Conn.
	if conn, err = sqlite.OpenConn(dbURI, flags); err != nil {
		err = fmt.Errorf("sqlite.OpenConn(%q, %x): %w", dbURI, flags, err)
		return
	}
	defer func() {
		if err != nil {
			if err := conn.Close(); err != nil {
				log.Error(err)
			}
			if err := os.Remove(dbURI); err != nil && !os.IsNotExist(err) {
				log.Errorf("os.Remove(): %w", err)
			}
		}
	}()

	// Set the interrupt
	conn.SetInterrupt(ctx.Done())

	if err = checkOrSetApplicationID(conn, "main"); err != nil {
		return
	}

	if err = applyMigrations(conn, initSchema, migrations); err != nil {
		return
	}

	// Foreign key checks are disabled on connections by default. We want
	// these checks on the main write database connection, but they are not
	// needed on the read connections.
	if err = enableForeignKeyChecks(conn); err != nil {
		return
	}

	// Snapshots are unreliable if auto checkpointing is enabled. So we
	// manually checkpoint in the engine every new EBlock and in Close.
	if err = disableAutoCheckpoint(conn); err != nil {
		return
	}

	// Ensure WAL file is created and ready for snapshots by ensuring at
	// least one transaction exists in the WAL.
	if err = ensureTransactionInWAL(conn); err != nil {
		return
	}

	if err = setupTempTables(conn, initSchema); err != nil {
		return
	}

	// Open Pool.
	flags = baseFlags | sqlite.SQLITE_OPEN_READONLY
	if pool, err = sqlitex.Open(dbURI, flags, poolSize); err != nil {
		err = fmt.Errorf("sqlitex.Open(%q, %x, %v): %w",
			dbURI, flags, poolSize, err)
		return
	}
	defer func() {
		if err != nil {
			if err := pool.Close(); err != nil {
				log.Error(err)
			}
		}
	}()

	// Prime pool for snapshot reads.
	// https://www.sqlite.org/c3ref/snapshot_open.html
	for i := 0; i < poolSize; i++ {
		if err = func() error {
			conn := pool.Get(ctx)
			defer pool.Put(conn)
			if err := setupTempTables(conn, initSchema); err != nil {
				return err
			}
			err := sqlitex.ExecScript(conn, "PRAGMA application_id;")
			if err != nil {
				return err
			}
			return nil
		}(); err != nil {
			return
		}
	}
	return
}

// Close all database connections. Log any errors.
func Close(conn *sqlite.Conn, pool *sqlitex.Pool) error {
	if err := pool.Close(); err != nil {
		return fmt.Errorf("pool.Close(): %v", err)
	}
	conn.SetInterrupt(nil)
	if err := sqlitex.ExecScript(conn, `PRAGMA wal_checkpoint;`); err != nil {
		return err
	}
	// Close this last so that the wal and shm files are removed.
	if err := conn.Close(); err != nil {
		return fmt.Errorf("conn.Close(): %v", err)
	}
	return nil
}

func checkOrSetApplicationID(conn *sqlite.Conn, db string) error {
	var appID int32
	if err := sqlitex.ExecTransient(conn,
		fmt.Sprintf(`PRAGMA %q."application_id";`, db),
		func(stmt *sqlite.Stmt) error {
			appID = stmt.ColumnInt32(0)
			return nil
		}); err != nil {
		return err
	}
	switch appID {
	case 0: // ApplicationID not set
		return sqlitex.ExecTransient(conn,
			fmt.Sprintf(`PRAGMA %q."application_id" = %v;`,
				db, ApplicationID),
			nil)
	case ApplicationID:
		return nil
	}
	return fmt.Errorf("invalid database: application_id")
}

func enableForeignKeyChecks(conn *sqlite.Conn) error {
	return sqlitex.ExecScript(conn, `PRAGMA foreign_keys = ON;`)
}

func disableAutoCheckpoint(conn *sqlite.Conn) error {
	return sqlitex.ExecScript(conn, `PRAGMA wal_autocheckpoint = 0;`)
}

func ensureTransactionInWAL(conn *sqlite.Conn) error {
	var uv int
	if err := sqlitex.ExecTransient(conn, `PRAGMA user_version;`,
		func(stmt *sqlite.Stmt) error {
			uv = stmt.ColumnInt(0)
			return nil
		}); err != nil {
		return err
	}
	if err := sqlitex.ExecScript(conn,
		fmt.Sprintf(`PRAGMA user_version = %v;`, uv)); err != nil {
		return err
	}
	return nil
}

func setupTempTables(conn *sqlite.Conn, initSchema string) error {
	return sqlitex.ExecScript(conn, fmt.Sprintf(initSchema+`
                        PRAGMA %[1]q.application_id;`, "temp"))
}
