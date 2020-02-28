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

// Package entry provides functions and SQL framents for working with the
// "entry" table, which stores factom.Entry with a valid flag.
package entry

import (
	"fmt"
	"strings"
	"time"

	"crawshaw.io/sqlite"
	"crawshaw.io/sqlite/sqlitex"
	"github.com/AdamSLevy/sqlbuilder"
	"github.com/Factom-Asset-Tokens/factom"
	"github.com/Factom-Asset-Tokens/factom/fat1"
	"github.com/Factom-Asset-Tokens/fatd/internal/flag"
)

// CreateTable is a SQL string that creates the "entry" table.
//
// The "entry" table has a foreign key reference to the "eblock" table, which
// must exist first.
const CreateTable = `
CREATE TABLE IF NOT EXISTS %[1]q."entry" (
        "id"            INTEGER PRIMARY KEY,
        "eb_seq"        INTEGER NOT NULL,
        "timestamp"     INTEGER NOT NULL,
        "valid"         BOOL NOT NULL DEFAULT FALSE,
        "hash"          BLOB NOT NULL,
        "data"          BLOB NOT NULL,

        FOREIGN KEY("eb_seq") REFERENCES "eblock"
);
CREATE INDEX IF NOT EXISTS
        %[1]q"idx_entry_eb_seq" ON "entry"("eb_seq");
CREATE INDEX IF NOT EXISTS
        %[1]q"idx_entry_hash" ON "entry"("hash");
CREATE VIEW IF NOT EXISTS "temp"."entry_all" AS
        SELECT * FROM "temp"."entry"
        UNION ALL
        SELECT * FROM "main"."entry";
`

func Commit(conn *sqlite.Conn) error {
	return sqlitex.ExecScript(conn, `
INSERT INTO "main"."entry" SELECT * FROM "temp"."entry";
DELETE FROM "temp"."entry";
`)
}

// Insert e into the "entry" table with the EBlock reference ebSeq. If
// successful, the new row id of e is returned.
func Insert(conn *sqlite.Conn, e factom.Entry, ebSeq uint32) (int64, error) {
	// Defer FK checks until the outermost TRANSACTION is committed so we
	// can insert rows into "temp" tables that temporarily violate FK
	// constraints.
	if err := sqlitex.Exec(conn,
		`PRAGMA defer_foreign_keys = true;`, nil); err != nil {
		return -1, err
	}
	data, err := e.MarshalBinary()
	if err != nil {
		panic(fmt.Errorf("factom.Entry.MarshalBinary(): %w", err))
	}

	stmt := conn.Prep(`INSERT INTO "temp"."entry"
                ("id", "eb_seq", "timestamp", "hash", "data")
                VALUES ((SELECT max("id")+1 FROM "temp"."entry_all"),
                        ?, ?, ?, ?);`)
	i := sqlite.BindIncrementor()
	stmt.BindInt64(i(), int64(int32(ebSeq))) // Preserve uint32(-1) as -1
	stmt.BindInt64(i(), int64(e.Timestamp.Unix()))
	stmt.BindBytes(i(), e.Hash[:])
	stmt.BindBytes(i(), data)

	if _, err := stmt.Step(); err != nil {
		return -1, err
	}
	return conn.LastInsertRowID(), nil
}

// SetValid marks the entry valid at the id'th row of the "entry" table.
func SetValid(conn *sqlite.Conn, id int64) error {
	stmt := conn.Prep(`UPDATE "temp"."entry" SET "valid" = 1 WHERE "id" = ?;`)
	stmt.BindInt64(sqlite.BindIndexStart, id)
	_, err := stmt.Step()
	if err != nil {
		return err
	}
	if conn.Changes() == 0 {
		panic("no entry updated")
	}
	return nil
}

// SelectWhere is a SQL fragment for retrieving rows from the "entry" table
// with Select().
const SelectWhere = `SELECT "hash", "data", "timestamp" FROM "temp"."entry_all" WHERE `

// Select the next factom.Entry from the given prepared Stmt.
//
// The Stmt must be created with a SQL string starting with SelectWhere.
func Select(stmt *sqlite.Stmt) (factom.Entry, error) {
	var e factom.Entry
	hasRow, err := stmt.Step()
	if err != nil {
		return e, err
	}
	if !hasRow {
		return e, nil
	}

	i := sqlite.ColumnIncrementor()
	e.Hash = new(factom.Bytes32)
	if stmt.ColumnBytes(i(), e.Hash[:]) != len(e.Hash) {
		panic("invalid hash length")
	}

	var data []byte
	{
		i := i()
		data = make([]byte, stmt.ColumnLen(i))
		stmt.ColumnBytes(i, data)
	}
	if err := e.UnmarshalBinary(data); err != nil {
		panic(fmt.Errorf("factom.Entry.UnmarshalBinary(%v): %w",
			factom.Bytes(data), err))
	}

	e.Timestamp = time.Unix(stmt.ColumnInt64(i()), 0)

	return e, nil
}

// SelectByID returns the factom.Entry at row id.
func SelectByID(conn *sqlite.Conn, id int64) (factom.Entry, error) {
	stmt := conn.Prep(SelectWhere + `"id" = ?;`)
	stmt.BindInt64(1, id)
	defer stmt.Reset()
	return Select(stmt)
}

// SelectByHash returns the first factom.Entry with hash.
func SelectByHash(conn *sqlite.Conn, hash *factom.Bytes32) (factom.Entry, error) {
	stmt := conn.Prep(SelectWhere + `"hash" = ?;`)
	stmt.BindBytes(sqlite.BindIndexStart, hash[:])
	defer stmt.Reset()
	return Select(stmt)
}

