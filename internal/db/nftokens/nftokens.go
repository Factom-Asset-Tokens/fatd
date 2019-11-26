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

// Package nftokens provides functions and SQL framents for working with the
// "nf_tokens" table, which stores fat.NFToken with owner, creation id, and
// metadata.
package nftokens

import (
	"encoding/json"
	"fmt"

	"crawshaw.io/sqlite"
	"crawshaw.io/sqlite/sqlitex"
	"github.com/AdamSLevy/sqlbuilder"
	"github.com/Factom-Asset-Tokens/factom"
	"github.com/Factom-Asset-Tokens/factom/fat1"
)

// CreateTable is a SQL string that creates the "nf_tokens" table.
//
// The "nf_tokens" table has foreign key references to the "entries" and
// "addresses" tables, which must exist first.
const CreateTable = `CREATE TABLE "nf_tokens" (
        "id"                      INTEGER PRIMARY KEY,
        "metadata"                BLOB,
        "creation_entry_id"       INTEGER NOT NULL,
        "owner_id"                INTEGER NOT NULL,

        FOREIGN KEY("creation_entry_id") REFERENCES "entries",
        FOREIGN KEY("owner_id") REFERENCES "addresses"
);
CREATE INDEX "idx_nf_tokens_metadata" ON nf_tokens("metadata");
CREATE INDEX "idx_nf_tokens_owner_id" ON nf_tokens("owner_id");
CREATE VIEW "nf_tokens_addresses" AS
        SELECT "nf_tokens"."id" AS "id",
                "metadata",
                "hash" AS "creation_hash",
                "address" AS "owner" FROM
                        "nf_tokens", "addresses", "entries" ON
                                "owner_id" = "addresses"."id" AND
                                "creation_entry_id" = "entries"."id";
`

// Insert a new NFToken with "owner_id" set to the "addresses" foreign key
// adrID and the "creation_entry_id" set to the "entries" foreign key entryID.
func Insert(conn *sqlite.Conn, nfID fat1.NFTokenID, adrID, entryID int64) (error, error) {
	stmt := conn.Prep(`INSERT INTO "nf_tokens"
                ("id", "owner_id", "creation_entry_id") VALUES (?, ?, ?);`)
	stmt.BindInt64(1, int64(nfID))
	stmt.BindInt64(2, adrID)
	stmt.BindInt64(3, entryID)
	if _, err := stmt.Step(); err != nil {
		if sqlite.ErrCode(err) == sqlite.SQLITE_CONSTRAINT_PRIMARYKEY {
			return fmt.Errorf("NFTokenID{%v} already exists", nfID), nil
		}
		return nil, err
	}
	return nil, nil
}

// SetOwner updates the "owner_id" of the given nfID to the given adrID.
//
// If the given adrID does not exist, a foreign key constraint error will be
// returned. If the nfID does not exist, this will panic.
//
// TODO: consider that the use of panic is inconsistent here. This function
// should never be called on an adrID that does not exist. Should it also panic
// on that constraint error too? Both reflect program integrity issues.
func SetOwner(conn *sqlite.Conn, nfID fat1.NFTokenID, adrID int64) error {
	stmt := conn.Prep(`UPDATE "nf_tokens" SET "owner_id" = ? WHERE "id" = ?;`)
	stmt.BindInt64(1, adrID)
	stmt.BindInt64(2, int64(nfID))
	_, err := stmt.Step()
	if conn.Changes() == 0 {
		panic("no NFTokenID updated")
	}
	return err
}

// SetMetadata updates the "metadata" to metadata for a given nfID.
//
// If the nfID does not exist, this will panic.
func SetMetadata(conn *sqlite.Conn, nfID fat1.NFTokenID, metadata json.RawMessage) error {
	stmt := conn.Prep(`UPDATE "nf_tokens" SET "metadata" = ? WHERE "id" = ?;`)
	stmt.BindBytes(1, metadata)
	stmt.BindInt64(2, int64(nfID))
	_, err := stmt.Step()
	if conn.Changes() == 0 {
		// This must only be called after the nfID has been inserted.
		panic("no NFTokenID updated")
	}
	return err
}

// SelectOwnerID returns the "owner_id" for the given nfID.
//
// If the nfID does not yet exist, (-1, nil) is returned.
func SelectOwnerID(conn *sqlite.Conn, nfID fat1.NFTokenID) (int64, error) {
	stmt := conn.Prep(`SELECT "owner_id" FROM "nf_tokens" WHERE "id" = ?;`)
	stmt.BindInt64(1, int64(nfID))
	ownerID, err := sqlitex.ResultInt64(stmt)
	if err != nil && err.Error() == "sqlite: statement has no results" {
		return -1, nil
	}
	if err != nil {
		return -1, err
	}
	return ownerID, nil
}

