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
	"fmt"
	"time"

	"crawshaw.io/sqlite"
	"github.com/Factom-Asset-Tokens/factom"
)

func (chain *Chain) insertEBlock(eb factom.EBlock, dbKeyMR *factom.Bytes32) error {
	// Ensure that this is the next EBlock.
	prevKeyMR, err := SelectKeyMR(chain.Conn, eb.Sequence-1)
	if *eb.PrevKeyMR != prevKeyMR {
		return fmt.Errorf("invalid EBlock{}.PrevKeyMR")
	}

	var data []byte
	data, err = eb.MarshalBinary()
	if err != nil {
		panic(fmt.Errorf("factom.EBlock{}.MarshalBinary(): %v", err))
	}
	stmt := chain.Conn.Prep(`INSERT INTO "eblocks"
                ("seq", "key_mr", "db_height", "db_key_mr", "timestamp", "data")
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
const SelectEBlockWhere = `SELECT "key_mr", "data", "timestamp" FROM "eblocks" WHERE `

// SelectEBlock uses stmt to populate and return a new factom.EBlock. Since
// column position is used to address the data, the stmt must start with
// `SELECT "key_mr", "data", "timestamp"`. This can be called repeatedly until
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
		panic("invalid key_mr length")
	}

	// Load timestamp so that entries have correct timestamps.
	eb.Timestamp = time.Unix(stmt.ColumnInt64(2), 0)

	data := make([]byte, stmt.ColumnLen(1))
	stmt.ColumnBytes(1, data)
	if err := eb.UnmarshalBinary(data); err != nil {
		panic(fmt.Errorf("factom.EBlock{}.UnmarshalBinary(%x): %v", data, err))
	}

	return eb, nil
}

func SelectEBlockByHeight(conn *sqlite.Conn, height uint32) (factom.EBlock, error) {
	stmt := conn.Prep(SelectEBlockWhere + `"db_height" = ?;`)
	stmt.BindInt64(1, int64(height))
	defer stmt.Reset()
	return SelectEBlock(stmt)
}

func SelectEBlockBySequence(conn *sqlite.Conn, seq uint32) (factom.EBlock, error) {
	stmt := conn.Prep(SelectEBlockWhere + `"seq" = ?;`)
	stmt.BindInt64(1, int64(seq))
	defer stmt.Reset()
	return SelectEBlock(stmt)
}

func SelectKeyMR(conn *sqlite.Conn, seq uint32) (factom.Bytes32, error) {
	var keyMR factom.Bytes32
	stmt := conn.Prep(`SELECT "key_mr" FROM "eblocks" WHERE "seq" = ?;`)
	stmt.BindInt64(1, int64(int32(seq))) // Preserve uint32(-1) as -1
	hasRow, err := stmt.Step()
	defer stmt.Reset()
	if err != nil {
		return keyMR, err
	}
	if !hasRow {
		return keyMR, fmt.Errorf("no such EBlock{Sequence: %v}", seq)
	}

	if stmt.ColumnBytes(0, keyMR[:]) != len(keyMR) {
		panic("invalid key_mr length")
	}

	return keyMR, nil
}

func SelectDBKeyMR(conn *sqlite.Conn, seq uint32) (factom.Bytes32, error) {
	var dbKeyMR factom.Bytes32
	stmt := conn.Prep(`SELECT "db_key_mr" FROM "eblocks" WHERE "seq" = ?;`)
	stmt.BindInt64(1, int64(int32(seq))) // Preserve uint32(-1) as -1
	hasRow, err := stmt.Step()
	defer stmt.Reset()
	if err != nil {
		return dbKeyMR, err
	}
	if !hasRow {
		return dbKeyMR, fmt.Errorf("no such EBlock{Sequence: %v}", seq)
	}

	if stmt.ColumnBytes(0, dbKeyMR[:]) != len(dbKeyMR) {
		panic("invalid key_mr length")
	}

	return dbKeyMR, nil
}

func SelectEBlockLatest(conn *sqlite.Conn) (factom.EBlock, factom.Bytes32, error) {
	var dbKeyMR factom.Bytes32
	stmt := conn.Prep(
		`SELECT "key_mr", "data", "timestamp", "db_key_mr" FROM "eblocks"
                        WHERE "seq" = (SELECT max("seq") FROM "eblocks");`)
	eb, err := SelectEBlock(stmt)
	defer stmt.Reset()
	if err != nil {
		return eb, dbKeyMR, err
	}
	if !eb.IsPopulated() {
		panic("no EBlocks")
	}

	if stmt.ColumnBytes(3, dbKeyMR[:]) != len(dbKeyMR) {
		panic("invalid db_key_mr length")
	}

	return eb, dbKeyMR, nil
}
