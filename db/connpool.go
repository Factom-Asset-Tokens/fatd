package db

import (
	"fmt"

	"crawshaw.io/sqlite"
	"crawshaw.io/sqlite/sqlitex"
	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/flag"
)

// ConnPool combines a READWRITE sqlite.Conn and a READ_ONLY sqlitex.Pool with
// the corresponding ChainID. Only a single thread may use the Read/Write Conn
// at a time. Multi-threaded readers must Pool.Get a read-only Conn from the
// and must Pool.Put the Conn back when finished.
type ConnPool struct {
	ChainID       *factom.Bytes32
	*sqlite.Conn  // Read/Write
	*sqlitex.Pool // Read Only Pool
}

// Open a new ConnPool for the given chainID within flag.DBPath. Validate or
// apply chainDBSchema.
func Open(chainID *factom.Bytes32) (ConnPool, error) {
	cp, err := open(chainID.String()+dbFileExtension, chainDBSchema)
	if err != nil {
		return cp, err
	}
	cp.ChainID = chainID
	return cp, nil
}

// open a READWRITE Conn and a READ_ONLY Pool for flag.DBPath + "/" + fname.
// Validate or apply schema.
func open(fname, schema string) (ConnPool, error) {
	const baseFlags = sqlite.SQLITE_OPEN_WAL |
		sqlite.SQLITE_OPEN_URI |
		sqlite.SQLITE_OPEN_NOMUTEX
	var cp ConnPool
	var err error
	path := flag.DBPath + "/" + fname
	flags := baseFlags | sqlite.SQLITE_OPEN_READWRITE | sqlite.SQLITE_OPEN_CREATE
	cp.Conn, err = sqlite.OpenConn(path, flags)
	if err != nil {
		return cp, fmt.Errorf("sqlite.OpenConn(%q, %x): %v",
			path, flags, err)
	}
	if err := validateOrApplySchema(cp.Conn, schema); err != nil {
		return cp, err
	}
	flags = baseFlags | sqlite.SQLITE_OPEN_READONLY
	cp.Pool, err = sqlitex.Open(path, flags, PoolSize)
	if err != nil {
		return cp, fmt.Errorf("sqlitex.Open(%q, %x, %v): %v",
			path, flags, PoolSize, err)
	}
	return cp, nil
}

// Close the Conn and Pool. Log any errors.
func (cp ConnPool) Close() {
	if err := cp.Conn.Close(); err != nil {
		log.Errorf("%v: cp.Conn.Close(): %v", cp.ChainID, err)
	}
	if err := cp.Pool.Close(); err != nil {
		log.Errorf("%v: cp.Pool.Close(): %v", cp.ChainID, err)
	}
}
