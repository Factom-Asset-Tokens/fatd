package address

import "crawshaw.io/sqlite"

// CreateTableTransaction is a SQL string that creates the "address_tx" table.
//
// The "address_tx" table has a foreign key reference to the "address" and
// "entry" tables, which must exist first.
const CreateTableTransaction = `CREATE TABLE "address_tx" (
        "entry_id"      INTEGER NOT NULL,
        "address_id"    INTEGER NOT NULL,
        "to"            BOOL NOT NULL,

        PRIMARY KEY("entry_id", "address_id", "to"),

        FOREIGN KEY("entry_id")   REFERENCES "entry",
        FOREIGN KEY("address_id") REFERENCES "address"
);
CREATE INDEX "idx_address_tx_address_id" ON "address_tx"("address_id");
CREATE INDEX "idx_address_tx_entry_id" ON "address_tx"("entry_id");
`

// InsertTransaction inserts a row into "address_tx" relating the adrID with
// the entryID and a transaction direction, to. If successful, the row id for
// the new row in "address_tx" is returned.
func InsertTransaction(conn *sqlite.Conn,
	adrID int64, entryID int64, to bool) (int64, error) {
	stmt := conn.Prep(`INSERT INTO "address_tx"
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
