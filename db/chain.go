package db

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"crawshaw.io/sqlite"
	"crawshaw.io/sqlite/sqlitex"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat"
	"github.com/Factom-Asset-Tokens/fatd/flag"
	_log "github.com/Factom-Asset-Tokens/fatd/log"
)

var (
	log _log.Log
)

const (
	dbDriver        = "sqlite3"
	dbFileExtension = ".sqlite3"
	dbFileNameLen   = len(factom.Bytes32{})*2 + len(dbFileExtension)

	PoolSize = 10
)

type Chain struct {
	ID            *factom.Bytes32
	TokenID       string
	IssuerChainID *factom.Bytes32
	Head          factom.EBlock
	DBKeyMR       *factom.Bytes32
	factom.Identity
	NetworkID factom.NetworkID

	SyncHeight  uint32
	SyncDBKeyMR *factom.Bytes32

	fat.Issuance
	NumIssued uint64

	*sqlite.Conn  // Read/Write
	*sqlitex.Pool // Read Only Pool
	Log           _log.Log

	apply applyFunc
}

func OpenNew(dbKeyMR *factom.Bytes32, eb factom.EBlock, networkID factom.NetworkID,
	identity factom.Identity) (chain *Chain, err error) {
	fname := eb.ChainID.String() + dbFileExtension
	path := flag.DBPath + "/" + fname

	nameIDs := eb.Entries[0].ExtIDs
	if !fat.ValidTokenNameIDs(nameIDs) {
		return nil, fmt.Errorf("invalid token chain Name IDs")
	}

	// Ensure that the database file doesn't already exist.
	_, err = os.Stat(path)
	if err == nil {
		return nil, fmt.Errorf("already exists: %v", path)
	}
	if !os.IsNotExist(err) { // Any other error is unexpected.
		return nil, err
	}

	chain, err = open(fname)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			chain.Close()
			if err := os.Remove(path); err != nil {
				chain.Log.Errorf("os.Remove(): %v", err)
			}
		}
	}()
	chain.ID = eb.ChainID
	chain.TokenID, chain.IssuerChainID = fat.TokenIssuer(nameIDs)
	chain.DBKeyMR = dbKeyMR
	chain.Identity = identity
	chain.SyncHeight = eb.Height
	chain.SyncDBKeyMR = dbKeyMR
	chain.NetworkID = networkID

	if err := chain.insertMetadata(); err != nil {
		return nil, err
	}

	// Ensure that the coinbase address has rowid = 1.
	coinbase := fat.Coinbase()
	if _, err := chain.addressAdd(&coinbase, 0); err != nil {
		return nil, err
	}

	chain.setApplyFunc()
	if err := chain.Apply(dbKeyMR, eb); err != nil {
		return nil, err
	}

	return chain, nil
}

func Open(fname string) (*Chain, error) {
	chain, err := open(fname)
	if err != nil {
		return nil, err
	}

	if err := chain.loadMetadata(); err != nil {
		return nil, err
	}

	return chain, nil
}

func OpenAll() (chains []*Chain, err error) {
	log = _log.New("pkg", "db")
	// Try to create the database directory in case it doesn't already
	// exist.
	if err := os.Mkdir(flag.DBPath, 0755); err != nil {
		if !os.IsExist(err) {
			return nil, fmt.Errorf("os.Mkdir(%#v): %v", flag.DBPath, err)
		}
		log.Debug("Using existing database directory...")
	}

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
	files, err := ioutil.ReadDir(flag.DBPath)
	if err != nil {
		return nil, fmt.Errorf("ioutil.ReadDir(%q): %v", flag.DBPath, err)
	}
	chains = make([]*Chain, 0, len(files))
	for _, f := range files {
		fname := f.Name()
		chainID, err := fnameToChainID(fname)
		if err != nil {
			continue
		}
		log.Debugf("Loading chain: %v", chainID)
		chain, err := Open(fname)
		if err != nil {
			return nil, err
		}
		if *chainID != *chain.ID {
			return nil, fmt.Errorf("chain id does not match filename")
		}
		chains = append(chains, chain)
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
	chainID := factom.NewBytes32FromString(fname[0:64])
	if chainID == nil {
		return nil, invalidFName
	}
	return chainID, nil
}

func open(fname string) (*Chain, error) {
	const baseFlags = sqlite.SQLITE_OPEN_WAL |
		sqlite.SQLITE_OPEN_URI |
		sqlite.SQLITE_OPEN_NOMUTEX
	path := flag.DBPath + "/" + fname
	flags := baseFlags | sqlite.SQLITE_OPEN_READWRITE | sqlite.SQLITE_OPEN_CREATE
	conn, err := sqlite.OpenConn(path, flags)
	if err != nil {
		return nil, fmt.Errorf("sqlite.OpenConn(%q, %x): %v",
			path, flags, err)
	}
	if err := validateOrApplySchema(conn, chainDBSchema); err != nil {
		return nil, err
	}
	if err := sqlitex.ExecScript(conn, `PRAGMA foreign_keys = ON;`); err != nil {
		return nil, err
	}
	flags = baseFlags | sqlite.SQLITE_OPEN_READONLY
	pool, err := sqlitex.Open(path, flags, PoolSize)
	if err != nil {
		return nil, fmt.Errorf("sqlitex.Open(%q, %x, %v): %v",
			path, flags, PoolSize, err)
	}
	return &Chain{Conn: conn, Pool: pool,
		Log: _log.New("chain", strings.TrimRight(fname, dbFileExtension)),
	}, nil
}

func (chain *Chain) Close() {
	if err := chain.Pool.Close(); err != nil {
		chain.Log.Errorf("chain.Pool.Close(): %v", err)
	}
	// Close this last so that the wal and shm files are removed.
	if err := chain.Conn.Close(); err != nil {
		chain.Log.Errorf("chain.Conn.Close(): %v", err)
	}
}
