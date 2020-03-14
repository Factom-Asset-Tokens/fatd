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
	"os"
	"strings"
	"time"

	"crawshaw.io/sqlite"
	"crawshaw.io/sqlite/sqlitex"
	"github.com/Factom-Asset-Tokens/factom"
	"github.com/Factom-Asset-Tokens/fatd/internal/db/eblock"
	"github.com/Factom-Asset-Tokens/fatd/internal/db/metadata"
	_log "github.com/Factom-Asset-Tokens/fatd/internal/log"
	"github.com/subchen/go-trylock/v2"
)

type FactomChain struct {
	ID          *factom.Bytes32
	Head        factom.EBlock
	HeadDBKeyMR *factom.Bytes32
	NetworkID   factom.NetworkID
	SyncHeight  uint32
	SyncDBKeyMR *factom.Bytes32

	DBPath    string
	DBFile    string
	Conn      *sqlite.Conn  // Read/Write
	Pool      *sqlitex.Pool // Read Only Pool
	CloseMtx  trylock.TryLocker
	SaveDepth int

	Log _log.Log
}

func NewFactomChain(ctx context.Context,
	dbPath string, chainID *factom.Bytes32,
	networkID factom.NetworkID) (_ FactomChain, err error) {

	fname := chainID.String() + dbFileExtension
	path := dbPath + fname

	// Ensure that the database file doesn't already exist.
	_, err = os.Stat(path)
	if err == nil {
		err = fmt.Errorf("already exists: %v", path)
		return
	}
	if !os.IsNotExist(err) { // Any other error is unexpected.
		return
	}

	log := _log.New("chain", strings.TrimRight(fname, dbFileExtension))
	conn, pool, err := OpenConnPoolChain(ctx, dbPath+fname)
	if err != nil {
		err = fmt.Errorf("db.OpenConnPool(): %w", err)
		return
	}

	defer func() {
		if err != nil {
			Close(conn, pool)
			if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
				log.Errorf("os.Remove(): %w", err)
			}
		}
	}()

	if err = metadata.InsertFactomChain(conn, networkID, chainID); err != nil {
		err = fmt.Errorf("db/metadata.InsertFactomChain(): %w", err)
		return
	}

	var zero factom.Bytes32

	return FactomChain{
		ID:        chainID,
		NetworkID: networkID,

		Head:        factom.EBlock{ChainID: chainID, KeyMR: &zero},
		SyncDBKeyMR: &zero,

		DBPath:   path,
		DBFile:   fname,
		Conn:     conn,
		Pool:     pool,
		CloseMtx: trylock.New(),

		Log: log,
	}, nil
}

func OpenFactomChain(ctx context.Context,
	dbPath, fname string) (_ FactomChain, err error) {
	log := _log.New("chain", strings.TrimRight(fname, dbFileExtension))
	log.Info("Opening...")
	conn, pool, err := OpenConnPoolChain(ctx, dbPath+fname)
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			Close(conn, pool)
		}
	}()

	chain := FactomChain{
		DBFile:   fname,
		Conn:     conn,
		Pool:     pool,
		CloseMtx: trylock.New(),
	}

	if err = chain.load(); err != nil {
		return
	}

	return chain, nil
}
func (chain *FactomChain) load() error {
	// Load Chain Head
	eb, dbKeyMR, err := eblock.SelectLatest(chain.Conn)
	if err != nil {
		return err
	}
	if !eb.IsPopulated() {
		// A database must always have at least one EBlock.
		return fmt.Errorf("no eblock in database")
	}

	syncHeight, syncDBKeyMR, networkID, err := metadata.SelectFactomChain(chain.Conn)
	if err != nil {
		return err
	}

	chain.ID = eb.ChainID
	chain.Head = eb
	chain.HeadDBKeyMR = &dbKeyMR

	chain.SyncHeight = syncHeight
	chain.SyncDBKeyMR = &syncDBKeyMR
	chain.NetworkID = networkID
	chain.CloseMtx = trylock.New()

	return nil
}

// Close all database connections. Log any errors.
func (chain *FactomChain) Close() error {
	chain.CloseMtx.Lock()
	err := Close(chain.Conn, chain.Pool)
	if err != nil {
		chain.Log.Error(err)
	}
	return err
}

func (chain *FactomChain) LatestEntryTimestamp() time.Time {
	entries := chain.Head.Entries
	lastID := len(entries) - 1
	return entries[lastID].Timestamp
}

func (chain *FactomChain) SetSync(height uint32, dbKeyMR *factom.Bytes32) error {
	if height <= chain.SyncHeight {
		return nil
	}
	if err := metadata.SetSync(chain.Conn, height, dbKeyMR); err != nil {
		return err
	}
	chain.SyncHeight = height
	chain.SyncDBKeyMR = dbKeyMR
	return nil
}
