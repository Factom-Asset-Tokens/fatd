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

package address

import (
	"fmt"

	"crawshaw.io/sqlite"
	"github.com/Factom-Asset-Tokens/factom"
)

const CreateTableContract = `
CREATE TABLE IF NOT EXISTS %[1]q."address_contract" (
        "address_id"    INTEGER PRIMARY KEY,
        "chainid"       BLOB,
        FOREIGN KEY("address_id") REFERENCES "address"
);
CREATE INDEX IF NOT EXISTS
        %[1]q."idx_address_contract_chainid" ON "address_contract"("chainid");
CREATE VIEW IF NOT EXISTS "temp"."address_contract_all" AS
        SELECT * FROM "temp"."address_contract" WHERE "chainid" IS NOT NULL
        UNION ALL
        SELECT * FROM "main"."address_contract" WHERE "address_id" NOT IN (
                SELECT "address_id" FROM "temp"."address_contract");
`

func InsertContract(conn *sqlite.Conn,
	adrID int64, chainID *factom.Bytes32) (txErr, err error) {
	// Ensure contract does not already exist in "main" or "temp".
	cID, err := SelectContractChainID(conn, adrID)
	if err != nil {
		return nil, err
	}
	if !cID.IsZero() {
		return fmt.Errorf("address already delegated"), nil
	}
	stmt := conn.Prep(`INSERT INTO "temp"."address_contract"
                ("address_id", "chainid") VALUES (?, ?);`)
	i := sqlite.BindIncrementor()
	stmt.BindInt64(i(), adrID)
	stmt.BindBytes(i(), chainID[:])
	if _, err := stmt.Step(); err != nil {
		return nil, err
	}
	return nil, nil
}

func DeleteContract(conn *sqlite.Conn, adrID int64) error {
	stmt := conn.Prep(`
INSERT INTO "temp"."address_contract" VALUES (?, NULL)
        ON CONFLICT("address_id") DO
                UPDATE SET "chainid" = NULL;`)
	stmt.BindInt64(sqlite.BindIndexStart, adrID)

	_, err := stmt.Step()
	if conn.Changes() == 0 {
		panic("no rows updated")
	}

	return err
}

func SelectContractChainID(conn *sqlite.Conn, adrID int64) (factom.Bytes32, error) {
	stmt := conn.Prep(`SELECT "chainid" FROM "temp"."address_contract_all"
        WHERE "address_id" = ?;`)
	stmt.BindInt64(sqlite.BindIndexStart, adrID)

	var chainID factom.Bytes32
	if hasRow, err := stmt.Step(); err != nil || !hasRow {
		return chainID, err
	}

	if stmt.ColumnBytes(sqlite.ColumnIndexStart, chainID[:]) != len(chainID) {
		panic("invalid chainid length")
	}

	return chainID, nil
}
