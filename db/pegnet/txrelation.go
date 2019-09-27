package pegnet

import "crawshaw.io/sqlite"

// CreateTableTransactions is a SQL string that creates the
// "pn_address_transactions" table.
//
// The "pn_address_transactions" table has a foreign key reference to the
// "pn_addresses" and "entries" tables, which must exist first.
const CreateTableTransactions = `CREATE TABLE "pn_address_transactions" (
        "entry_id"      INTEGER NOT NULL,
        "address_id"    INTEGER NOT NULL,
        "tx_index"      INTEGER NOT NULL,
        "to"            BOOL NOT NULL,

        PRIMARY KEY("entry_id", "address_id"),

        FOREIGN KEY("entry_id")   REFERENCES "entries",
        FOREIGN KEY("address_id") REFERENCES "pn_addresses"
);
CREATE INDEX "idx_address_transactions_address_id" ON "pn_address_transactions"("address_id");
`

// InsertTransactionRelation inserts a row into "pnaddress_transactions" relating
// the adrID with the entryID with the given transaction direction, to. If
// successful, the row id for the new row in "pn_address_transactions" is
// returned.
func InsertTransactionRelation(conn *sqlite.Conn,
	adrID int64, entryID int64, txIndex int64, to bool) (int64, error) {
	stmt := conn.Prep(`INSERT INTO "pn_address_transactions"
                ("address_id", "entry_id", "tx_index", "to") VALUES
                (?, ?, ?, ?)`)
	stmt.BindInt64(1, adrID)
	stmt.BindInt64(2, entryID)
	stmt.BindInt64(3, txIndex)
	stmt.BindBool(4, to)
	_, err := stmt.Step()
	if err != nil {
		return -1, err
	}
	return conn.LastInsertRowID(), nil
}
