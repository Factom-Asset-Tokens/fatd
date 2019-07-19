package db

import (
	"fmt"
	"time"

	"crawshaw.io/sqlite"
	"github.com/Factom-Asset-Tokens/fatd/factom"
)

// InsertEBlock inserts eb only if it is the next factom.EBlock in the
// sequence.
func InsertEBlock(conn *sqlite.Conn, eb factom.EBlock, dbKeyMR *factom.Bytes32) error {
	// Ensure that this is the next EBlock.
	prevKeyMR, err := SelectKeyMR(conn, eb.Sequence-1)
	if *eb.PrevKeyMR != prevKeyMR {
		return fmt.Errorf("invalid EBlock{}.PrevKeyMR")
	}

	var data []byte
	data, err = eb.MarshalBinary()
	if err != nil {
		return fmt.Errorf("factom.EBlock{}.MarshalBinary(): %v", err)
	}
	stmt := conn.Prep(`INSERT INTO eblocks
                (seq, key_mr, db_height, db_key_mr, timestamp, data)
                VALUES (?, ?, ?, ?, ?, ?);`)
	stmt.BindInt64(1, int64(eb.Sequence))
	stmt.BindBytes(2, eb.KeyMR[:])
	stmt.BindInt64(3, int64(eb.Height))
	stmt.BindBytes(4, dbKeyMR[:])
	stmt.BindInt64(5, eb.Timestamp.Unix())
	stmt.BindBytes(6, data)

	_, err = stmt.Step()
	return err
}

// SelectEBlockWhere is a SQL fragment that must be appended with the condition
// of a WHERE clause and a final semi-colon.
const SelectEBlockWhere = `SELECT key_mr, data, timestamp FROM eblocks WHERE `

// SelectEBlock uses stmt to populate and return a new factom.EBlock. Since
// column position is used to address the data, the stmt must start with
// `SELECT key_mr, data, timestamp`. This can be called repeatedly until
// stmt.Step() returns false, in which case the returned factom.EBlock will not
// be populated.
func SelectEBlock(stmt *sqlite.Stmt) (factom.EBlock, error) {
	var eb factom.EBlock
	hasRow, err := stmt.Step()
	if err != nil {
		return eb, err
	}
	if !hasRow {
		return eb, nil
	}

	eb.KeyMR = new(factom.Bytes32)
	if stmt.ColumnBytes(0, eb.KeyMR[:]) != len(eb.KeyMR) {
		return eb, fmt.Errorf("invalid key_mr length")
	}

	// Load timestamp so that entries have correct timestamps.
	eb.Timestamp = time.Unix(stmt.ColumnInt64(2), 0)

	data := make([]byte, stmt.ColumnLen(1))
	stmt.ColumnBytes(1, data)
	if err := eb.UnmarshalBinary(data); err != nil {
		return eb, fmt.Errorf("factom.EBlock{}.UnmarshalBinary(%x): %v",
			data, err)
	}

	return eb, nil
}

func SelectEBlockByHeight(conn *sqlite.Conn, height uint32) (factom.EBlock, error) {
	stmt := conn.Prep(SelectEBlockWhere + `db_height = ?;`)
	stmt.BindInt64(1, int64(height))
	return SelectEBlock(stmt)
}

func SelectEBlockBySequence(conn *sqlite.Conn, seq uint32) (factom.EBlock, error) {
	stmt := conn.Prep(SelectEBlockWhere + `seq = ?;`)
	stmt.BindInt64(1, int64(seq))
	return SelectEBlock(stmt)
}

func SelectKeyMR(conn *sqlite.Conn, seq uint32) (factom.Bytes32, error) {
	var keyMR factom.Bytes32
	stmt := conn.Prep(`SELECT key_mr FROM eblocks WHERE seq = ?;`)
	stmt.BindInt64(1, int64(int32(seq))) // Preserve uint32(-1) as -1
	hasRow, err := stmt.Step()
	if err != nil {
		return keyMR, err
	}
	if !hasRow {
		return keyMR, nil
	}

	if stmt.ColumnBytes(0, keyMR[:]) != len(keyMR) {
		return keyMR, fmt.Errorf("invalid key_mr length")
	}

	return keyMR, nil
}

func SelectLatestEBlock(conn *sqlite.Conn) (factom.EBlock, factom.Bytes32, error) {
	var dbKeyMR factom.Bytes32
	stmt := conn.Prep(`SELECT key_mr, data, timestamp, db_key_mr FROM eblocks
                WHERE seq = (SELECT MAX(seq) FROM eblocks);`)
	eb, err := SelectEBlock(stmt)
	if err != nil {
		return eb, dbKeyMR, err
	}
	if !eb.IsPopulated() {
		return eb, dbKeyMR, nil
	}

	if stmt.ColumnBytes(3, dbKeyMR[:]) != len(dbKeyMR) {
		return eb, dbKeyMR, fmt.Errorf("invalid db_key_mr length")
	}

	return eb, dbKeyMR, nil
}
