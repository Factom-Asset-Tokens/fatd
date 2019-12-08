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

// Package addresses provides functions and SQL framents for working with the
// "addresses" table, which stores factom.FAAddress with its balance.
package addresses

import (
	"fmt"

	"crawshaw.io/sqlite"
	"crawshaw.io/sqlite/sqlitex"
	"github.com/Factom-Asset-Tokens/factom"
)

// CreateTable is a SQL string that creates the "addresses" table.
const CreateTable = `CREATE TABLE "addresses" (
        "id"            INTEGER PRIMARY KEY,
        "address"       BLOB NOT NULL UNIQUE,
        "balance"       INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("balance" >= 0)
);
`

// Add adds add to the balance of adr, creating a new row in "addresses" if it
// does not exist. If successful, the row id of adr is returned.
func Add(conn *sqlite.Conn, adr *factom.FAAddress, add uint64) (int64, error) {
	stmt := conn.Prep(`INSERT INTO "addresses"
                ("address", "balance") VALUES (?, ?)
                ON CONFLICT("address") DO
                UPDATE SET "balance" = "balance" + "excluded"."balance";`)
	stmt.BindBytes(1, adr[:])
	stmt.BindInt64(2, int64(add))
	_, err := stmt.Step()
	if err != nil {
		return -1, err
	}
	return SelectID(conn, adr)
}

const sqlitexNoResultsErr = "sqlite: statement has no results"

// Sub subtracts sub from the balance of adr creating the row in "addresses" if
// it does not exist and sub is 0. If successful, the row id of adr is
// returned. If subtracting sub would result in a negative balance, txErr is
// not nil and starts with "insufficient balance".
func Sub(conn *sqlite.Conn,
	adr *factom.FAAddress, sub uint64) (id int64, txErr, err error) {
	if sub == 0 {
		// Allow tx's with zeros to result in an INSERT.
		id, err = Add(conn, adr, 0)
		return id, nil, err
	}
	id, err = SelectID(conn, adr)
	if err != nil {
		if err.Error() == sqlitexNoResultsErr {
			return id, fmt.Errorf("insufficient balance: %v", adr), nil
		}
		return id, nil, err
	}
	if id < 0 {
		return id, fmt.Errorf("insufficient balance: %v", adr), nil
	}
	stmt := conn.Prep(
		`UPDATE "addresses" SET "balance" = "balance" - ? WHERE "rowid" = ?;`)
	stmt.BindInt64(1, int64(sub))
	stmt.BindInt64(2, id)
	if _, err := stmt.Step(); err != nil {
		if sqlite.ErrCode(err) == sqlite.SQLITE_CONSTRAINT_CHECK {
			return id, fmt.Errorf("insufficient balance: %v", adr), nil
		}
		return id, nil, err
	}
	if conn.Changes() == 0 {
		panic("no balances updated")
	}
	return id, nil, nil
}

// SelectIDBalance returns the row id and balance for the given adr.
func SelectIDBalance(conn *sqlite.Conn,
	adr *factom.FAAddress) (adrID int64, bal uint64, err error) {
	adrID = -1
	stmt := conn.Prep(`SELECT "id", "balance" FROM "addresses" WHERE "address" = ?;`)
	defer stmt.Reset()
	stmt.BindBytes(1, adr[:])
	hasRow, err := stmt.Step()
	if err != nil {
		return
	}
	if !hasRow {
		return
	}
	adrID = stmt.ColumnInt64(0)
	bal = uint64(stmt.ColumnInt64(1))
	return
}

// SelectID returns the row id for the given adr.
func SelectID(conn *sqlite.Conn, adr *factom.FAAddress) (int64, error) {
	stmt := conn.Prep(`SELECT "id" FROM "addresses" WHERE "address" = ?;`)
	stmt.BindBytes(1, adr[:])
	return sqlitex.ResultInt64(stmt)
}

// SelectCount returns the number of rows in "addresses". If nonZeroOnly is
// true, then only count the addresses with a non zero balance.
func SelectCount(conn *sqlite.Conn, nonZeroOnly bool) (int64, error) {
	stmt := conn.Prep(`SELECT count(*) FROM "addresses" WHERE "id" != 1
                AND (? OR "balance" > 0);`)
	stmt.BindBool(1, !nonZeroOnly)
	return sqlitex.ResultInt64(stmt)
}
