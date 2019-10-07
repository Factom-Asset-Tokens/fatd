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

package db

import (
	"fmt"

	"crawshaw.io/sqlite"
	"crawshaw.io/sqlite/sqlitex"
	"github.com/Factom-Asset-Tokens/fatd/db/addresses"
	"github.com/Factom-Asset-Tokens/fatd/db/eblocks"
	"github.com/Factom-Asset-Tokens/fatd/db/entries"
	"github.com/Factom-Asset-Tokens/fatd/db/metadata"
	"github.com/Factom-Asset-Tokens/fatd/db/nftokens"
)

const (
	// For the sake of simplicity, all chain DBs use the exact same schema,
	// regardless of whether they actually make use of the NFTokens tables.
	chainDBSchema = eblocks.CreateTable +
		entries.CreateTable +
		addresses.CreateTable +
		addresses.CreateTableTransactions +
		nftokens.CreateTable +
		nftokens.CreateTableTransactions +
		metadata.CreateTable
)

// validateOrApplySchema compares schema with the database connected to by
// conn. If the database has no schema, the schema is applied. Otherwise an
// error will be returned if the schema is not an exact match.
func validateOrApplySchema(conn *sqlite.Conn, schema string) error {
	fullSchema, err := getFullSchema(conn)
	if err != nil {
		return err
	}
	if len(fullSchema) == 0 {
		if err := sqlitex.ExecScript(conn, schema); err != nil {
			return fmt.Errorf("failed to apply schema: %w", err)
		}
		return nil
	}
	if fullSchema != schema {
		return fmt.Errorf("invalid schema: %v\n expected: %v",
			fullSchema, schema)
	}
	return nil
}
func getFullSchema(conn *sqlite.Conn) (string, error) {
	const selectSchema = `SELECT "sql" FROM "sqlite_master";`
	var schema string
	err := sqlitex.ExecTransient(conn, selectSchema,
		func(stmt *sqlite.Stmt) error {
			// Concatenate all non-empty table schemas.
			if tableSchema := stmt.ColumnText(0); len(tableSchema) > 0 {
				schema += tableSchema + ";\n"
			}
			return nil
		})
	if err != nil {
		return "", err
	}
	return schema, nil
}
