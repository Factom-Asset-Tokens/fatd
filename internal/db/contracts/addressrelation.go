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

package contracts

import (
	"crawshaw.io/sqlite"
	"github.com/Factom-Asset-Tokens/factom"
)

const CreateTableAddressContracts = `CREATE TABLE "address_contracts" (
        "address_id"    INTEGER PRIMARY KEY,
        "chainid"       BLOB NOT NULL,
        FOREIGN KEY("address_id") REFERENCES "addresses"
);
CREATE INDEX "idx_address_contracts_chainid" ON "address_contracts"("chainid");`

func InsertAddressContract(conn *sqlite.Conn,
	adrID int64, chainID *factom.Bytes32) error {

	stmt := conn.Prep(`INSERT INTO "address_contracts"
                ("address_id", "chainid") VALUES (?, ?);`)
	stmt.BindInt64(1, adrID)
	stmt.BindBytes(2, chainID[:])
	if _, err := stmt.Step(); err != nil {
		return err
	}
	return nil
}

func SelectAddressContract(conn *sqlite.Conn,
	adrID int64) (factom.Bytes32, error) {
	stmt := conn.Prep(`SELECT "chainid" FROM "address_contracts"
                WHERE "address_id" = ?;`)
	stmt.BindInt64(1, adrID)

	var chainID factom.Bytes32
	if hasRow, err := stmt.Step(); err != nil || !hasRow {
		return chainID, err
	}

	if stmt.ColumnBytes(0, chainID[:]) != len(chainID) {
		panic("invalid chainid length")
	}

	return chainID, nil
}

func DeleteAddressContract(conn *sqlite.Conn, adrID int64) error {
	stmt := conn.Prep(`DELETE FROM "address_contracts"
                WHERE "address_id" = ?;`)
	stmt.BindInt64(1, adrID)

	_, err := stmt.Step()
	return err
}
