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
)

const (
	createTableEBlocks = `CREATE TABLE "eblocks" (
        "seq"           INTEGER PRIMARY KEY,
        "key_mr"        BLOB NOT NULL UNIQUE,
        "db_height"     INTEGER NOT NULL UNIQUE,
        "db_key_mr"     BLOB NOT NULL UNIQUE,
        "timestamp"     INTEGER NOT NULL,
        "data"          BLOB NOT NULL
);
`

	createTableEntries = `CREATE TABLE "entries" (
        "id"            INTEGER PRIMARY KEY,
        "eb_seq"        INTEGER NOT NULL,
        "timestamp"     INTEGER NOT NULL,
        "valid"         BOOL NOT NULL DEFAULT FALSE,
        "hash"          BLOB NOT NULL,
        "data"          BLOB NOT NULL,

        FOREIGN KEY("eb_seq") REFERENCES "eblocks"
);
CREATE INDEX "idx_entries_eb_seq" ON "entries"("eb_seq");
CREATE INDEX "idx_entries_hash" ON "entries"("hash");
`

	createTableAddresses = `CREATE TABLE "addresses" (
        "id"            INTEGER PRIMARY KEY,
        "address"       BLOB NOT NULL UNIQUE,
        "balance"       INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("balance" >= 0)
);
`

	createTableAddressTransactions = `CREATE TABLE "address_transactions" (
        "entry_id"      INTEGER NOT NULL,
        "address_id"    INTEGER NOT NULL,
        "to"            BOOL NOT NULL,

        PRIMARY KEY("entry_id", "address_id"),

        FOREIGN KEY("entry_id")   REFERENCES "entries",
        FOREIGN KEY("address_id") REFERENCES "addresses"
);
CREATE INDEX "idx_address_transactions_address_id" ON "address_transactions"("address_id");
`

	createTableNFTokens = `CREATE TABLE "nf_tokens" (
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

	createTableNFTokensTransactions = `CREATE TABLE "nf_token_transactions" (
        "adr_tx_id"     INTEGER NOT NULL,
        "nf_tkn_id"     INTEGER NOT NULL,

        UNIQUE("adr_tx_id", "nf_tkn_id"),

        FOREIGN KEY("nf_tkn_id") REFERENCES "nf_tokens",
        FOREIGN KEY("adr_tx_id") REFERENCES "address_transactions"
);
CREATE INDEX "idx_nf_token_transactions_adr_tx_id" ON
        "nf_token_transactions"("adr_tx_id");
CREATE INDEX "idx_nf_token_transactions_nf_tkn_id" ON
        "nf_token_transactions"("nf_tkn_id");
CREATE VIEW "nf_token_address_transactions" AS
        SELECT "entry_id", "address_id", "nf_tkn_id", "to" FROM
                "address_transactions" AS "adr_tx",
                "nf_token_transactions" AS "tkn_tx"
                        ON "adr_tx"."rowid" = "tkn_tx"."adr_tx_id";
`

	createTableMetadata = `CREATE TABLE "metadata" (
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

	// For the sake of simplicity, all chain DBs use the exact same schema,
	// regardless of whether they actually make use of the NFTokens tables.
	chainDBSchema = createTableEBlocks +
		createTableEntries +
		createTableAddresses +
		createTableAddressTransactions +
		createTableNFTokens +
		createTableNFTokensTransactions +
		createTableMetadata
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
			return fmt.Errorf("failed to apply schema: %v", err)
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