// SelectValidByHash returns the first valid factom.Entry with hash.
func SelectValidByHash(conn *sqlite.Conn, hash *factom.Bytes32) (factom.Entry, error) {
	stmt := conn.Prep(SelectWhere + `"hash" = ? AND "valid" = true;`)
	stmt.BindBytes(sqlite.BindIndexStart, hash[:])
	defer stmt.Reset()
	return Select(stmt)
}

// SelectCount returns the total number of rows in the "entry" table. If
// validOnly is true, only the rows where "valid" = true are counted.
func SelectCount(conn *sqlite.Conn, validOnly bool) (int64, error) {
	stmt := conn.Prep(`SELECT count(*) FROM "temp"."entry_all" WHERE
                (? OR "valid" = true);`)
	stmt.BindBool(sqlite.BindIndexStart, !validOnly)
	return sqlitex.ResultInt64(stmt)
}

func init() {
	sqlbuilder.LimitMaxDefault = uint(flag.APIMaxLimit)
}

// SelectByAddress returns all the factom.Entry where adrs and nfTkns were
// involved in the valid transaction, for the given pagination range.
//
// Pages start at 1.
//
// TODO: This should probably be moved out of the entry package and into a db
// package that is more specific to FAT0 and FAT1.
func SelectByAddress(conn *sqlite.Conn, startHash *factom.Bytes32,
	adrs []factom.FAAddress, nfTkns fat1.NFTokens,
	toFrom, order string,
	page, limit uint) ([]factom.Entry, error) {
	if page == 0 {
		return nil, fmt.Errorf("invalid page")
	}
	var sql sqlbuilder.SQLBuilder
	sql.Append(SelectWhere + `"valid" = true`)
	if startHash != nil {
		sql.Append(` AND "id" >= (SELECT "id" FROM "temp"."entry_all"
                                                WHERE "hash" = ?)`,
			func(s *sqlite.Stmt, p int) int {
				s.BindBytes(p, startHash[:])
				return 1
			})
	}
	var to bool
	switch strings.ToLower(toFrom) {
	case "to":
		to = true
	case "from", "":
	default:
		panic(fmt.Errorf("invalid toFrom: %v", toFrom))
	}
	if len(nfTkns) > 0 {
		sql.WriteString(` AND "id" IN (
                                SELECT "entry_id" FROM "nf_token_address_transactions"
                                        WHERE "nf_tkn_id" IN (`) // 2 open (
		sql.BindNParams(len(nfTkns), func(s *sqlite.Stmt, p int) int {
			i := 0
			for nfTkn := range nfTkns {
				s.BindInt64(p+i, int64(nfTkn))
				i++
			}
			return len(nfTkns)
		})
		sql.WriteString(`)`) // 1 open (
		if len(adrs) > 0 {
			sql.WriteString(` AND "address_id" IN (
                                SELECT "id" FROM "address"
                                        WHERE "address" IN (`) // 3 open (
			sql.BindNParams(len(adrs), func(s *sqlite.Stmt, p int) int {
				for i, adr := range adrs {
					s.BindBytes(p+i, adr[:])
				}
				return len(adrs)
			})
			sql.WriteString(`))`) // 1 open (
		}
		if len(toFrom) > 0 {
			sql.Append(` AND "to" = ?`, func(s *sqlite.Stmt, p int) int {
				s.BindBool(p, to)
				return 1
			})
		}
		sql.WriteString(`)`) // 0 open {
	} else if len(adrs) > 0 {
		sql.WriteString(` AND "id" IN (
                                SELECT "entry_id" FROM "address_transactions"
                                        WHERE "address_id" IN (
                                                SELECT "id" FROM "address"
                                                        WHERE "address" IN (`) // 3 open (

		sql.BindNParams(len(adrs), func(s *sqlite.Stmt, p int) int {
			for i, adr := range adrs {
				s.BindBytes(p+i, adr[:])
			}
			return len(adrs)
		})
		sql.WriteString(`))`) // 1 open (
		if len(toFrom) > 0 {
			sql.Append(` AND "to" = ?`, func(s *sqlite.Stmt, p int) int {
				s.BindBool(p, to)
				return 1
			})
		}
		sql.WriteString(`)`) // 0 open (
	}

	sql.OrderByPaginate("id", order, page, limit)

	stmt := sql.Prep(conn)
	defer stmt.Reset()

	var entry []factom.Entry
	for {
		e, err := Select(stmt)
		if err != nil {
			return nil, err
		}
		if !e.IsPopulated() {
			break
		}
		entry = append(entry, e)
	}

	return entry, nil
}

// CheckUniquelyValid returns true if there are no valid entry earlier than
// id that have the same hash. If id is 0, then all entry are checked.
func CheckUniquelyValid(conn *sqlite.Conn,
	id int64, hash *factom.Bytes32) (bool, error) {
	stmt := conn.Prep(`SELECT count(*) FROM "temp"."entry_all" WHERE
                "valid" = true AND (? OR "id" < ?) AND "hash" = ?;`)
	stmt.BindBool(1, id > 0)
	stmt.BindInt64(2, id)
	stmt.BindBytes(3, hash[:])
	val, err := sqlitex.ResultInt(stmt)
	if err != nil {
		return false, err
	}
	return val == 0, nil
}

// SelectLatestValid returns the most recent valid factom.Entry.
func SelectLatestValid(conn *sqlite.Conn) (factom.Entry, error) {
	stmt := conn.Prep(SelectWhere +
		`"id" = (SELECT max("id") FROM "temp"."entry_all" WHERE "valid" = true);`)
	e, err := Select(stmt)
	defer stmt.Reset()
	if err != nil {
		return e, err
	}
	return e, nil
}
