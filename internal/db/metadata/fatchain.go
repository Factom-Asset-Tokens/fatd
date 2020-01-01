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

package metadata

import (
	"fmt"

	"crawshaw.io/sqlite"
	"github.com/Factom-Asset-Tokens/factom"
	"github.com/Factom-Asset-Tokens/factom/fat"
	"github.com/Factom-Asset-Tokens/fatd/internal/db/entry"
)

// CreateTableFATChain is a SQL string that creates the "fat_chain" metadata
// table.
//
// The "fat_chain" table has a foreign key reference to the "entry" table,
// which must exist first.
const CreateTableFATChain = `CREATE TABLE "fat_chain" (
        "id"                    INTEGER PRIMARY KEY,

        "token"                 TEXT NOT NULL,
        "issuer"                BLOB NOT NULL,

        "id_key_entry_data"     BLOB,
        "id_key_height"         INTEGER,

        "init_entry_id"         INTEGER,
        "num_issued"            INTEGER,

        FOREIGN KEY("init_entry_id") REFERENCES "entry"
);
`

// Insert the TokenID and issuer Chain ID into the first row of the "fat_chain"
// table. This may only ever be called once for a given database.
func InsertFATChain(conn *sqlite.Conn, tokenID string, issuer *factom.Bytes32) error {
	stmt := conn.Prep(`INSERT INTO "fat_chain"
                ("id", "token", "issuer")
                VALUES (0, ?, ?);`)
	stmt.BindText(1, tokenID)
	stmt.BindBytes(2, issuer[:])
	_, err := stmt.Step()
	return err
}

func UpdateIdentity(conn *sqlite.Conn, identity factom.Identity) error {
	stmt := conn.Prep(`UPDATE "fat_chain" SET
                ("id_key_entry_data", "id_key_height") = (?, ?);`)
	data, err := identity.MarshalBinary()
	if err != nil {
		return fmt.Errorf("factom.Identity.MarshalBinary(): %w", err)
	}
	stmt.BindBytes(1, data)
	stmt.BindInt64(2, int64(identity.Height))

	_, err = stmt.Step()
	return err
}

// SetInitEntryID updates the "init_entry_id"
func SetInitEntryID(conn *sqlite.Conn, id int64) error {
	stmt := conn.Prep(`UPDATE "fat_chain" SET
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
	stmt := conn.Prep(`UPDATE "fat_chain" SET
                "num_issued" = "num_issued" + ?;`)
	stmt.BindInt64(1, int64(add))
	_, err := stmt.Step()
	if err != nil && conn.Changes() != 1 {
		panic(fmt.Errorf("expected exactly 1 change but got %v",
			conn.Changes()))
	}
	return err
}

func SelectFATChain(conn *sqlite.Conn) (numIssued uint64,
	tokenID string,
	identity factom.Identity,
	issuance fat.Issuance,
	err error) {
	stmt := conn.Prep(`SELECT "id_key_entry_data", "id_key_height",
                        "init_entry_id", "num_issued", "token", "issuer"
                        FROM "fat_chain";`)
	hasRow, err := stmt.Step()
	defer stmt.Reset()
	if err != nil {
		return
	}
	if !hasRow {
		err = fmt.Errorf("no saved metadata")
		return
	}

	tokenID = stmt.ColumnText(4)
	identity.ChainID = new(factom.Bytes32)
	if stmt.ColumnBytes(5, identity.ChainID[:]) != len(identity.ChainID) {
		panic("invalid identity chain id len")
	}

	// Load chain.Identity...
	if stmt.ColumnType(0) == sqlite.SQLITE_NULL {
		// No Identity, therefore no Issuance.
		return
	}
	idKeyEntryData := make(factom.Bytes, stmt.ColumnLen(0))
	stmt.ColumnBytes(0, idKeyEntryData)
	if err = identity.UnmarshalBinary(idKeyEntryData); err != nil {
		err = fmt.Errorf("identity.UnmarshalBinary(): %w", err)
		return
	}

	identity.Height = uint32(stmt.ColumnInt64(1))

	// Load chain.Issuance...
	if stmt.ColumnType(2) == sqlite.SQLITE_NULL {
		// No issuance entry so far...
		return
	}
	initEntryID := stmt.ColumnInt64(2)
	e, err := entry.SelectByID(conn, initEntryID)
	if err != nil {
		return
	}

	issuance, err = fat.NewIssuance(e, (*factom.Bytes32)(identity.ID1Key))
	if err != nil {
		return
	}

	numIssued = uint64(stmt.ColumnInt64(3))

	return
}
