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

// Package address provides functions and SQL framents for working with the
// "address" table, which stores factom.FAAddress with its balance.
package address

import (
	"fmt"

	"crawshaw.io/sqlite"
	"crawshaw.io/sqlite/sqlitex"
	"github.com/Factom-Asset-Tokens/factom"
)

// Commit updates all main.address balances from temp.address, making them
// permanent and visible to all database connections, and then Resets the temp
// database. This causes the conn's currently held read transaction to become a
// write transaction. It will prevent any other open read transactions from
// being committed. Only one thread should be responsible for Committing
// official changes.
func Commit(conn *sqlite.Conn) error {
	return sqlitex.ExecScript(conn, `
INSERT INTO "main"."address" SELECT * FROM "temp"."address" WHERE true
        ON CONFLICT("id") DO UPDATE SET "balance" = "excluded"."balance";

INSERT INTO "main"."address_tx"("rowid", "entry_id", "address_id", "to")
        SELECT "rowid", * FROM "temp"."address_tx";

DELETE FROM "main"."address_contract" WHERE "address_id" IN (
        SELECT "address_id" FROM "temp"."address_contract" WHERE "chainid" IS NULL);
INSERT INTO "main"."address_contract"
        SELECT * FROM "temp"."address_contract" WHERE "chainid" IS NOT NULL;
DELETE FROM "temp"."address";
DELETE FROM "temp"."address_tx";
DELETE FROM "temp"."address_contract";
`)
}

// CreateTable uses the first fmt argument as the database ("main" or "temp")
// to return a SQL string that creates the %[1]q."address" table.
const CreateTable = `
CREATE TABLE IF NOT EXISTS %[1]q."address" (
        "id"            INTEGER PRIMARY KEY,
        "address"       BLOB NOT NULL UNIQUE,
        "balance"       INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("balance" >= 0)
);
CREATE VIEW IF NOT EXISTS "temp"."address_all" AS
        SELECT * FROM "temp"."address"
        UNION ALL
        SELECT * FROM "main"."address" WHERE
                "id" NOT IN (SELECT "id" FROM "temp"."address");

`

// Add amount to the balance of adr, creating a new row if it does not already
// exist. If successful, the row id of adr is returned.
//
// Add only makes changes to the temp database, which allows it to be called
// concurrently on other conns without blocking serially. However the changes
// will only be visible to this conn until Commit is called.
//
// Add must be called within a read transaction on the conn to ensure a
// consistent view of the state, as this depends on the main database not
// changing after changes to temp have been added.
func Add(conn *sqlite.Conn, adr *factom.FAAddress, amount uint64) (int64, error) {
	return add(conn, adr, int64(amount))
}
func add(conn *sqlite.Conn, adr *factom.FAAddress, amount int64) (int64, error) {
	stmt := conn.Prep(`
INSERT INTO "temp"."address"("id", "address", "balance")
    SELECT "id", $adr, "balance" + $amount FROM (
        SELECT "id", "balance" FROM "main"."address" WHERE "address" = $adr
        UNION ALL
        SELECT max("id")+1 AS "id", 0 AS "balance" FROM "temp"."address_all"
    ) ORDER BY "id" ASC LIMIT 1
    ON CONFLICT("address") DO UPDATE SET "balance" = "balance" + $amount;`)
	stmt.SetBytes("$adr", adr[:])
	stmt.SetInt64("$amount", amount)
	_, err := stmt.Step()
	if err != nil {
		return -1, err
	}
	return SelectID(conn, adr)
}

const sqlitexNoResultsErr = "sqlite: statement has no results"

// Sub subtracts sub from the balance of adr creating the row in "address" if
// it does not exist and sub is 0. If successful, the row id of adr is
// returned. If subtracting sub would result in a negative balance, txErr is
// not nil and starts with "insufficient balance".
func Sub(conn *sqlite.Conn, adr *factom.FAAddress,
	amount uint64) (id int64, txErr, err error) {
	id, err = add(conn, adr, -int64(amount))
	if err != nil {
		if sqlite.ErrCode(err) == sqlite.SQLITE_CONSTRAINT_CHECK {
			return id, fmt.Errorf("insufficient balance: %v", adr), nil
		}
		return id, nil, err
	}
	return id, nil, nil
}

// SelectIDBalance returns the row id and balance for the given adr.
func SelectIDBalance(conn *sqlite.Conn,
	adr *factom.FAAddress) (adrID int64, bal uint64, err error) {
	adrID = -1
	stmt := conn.Prep(`SELECT "id", "balance" FROM "temp"."address_all"
                WHERE "address" = ?;`)
	defer stmt.Reset()
	stmt.BindBytes(sqlite.BindIndexStart, adr[:])
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
	stmt := conn.Prep(`SELECT "id" FROM "temp"."address_all"
		WHERE "address" = ?;`)
	stmt.BindBytes(sqlite.BindIndexStart, adr[:])
	return sqlitex.ResultInt64(stmt)
}

// SelectCount returns the number of rows in "address". If nonZeroOnly is true,
// then only count the address with a non zero balance.
func SelectCount(conn *sqlite.Conn, nonZeroOnly bool) (int64, error) {
	stmt := conn.Prep(`SELECT count(*) FROM "temp"."address_all"
		WHERE "id" != 1 AND (? OR "balance" > 0);`)
	stmt.BindBool(sqlite.BindIndexStart, !nonZeroOnly)
	return sqlitex.ResultInt64(stmt)
}
