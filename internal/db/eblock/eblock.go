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

// Package eblock provides functions and SQL framents for working with the
// "eblock" table, which stores factom.EBlock.
package eblock

import (
	"fmt"
	"time"

	"crawshaw.io/sqlite"
	"github.com/Factom-Asset-Tokens/factom"
)

// CreateTable is a SQL string that creates the "eblock" table.
const CreateTable = `
CREATE TABLE IF NOT EXISTS "eblock" (
        "seq"           INTEGER PRIMARY KEY,
        "key_mr"        BLOB NOT NULL UNIQUE,
        "db_height"     INTEGER NOT NULL UNIQUE,
        "db_key_mr"     BLOB NOT NULL UNIQUE,
        "timestamp"     INTEGER NOT NULL,
        "data"          BLOB NOT NULL
);
`

// Insert eb into the "eblock" table with dbKeyMR.
func Insert(conn *sqlite.Conn, eb factom.EBlock, dbKeyMR *factom.Bytes32) error {
	// Ensure that this is the next EBlock.
	prevKeyMR, err := SelectKeyMR(conn, eb.Sequence-1)
	if err != nil {
		return fmt.Errorf("eblock.SelectKeyMR(seq: %v): %w",
			eb.Sequence-1, err)
	}
	if *eb.PrevKeyMR != prevKeyMR {
		return fmt.Errorf("invalid EBlock.PrevKeyMR, database:%v but eblock:%v",
			prevKeyMR, eb.PrevKeyMR)
	}

	var data []byte
	data, err = eb.MarshalBinary()
	if err != nil {
		panic(fmt.Errorf("factom.EBlock.MarshalBinary(): %w", err))
	}
	stmt := conn.Prep(`INSERT INTO "eblock"
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

// SelectWhere is a SQL fragment for retrieving rows from the "eblock" table
// with Select().
const SelectWhere = `SELECT "key_mr", "data", "timestamp" FROM "eblock" WHERE `

// Select the next factom.EBlock from the given prepared Stmt.
//
// The Stmt must be created with a SQL string starting with SelectWhere.
func Select(stmt *sqlite.Stmt) (factom.EBlock, error) {
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
		panic(fmt.Errorf("factom.EBlock.UnmarshalBinary(%v): %w",
			factom.Bytes(data), err))
	}

	return eb, nil
}

// SelectByHeight returns the factom.EBlock with the given height.
func SelectByHeight(conn *sqlite.Conn, height uint32) (factom.EBlock, error) {
	stmt := conn.Prep(SelectWhere + `"db_height" = ?;`)
	stmt.BindInt64(1, int64(height))
	defer stmt.Reset()
	return Select(stmt)
}

// SelectBySequence returns the factom.EBlock with sequence seq.
func SelectBySequence(conn *sqlite.Conn, seq uint32) (factom.EBlock, error) {
	stmt := conn.Prep(SelectWhere + `"seq" = ?;`)
	stmt.BindInt64(1, int64(seq))
	defer stmt.Reset()
	return Select(stmt)
}

// SelectKeyMR returns the KeyMR for the EBlock with sequence seq.
func SelectKeyMR(conn *sqlite.Conn, seq uint32) (factom.Bytes32, error) {
	var keyMR factom.Bytes32
	stmt := conn.Prep(`SELECT "key_mr" FROM "eblock" WHERE "seq" = ?;`)
	stmt.BindInt64(1, int64(int32(seq))) // Preserve uint32(-1) as -1
	hasRow, err := stmt.Step()
	defer stmt.Reset()
	if err != nil {
		return keyMR, err
	}
	if !hasRow {
		return keyMR, nil
	}

	if stmt.ColumnBytes(0, keyMR[:]) != len(keyMR) {
		panic("invalid key_mr length")
	}

	return keyMR, nil
}

// SelectDBKeyMR returns the DBKeyMR for the EBlock with sequence seq.
func SelectDBKeyMR(conn *sqlite.Conn, seq uint32) (factom.Bytes32, error) {
	var dbKeyMR factom.Bytes32
	stmt := conn.Prep(`SELECT "db_key_mr" FROM "eblock" WHERE "seq" = ?;`)
	stmt.BindInt64(1, int64(int32(seq))) // Preserve uint32(-1) as -1
	hasRow, err := stmt.Step()
	defer stmt.Reset()
	if err != nil {
		return dbKeyMR, err
	}
	if !hasRow {
		return dbKeyMR, nil
	}

	if stmt.ColumnBytes(0, dbKeyMR[:]) != len(dbKeyMR) {
		panic("invalid key_mr length")
	}

	return dbKeyMR, nil
}

// SelectLatest returns the most recent factom.EBlock.
func SelectLatest(conn *sqlite.Conn) (factom.EBlock, factom.Bytes32, error) {
	var dbKeyMR factom.Bytes32
	stmt := conn.Prep(
		`SELECT "key_mr", "data", "timestamp", "db_key_mr" FROM "eblock"
                        WHERE "seq" = (SELECT max("seq") FROM "eblock");`)
	eb, err := Select(stmt)
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
