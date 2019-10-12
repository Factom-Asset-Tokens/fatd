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

// Package metadata provides functions and SQL framents for working with the
// "metadata" table, which stores the sync height, sync DBKeyMR,
// factom.NetworkID, and factom.Identity.
package metadata

import (
	"fmt"

	"crawshaw.io/sqlite"
	"github.com/Factom-Asset-Tokens/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat"
	"github.com/Factom-Asset-Tokens/fatd/internal/db/entries"
)

// CreateTable is a SQL string that creates the "metadata" table.
//
// The "metadata" table has a foreign key reference to the "entries" table,
// which must exist first.
const CreateTable = `CREATE TABLE "metadata" (
        "id"                    INTEGER PRIMARY KEY,
        "sync_height"           INTEGER NOT NULL,
        "sync_db_key_mr"        BLOB NOT NULL,
        "network_id"            BLOB NOT NULL,
        "id_key_entry"          BLOB,
        "id_key_height"         INTEGER,

        "init_entry_id"         INTEGER,
        "num_issued"            INTEGER,

        FOREIGN KEY("init_entry_id") REFERENCES "entries"
);
`

// Insert the syncHeight, syncDBKeyMR, networkID, and if populated, the
// identity into the first row of the "metadata" table. This may only ever be
// called once for a given database.
func Insert(conn *sqlite.Conn, syncHeight uint32, syncDBKeyMR *factom.Bytes32,
	networkID factom.NetworkID, identity factom.Identity) error {
	stmt := conn.Prep(`INSERT INTO "metadata"
                ("id", "sync_height", "sync_db_key_mr",
                        "network_id", "id_key_entry", "id_key_height")
                VALUES (0, ?, ?, ?, ?, ?);`)
	stmt.BindInt64(1, int64(syncHeight))
	stmt.BindBytes(2, syncDBKeyMR[:])
	stmt.BindBytes(3, networkID[:])
	if identity.IsPopulated() {
		data, err := identity.MarshalBinary()
		if err != nil {
			return err
		}
		stmt.BindBytes(4, data)
		stmt.BindInt64(5, int64(identity.Height))
	} else {
		stmt.BindNull(4)
		stmt.BindNull(5)
	}
	_, err := stmt.Step()
	return err
}

// SetSync updates the "sync_height" and "sync_db_key_mr".
func SetSync(conn *sqlite.Conn, height uint32, dbKeyMR *factom.Bytes32) error {
	stmt := conn.Prep(`UPDATE "metadata" SET
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

// SetInitEntryID updates the "init_entry_id"
func SetInitEntryID(conn *sqlite.Conn, id int64) error {
	stmt := conn.Prep(`UPDATE "metadata" SET
                ("init_entry_id", "num_issued") = (?, 0);`)
	stmt.BindInt64(1, id)
	_, err := stmt.Step()
	if err != nil && conn.Changes() != 1 {
		panic(fmt.Errorf("expected exactly 1 change but got %v",
			conn.Changes()))
	}
	return err
}

func AddNumIssued(conn *sqlite.Conn, add uint64) error {
	stmt := conn.Prep(`UPDATE "metadata" SET
                "num_issued" = "num_issued" + ?;`)
	stmt.BindInt64(1, int64(add))
	_, err := stmt.Step()
	if err != nil && conn.Changes() != 1 {
		panic(fmt.Errorf("expected exactly 1 change but got %v",
			conn.Changes()))
	}
	return err
}

func Select(conn *sqlite.Conn) (syncHeight uint32, numIssued uint64,
	syncDBKeyMR *factom.Bytes32,
	networkID factom.NetworkID,
	identity factom.Identity,
	issuance fat.Issuance,
	err error) {
	stmt := conn.Prep(`SELECT "sync_height", "sync_db_key_mr", "network_id",
                "id_key_entry", "id_key_height", "init_entry_id", "num_issued"
                        FROM "metadata";`)
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

	syncDBKeyMR = new(factom.Bytes32)
	if stmt.ColumnBytes(1, syncDBKeyMR[:]) != len(syncDBKeyMR) {
		panic("invalid sync_db_key_mr length")
	}

	if stmt.ColumnBytes(2, networkID[:]) != len(networkID) {
		panic("invalid network_id length")
	}

	// Load chain.Identity...
	if stmt.ColumnType(3) == sqlite.SQLITE_NULL {
		// No Identity, therefore no Issuance.
		return
	}
	idKeyEntryData := make(factom.Bytes, stmt.ColumnLen(3))
	stmt.ColumnBytes(3, idKeyEntryData)
	if err = identity.UnmarshalBinary(idKeyEntryData); err != nil {
		err = fmt.Errorf("identity.UnmarshalBinary(): %w", err)
		return
	}
	identity.Height = uint32(stmt.ColumnInt64(4))

	// Load chain.Issuance...
	if stmt.ColumnType(5) == sqlite.SQLITE_NULL {
		// No issuance entry so far...
		return
	}
	initEntryID := stmt.ColumnInt64(5)
	issuance.Entry.Entry, err = entries.SelectByID(conn, initEntryID)
	if err != nil {
		return
	}
	if err = issuance.Validate(identity.ID1); err != nil {
		return
	}

	numIssued = uint64(stmt.ColumnInt64(6))

	return
}
