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
	"github.com/Factom-Asset-Tokens/factom/fat"
	"github.com/Factom-Asset-Tokens/fatd/internal/db/address"
	"github.com/Factom-Asset-Tokens/fatd/internal/db/contract"
	"github.com/Factom-Asset-Tokens/fatd/internal/db/eblock"
	"github.com/Factom-Asset-Tokens/fatd/internal/db/entry"
	"github.com/Factom-Asset-Tokens/fatd/internal/db/metadata"
	"github.com/Factom-Asset-Tokens/fatd/internal/db/nftoken"
)

const (
	// For the sake of simplicity, all chain DBs use the exact same schema,
	// regardless of whether they actually make use of the NFTokens tables.
	chainDBSchema = eblock.CreateTable +
		entry.CreateTable +
		address.CreateTable +
		address.CreateTableTransaction +
		address.CreateTableContract +
		nftoken.CreateTable +
		nftoken.CreateTableTransaction +
		metadata.CreateTableFactomChain +
		metadata.CreateTableFATChain +
		contract.CreateTable

	currentDBVersion = 2
)

var migrations = []func(*sqlite.Conn) error{
	func(conn *sqlite.Conn) error {
		err := sqlitex.ExecScript(conn,
			metadata.CreateTableFactomChain+
				metadata.CreateTableFATChain+`
ALTER TABLE "eblocks" RENAME TO "eblock";
ALTER TABLE "entries" RENAME TO "entry";
ALTER TABLE "addresses" RENAME TO "address";
ALTER TABLE "address_transactions" RENAME TO "address_tx";
ALTER TABLE "nf_tokens" RENAME TO "nftoken";
ALTER TABLE "nf_token_transactions" RENAME TO "nftoken_tx";
ALTER TABLE "nftoken_tx" RENAME COLUMN "adr_tx_id" TO
        "address_tx_id";
ALTER TABLE "nftoken_tx" RENAME COLUMN "nf_tkn_id" TO
        "nftoken_id";

DROP VIEW "nf_token_address_transactions";
CREATE VIEW "nftoken_address_tx" AS
        SELECT "entry_id", "address_id", "nftoken_id", "to" FROM
                "address_tx" AS "adr_tx",
                "nftoken_tx" AS "tkn_tx"
                        ON "adr_tx"."rowid" = "tkn_tx"."address_tx_id";

DROP VIEW "nf_tokens_addresses";
CREATE VIEW "nftoken_address" AS
        SELECT "nftoken"."id" AS "id",
                "metadata",
                "hash" AS "creation_hash",
                "address" AS "owner" FROM
                        "nftoken", "address", "entry" ON
                                "owner_id" = "address"."id" AND
                                "creation_entry_id" = "entry"."id";

DROP INDEX "idx_address_transactions_address_id";
CREATE INDEX "idx_address_tx_address_id" ON "address_tx"("address_id");

DROP INDEX "idx_address_transactions_entry_id";
CREATE INDEX "idx_address_tx_entry_id" ON "address_tx"("entry_id");

DROP INDEX "idx_entries_eb_seq";
CREATE INDEX "idx_entry_eb_seq" ON "entry"("eb_seq");

DROP INDEX "idx_entries_hash";
CREATE INDEX "idx_entry_hash" ON "entry"("hash");

DROP INDEX "idx_nf_token_transactions_adr_tx_id";
CREATE INDEX "idx_nftoken_tx_address_tx_id" ON "nftoken_tx"("address_tx_id");

DROP INDEX "idx_nf_token_transactions_nf_tkn_id";
CREATE INDEX "idx_nftoken_tx_nftoken_id" ON "nftoken_tx"("nftoken_id");

DROP INDEX "idx_nf_tokens_metadata";
CREATE INDEX "idx_nftoken_metadata" ON nftoken("metadata");

DROP INDEX "idx_nf_tokens_owner_id";
CREATE INDEX "idx_nftoken_owner_id" ON nftoken("owner_id");`)
		if err != nil {
			return err
		}

		e, err := entry.SelectByID(conn, 1)
		if err != nil {
			return err
		}

		err = sqlitex.ExecTransient(conn,
			`INSERT INTO "factom_chain" ("chain_id", "network_id")
                                VALUES (?,
                                (SELECT "network_id" FROM "metadata"));`,
			nil, e.ChainID[:])
		if err != nil {
			return err
		}

		tokenID, issuerID := fat.ParseTokenIssuer(e.ExtIDs)
		err = sqlitex.ExecTransient(conn,
			`INSERT INTO "fat_chain" ("token", "issuer") VALUES (?, ?);`,
			nil, tokenID, issuerID[:])
		if err != nil {
			return err
		}

		return sqlitex.ExecScript(conn, `
UPDATE "factom_chain" SET
        (
                "sync_height",
                "sync_db_key_mr"
        ) = (SELECT
                "sync_height",
                "sync_db_key_mr"
        FROM "metadata");
UPDATE "fat_chain" SET
        (
                "id_key_entry_data",
                "id_key_height",
                "init_entry_id",
                "num_issued"
        ) = (SELECT
                "id_key_entry",
                "id_key_height",
                "init_entry_id",
                "num_issued"
        FROM "metadata");
DROP TABLE "metadata";`)
	},
	func(conn *sqlite.Conn) error {
		return sqlitex.ExecScript(conn,
			contract.CreateTable+
				address.CreateTableContract)
	},
}

func init() {
	if len(migrations) != currentDBVersion {
		panic("len(migrations) != currentDBVersion")
	}
}

func applyMigrations(conn *sqlite.Conn) (err error) {

	empty, err := isEmpty(conn)
	if err != nil {
		return
	}
	if empty {
		if err = sqlitex.ExecScript(conn, chainDBSchema); err != nil {
			return
		}
		return updateDBVersion(conn)
	}

	version, err := getDBVersion(conn)
	if err != nil {
		return
	}
	if int(version) == len(migrations) {
		return nil
	}
	if int(version) > len(migrations) {
		return fmt.Errorf("no migration exists for DB version: %v", version)
	}

	// Always VACUUM after a successful migration.
	defer func() {
		if err != nil {
			return
		}
		stmt, _, err := conn.PrepareTransient(`VACUUM;`)
		if err != nil {
			panic(err)
		}
		defer stmt.Finalize()
		if _, err := stmt.Step(); err != nil {
			panic(err)
		}
	}()

	defer sqlitex.Save(conn)(&err)

	for i, migration := range migrations[version:] {
		version := int(version) + i
		fmt.Printf("running migration: %v -> %v\n", version, version+1)
		if err = migration(conn); err != nil {
			return
		}
	}
	return updateDBVersion(conn)
}

func isEmpty(conn *sqlite.Conn) (bool, error) {
	var count int
	err := sqlitex.ExecTransient(conn, `SELECT count(*) from "sqlite_master";`,
		func(stmt *sqlite.Stmt) error {
			count = stmt.ColumnInt(0)
			return nil
		})
	return count == 0, err
}

func getDBVersion(conn *sqlite.Conn) (int64, error) {
	var version int64
	err := sqlitex.ExecTransient(conn, `PRAGMA user_version;`,
		func(stmt *sqlite.Stmt) error {
			version = stmt.ColumnInt64(0)
			return nil
		})
	return version, err
}

func updateDBVersion(conn *sqlite.Conn) error {
	return sqlitex.ExecScript(conn, fmt.Sprintf(`PRAGMA user_version = %v;`,
		currentDBVersion))
}
