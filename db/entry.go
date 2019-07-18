package db

import (
	"fmt"
	"time"

	"crawshaw.io/sqlite"
	"github.com/Factom-Asset-Tokens/fatd/factom"
)

// InsertEntry inserts e and returns the id if successful.
func InsertEntry(conn *sqlite.Conn, e factom.Entry, ebSeq uint32) (int64, error) {
	data, err := e.MarshalBinary()
	if err != nil {
		return -1, fmt.Errorf("factom.Entry{}.MarshalBinary(): %v", err)
	}

	stmt := conn.Prep(`INSERT INTO entries
                (eb_seq, timestamp, hash, data)
                VALUES (?, ?, ?, ?);`)
	stmt.BindInt64(1, int64(ebSeq))
	stmt.BindInt64(2, int64(e.Timestamp.Unix()))
	stmt.BindBytes(3, e.Hash[:])
	stmt.BindBytes(4, data)

	if _, err := stmt.Step(); err != nil {
		return 0, err
	}
	return conn.LastInsertRowID(), nil
}

// MarkEntryValid sets valid to true for the entry with the given id.
func MarkEntryValid(conn *sqlite.Conn, id int64) error {
	stmt := conn.Prep(`UPDATE entries SET valid = 1 WHERE id = ?;`)
	stmt.BindInt64(1, id)
	_, err := stmt.Step()
	return err
}

// SelectEntryWhere is a SQL fragment that must be appended with the condition
// of a WHERE clause and a final semi-colon.
const SelectEntryWhere = `SELECT hash, data, timestamp, valid FROM entries WHERE `

// SelectEntry uses stmt to populate and return a new factom.Entry and whether
// it is marked as valid. Since column position is used to address the data,
// the stmt must start with `SELECT hash, data, timestamp, valid`. This can be
// called repeatedly until stmt.Step() returns false, in which case the
// returned factom.Entry will not be populated.
func SelectEntry(stmt *sqlite.Stmt) (factom.Entry, bool, error) {
	var e factom.Entry
	hasRow, err := stmt.Step()
	if err != nil {
		return e, false, err
	}
	if !hasRow {
		return e, false, nil
	}

	e.Hash = new(factom.Bytes32)
	if stmt.ColumnBytes(0, e.Hash[:]) != len(e.Hash) {
		return e, false, fmt.Errorf("invalid hash length")
	}

	data := make([]byte, stmt.ColumnLen(1))
	stmt.ColumnBytes(1, data)
	if err := e.UnmarshalBinary(data); err != nil {
		return e, false,
			fmt.Errorf("factom.Entry{}.UnmarshalBinary(%x): %v",
				data, err)
	}

	e.Timestamp = time.Unix(stmt.ColumnInt64(2), 0)

	return e, stmt.ColumnInt(3) > 0, nil
}

// SelectEntryByID returns the factom.Entry with the given id.
func SelectEntryByID(conn *sqlite.Conn, id int64) (factom.Entry, bool, error) {
	stmt := conn.Prep(SelectEntryWhere + `id = ?;`)
	stmt.BindInt64(1, id)
	return SelectEntry(stmt)
}

// SelectEntryByID returns the factom.Entry with the given hash.
func SelectEntryByHash(conn *sqlite.Conn,
	hash *factom.Bytes32) (factom.Entry, bool, error) {
	stmt := conn.Prep(SelectEntryWhere + `hash = ?;`)
	stmt.BindBytes(1, hash[:])
	return SelectEntry(stmt)
}