// SelectData returns the owner address, the creation entry hash, and the
// NFToken metadata for the given nfID
//
// If the nfID doesn't exist, all zero values are returned. Namely, check
// IsZero on the returned creation entry hash.
func SelectData(conn *sqlite.Conn, nfID fat1.NFTokenID) (
	factom.FAAddress, factom.Bytes32, []byte, error) {

	var owner factom.FAAddress
	var creationHash factom.Bytes32
	stmt := conn.Prep(`SELECT "owner", "metadata", "creation_hash"
                        FROM "nf_tokens_addresses" WHERE "id" = ?;`)
	stmt.BindInt64(1, int64(nfID))
	hasRow, err := stmt.Step()
	defer stmt.Reset()
	if err != nil {
		return owner, creationHash, nil, err
	}
	if !hasRow {
		return owner, creationHash, nil, nil
	}
	if stmt.ColumnBytes(0, owner[:]) != len(owner) {
		panic("invalid address length")
	}
	metadata := make([]byte, stmt.ColumnLen(1))
	stmt.ColumnBytes(1, metadata)

	if stmt.ColumnBytes(2, creationHash[:]) != len(creationHash) {
		panic("invalid hash length")
	}
	return owner, creationHash, metadata, nil
}

// SelectDataAll returns the nfIDs, owner addresses, creation entry hashes, and
// the NFToken metadata for the given pagination range of NFTokens.
//
// Pages start at 1.
func SelectDataAll(conn *sqlite.Conn, order string, page, limit uint) (
	[]fat1.NFTokenID, []factom.FAAddress, []factom.Bytes32, [][]byte, error) {
	if page == 0 {
		return nil, nil, nil, nil, fmt.Errorf("invalid page")
	}
	stmt := conn.Prep(`SELECT "id", "owner", "creation_hash", "metadata"
                        FROM "nf_tokens_addresses";`)
	defer stmt.Reset()

	var tkns []fat1.NFTokenID
	var owners []factom.FAAddress
	var creationHashes []factom.Bytes32
	var metadata [][]byte
	for {
		hasRow, err := stmt.Step()
		if err != nil {
			return nil, nil, nil, nil, err
		}
		if !hasRow {
			break
		}
		tkns = append(tkns, fat1.NFTokenID(stmt.ColumnInt64(0)))

		var owner factom.FAAddress
		if stmt.ColumnBytes(1, owner[:]) != len(owner) {
			panic("invalid address length")
		}
		owners = append(owners, owner)

		var creationHash factom.Bytes32
		if stmt.ColumnBytes(2, creationHash[:]) != len(creationHash) {
			panic("invalid hash length")
		}
		creationHashes = append(creationHashes, creationHash)

		data := make([]byte, stmt.ColumnLen(3))
		stmt.ColumnBytes(3, data)
		metadata = append(metadata, data)
	}
	return tkns, owners, creationHashes, metadata, nil
}

// SelectByOwner returns the fat1.NFTokens owned by the given adr for the given
// pagination range.
//
// Pages start at 1.
func SelectByOwner(conn *sqlite.Conn, adr *factom.FAAddress,
	page, limit uint, order string) (fat1.NFTokens, error) {
	if page == 0 {
		return nil, fmt.Errorf("invalid page")
	}
	var sql sqlbuilder.SQLBuilder
	sql.Append(`SELECT "id" FROM "nf_tokens" WHERE "owner_id" = (
                SELECT "id" FROM "addresses" WHERE "address" = ?)`,
		func(s *sqlite.Stmt, c int) int {
			s.BindBytes(c, adr[:])
			return 1
		})
	sql.OrderByPaginate("id", order, page, limit)

	stmt := sql.Prep(conn)
	defer stmt.Reset()
	nfTkns := make(fat1.NFTokens)
	for {
		hasRow, err := stmt.Step()
		if err != nil {
			return nil, err
		}
		if !hasRow {
			break
		}
		colVal := stmt.ColumnInt64(0)
		if colVal < 0 {
			panic("negative NFTokenID")
		}
		nfTkns[fat1.NFTokenID(colVal)] = struct{}{}
	}
	return nfTkns, nil
}
