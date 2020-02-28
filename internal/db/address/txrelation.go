package address

import "crawshaw.io/sqlite"

// CreateTableTransaction is a SQL string that creates the "address_tx" table.
//
// The "address_tx" table has a foreign key reference to the "address" and
// "entry" tables, which must exist first.
const CreateTableTransaction = `
CREATE TABLE IF NOT EXISTS %[1]q."address_tx" (
        "entry_id"      INTEGER NOT NULL,
        "address_id"    INTEGER NOT NULL,
        "to"            BOOL NOT NULL,

        PRIMARY KEY("entry_id", "address_id", "to"),

        FOREIGN KEY("entry_id")   REFERENCES "entry",
        FOREIGN KEY("address_id") REFERENCES "address"
);
CREATE INDEX IF NOT EXISTS
        %[1]q."idx_address_tx_address_id" ON "address_tx"("address_id");
CREATE INDEX IF NOT EXISTS
        %[1]q."idx_address_tx_entry_id" ON "address_tx"("entry_id");
CREATE VIEW IF NOT EXISTS "temp"."address_tx_all" AS
        SELECT "rowid", * FROM "temp"."address_tx"
        UNION ALL
        SELECT "rowid", * FROM "main"."address_tx" ORDER BY "rowid";
`

// InsertTransaction inserts a row into "address_tx" relating the adrID with
// the entryID and a transaction direction, to. If successful, the row id for
// the new row in "address_tx" is returned.
func InsertTransaction(conn *sqlite.Conn,
	adrID int64, entryID int64, to bool) (int64, error) {
	stmt := conn.Prep(`INSERT INTO "temp"."address_tx"
                ("rowid", "address_id", "entry_id", "to") VALUES
                ((SELECT max("rowid")+1 FROM "temp"."address_tx_all"),
                        ?, ?, ?)`)
	i := sqlite.BindIncrementor()
	stmt.BindInt64(i(), adrID)
	stmt.BindInt64(i(), entryID)
	stmt.BindBool(i(), to)
	_, err := stmt.Step()
	if err != nil {
		return -1, err
	}
	return conn.LastInsertRowID(), nil
}
