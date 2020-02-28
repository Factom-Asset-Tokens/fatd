package nftoken

import (
	"crawshaw.io/sqlite"
	"github.com/Factom-Asset-Tokens/factom/fat1"
)

// CreateTableTransaction is a SQL string that creates the "nftoken_tx" table.
//
// The "nftoken_tx" table has a foreign key reference to the "address_tx" and
// "nftoken" tables, which must exist first.
const CreateTableTransaction = `
CREATE TABLE IF NOT EXISTS "nftoken_tx" (
        "address_tx_id" INTEGER NOT NULL,
        "nftoken_id"    INTEGER NOT NULL,

        PRIMARY KEY("address_tx_id", "nftoken_id"),

        FOREIGN KEY("nftoken_id") REFERENCES "nftoken",
        FOREIGN KEY("address_tx_id") REFERENCES "address_tx"
);
CREATE INDEX IF NOT EXISTS "idx_nftoken_tx_address_tx_id" ON
        "nftoken_tx"("address_tx_id");
CREATE INDEX IF NOT EXISTS "idx_nftoken_tx_nftoken_id" ON
        "nftoken_tx"("nftoken_id");
CREATE VIEW IF NOT EXISTS "nftoken_address_tx" AS
        SELECT "entry_id", "address_id", "nftoken_id", "to" FROM
                "address_tx" AS "adr_tx",
                "nftoken_tx" AS "tkn_tx"
                        ON "adr_tx"."rowid" = "tkn_tx"."address_tx_id";
`

// InsertTransaction inserts a row into "nftoken_entry" relating the given
// adrTxID, a foreign row id from the "address_tx" table, with the given nfID.
func InsertTransaction(conn *sqlite.Conn,
	nfID fat1.NFTokenID, adrEntryID int64) error {
	stmt := conn.Prep(`INSERT INTO "nftoken_tx"
                ("nftoken_id", "address_tx_id") VALUES (?, ?);`)
	stmt.BindInt64(1, int64(nfID))
	stmt.BindInt64(2, adrEntryID)

	_, err := stmt.Step()
	return err
}
