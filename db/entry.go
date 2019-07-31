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
	stmt.BindInt64(1, int64(int32(ebSeq))) // Preserve uint32(-1) as -1
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
const SelectEntryWhere = `SELECT hash, data, timestamp FROM entries WHERE `

// SelectEntry uses stmt to populate and return a new factom.Entry and whether
// it is marked as valid. Since column position is used to address the data,
// the stmt must start with `SELECT hash, data, timestamp`. This can be called
// repeatedly until stmt.Step() returns false, in which case the returned
// factom.Entry will not be populated.
func SelectEntry(stmt *sqlite.Stmt) (factom.Entry, error) {
	var e factom.Entry
	hasRow, err := stmt.Step()
	if err != nil {
		return e, err
	}
	if !hasRow {
		return e, nil
	}

	e.Hash = new(factom.Bytes32)
	if stmt.ColumnBytes(0, e.Hash[:]) != len(e.Hash) {
		return e, fmt.Errorf("invalid hash length")
	}

	data := make([]byte, stmt.ColumnLen(1))
	stmt.ColumnBytes(1, data)
	if err := e.UnmarshalBinary(data); err != nil {
		return e, fmt.Errorf("factom.Entry{}.UnmarshalBinary(%x): %v",
			data, err)
	}

	e.Timestamp = time.Unix(stmt.ColumnInt64(2), 0)

	return e, nil
}

// SelectEntryByID returns the factom.Entry with the given id.
func SelectEntryByID(conn *sqlite.Conn, id int64) (factom.Entry, error) {
	stmt := conn.Prep(SelectEntryWhere + `id = ?;`)
	stmt.BindInt64(1, id)
	return SelectEntry(stmt)
}

// SelectEntryByHash returns the factom.Entry with the given hash.
func SelectEntryByHash(conn *sqlite.Conn, hash *factom.Bytes32) (factom.Entry, error) {
	stmt := conn.Prep(SelectEntryWhere + `hash = ?;`)
	return SelectEntry(stmt)
}

// SelectValidEntryByHash returns the factom.Entry with the given hash only if
// it is marked as valid.
func SelectValidEntryByHash(conn *sqlite.Conn,
	hash *factom.Bytes32) (factom.Entry, error) {
	stmt := conn.Prep(SelectEntryWhere + `hash = ? AND valid = true;`)
	return SelectEntry(stmt)
}

// CheckEntryUniqueValid checks if there have been any Entries earlier than id
// with the same hash that are valid. If so, then this Entry is a replay and
// should be ignored. Note that if the entry hash has already been saved but
// was invalid at the time, then the entry may be valid now.
func CheckEntryUniqueValid(conn *sqlite.Conn,
	id int64, hash *factom.Bytes32) (bool, error) {
	stmt := conn.Prep(`SELECT count(*) FROM entries WHERE
                valid = true AND id < ? AND hash = ?;`)
	stmt.BindInt64(1, id)
	stmt.BindBytes(2, hash[:])
	hasRow, err := stmt.Step()
	if err != nil {
		return false, err
	}
	if !hasRow {
		panic("should always return one row")
	}
	return stmt.ColumnInt(0) == 0, nil
}
