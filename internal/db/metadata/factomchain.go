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

// Package metadata provides functions and SQL fragments for working with the
// "factom_chain" and "fat_chain" metadata tables, which store the sync height,
// sync DBKeyMR, factom.NetworkID, and factom.Identity.
package metadata

import (
	"fmt"

	"crawshaw.io/sqlite"
	"github.com/Factom-Asset-Tokens/factom"
)

// CreateTableFactomChain is a SQL string that creates the "factom_chain"
// metadata table.
const CreateTableFactomChain = `
CREATE TABLE IF NOT EXISTS "factom_chain" (
        "id"                    INTEGER PRIMARY KEY,
        "chain_id"              BLOB NOT NULL,
        "network_id"            BLOB NOT NULL,

        "sync_height"           INTEGER NULL,
        "sync_db_key_mr"        BLOB NULL
);
`

// Insert the syncHeight, syncDBKeyMR, and networkID into the first row of the
// "factom_chain" table. This may only ever be called once for a given
// database.
func InsertFactomChain(conn *sqlite.Conn,
	networkID factom.NetworkID,
	chainID *factom.Bytes32) error {

	stmt := conn.Prep(`INSERT INTO "factom_chain"
                ("id", "network_id", "chain_id")
                VALUES (0, ?, ?);`)
	stmt.BindBytes(1, networkID[:])
	stmt.BindBytes(2, chainID[:])
	_, err := stmt.Step()
	return err
}

// SetSync updates the "sync_height" and "sync_db_key_mr".
func SetSync(conn *sqlite.Conn, height uint32, dbKeyMR *factom.Bytes32) error {
	stmt := conn.Prep(`UPDATE "factom_chain" SET
                ("sync_height", "sync_db_key_mr") = (?, ?);`)
	stmt.BindInt64(1, int64(height))
	stmt.BindBytes(2, dbKeyMR[:])
	_, err := stmt.Step()
	if err != nil && conn.Changes() != 1 {
		panic(fmt.Errorf("expected exactly 1 change but got %v",
			conn.Changes()))
	}
	return err
}

func SelectFactomChain(conn *sqlite.Conn) (syncHeight uint32,
	syncDBKeyMR factom.Bytes32,
	networkID factom.NetworkID,
	err error) {
	stmt := conn.Prep(`SELECT "sync_height", "sync_db_key_mr", "network_id"
                        FROM "factom_chain";`)
	hasRow, err := stmt.Step()
	defer stmt.Reset()
	if err != nil {
		return
	}
	if !hasRow {
		err = fmt.Errorf("no saved metadata")
		return
	}

	syncHeight = uint32(stmt.ColumnInt64(0))

	if stmt.ColumnBytes(1, syncDBKeyMR[:]) != len(syncDBKeyMR) {
		panic("invalid sync_db_key_mr length")
	}

	if stmt.ColumnBytes(2, networkID[:]) != len(networkID) {
		panic("invalid network_id length")
	}

	return
}
