package nftokens

import (
	"crawshaw.io/sqlite"
	"github.com/Factom-Asset-Tokens/fatd/fat/fat1"
)

// CreateTableTransactions is a SQL string that creates the
// "nf_token_transactions" table.
//
// The "nf_token_transactions" table has a foreign key reference to the
// "address_transactions" and "nf_tokens" tables, which must exist first.
const CreateTableTransactions = `CREATE TABLE "nf_token_transactions" (
        "adr_tx_id"     INTEGER NOT NULL,
        "nf_tkn_id"     INTEGER NOT NULL,

        PRIMARY KEY("adr_tx_id", "nf_tkn_id"),

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

// InsertTransactionRelation inserts a row into "nf_token_transactions"
// relating the given adrTxID, a foreign row id from the "address_transactions"
// table, with the given nfID.
func InsertTransactionRelation(conn *sqlite.Conn,
	nfID fat1.NFTokenID, adrTxID int64) error {
	stmt := conn.Prep(`INSERT INTO "nf_token_transactions"
                ("nf_tkn_id", "adr_tx_id") VALUES (?, ?);`)
	stmt.BindInt64(1, int64(nfID))
	stmt.BindInt64(2, adrTxID)

	_, err := stmt.Step()
	return err
}
