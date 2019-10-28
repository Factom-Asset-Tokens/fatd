package addresses

import "crawshaw.io/sqlite"

// CreateTableTransactions is a SQL string that creates the
// "address_transactions" table.
//
// The "address_transactions" table has a foreign key reference to the
// "addresses" and "entries" tables, which must exist first.
const CreateTableTransactions = `CREATE TABLE "address_transactions" (
        "entry_id"      INTEGER NOT NULL,
        "address_id"    INTEGER NOT NULL,
        "to"            BOOL NOT NULL,

        PRIMARY KEY("entry_id", "address_id", "to"),

        FOREIGN KEY("entry_id")   REFERENCES "entries",
        FOREIGN KEY("address_id") REFERENCES "addresses"
);
CREATE INDEX "idx_address_transactions_address_id" ON "address_transactions"("address_id");
CREATE INDEX "idx_address_transactions_entry_id" ON "address_transactions"("entry_id");
`

// InsertTransactionRelation inserts a row into "address_transactions" relating
// the adrID with the entryID with the given transaction direction, to. If
// successful, the row id for the new row in "address_transactions" is
// returned.
func InsertTransactionRelation(conn *sqlite.Conn,
	adrID int64, entryID int64, to bool) (int64, error) {
	stmt := conn.Prep(`INSERT INTO "address_transactions"
                ("address_id", "entry_id", "to") VALUES
                (?, ?, ?)`)
	stmt.BindInt64(1, adrID)
	stmt.BindInt64(2, entryID)
	stmt.BindBool(3, to)
	_, err := stmt.Step()
	if err != nil {
		return -1, err
	}
	return conn.LastInsertRowID(), nil
}
