package db

import (
	"fmt"

	"crawshaw.io/sqlite"
	"crawshaw.io/sqlite/sqlitex"
)

const (
	createTableEBlocks = `CREATE TABLE eblocks (
        seq             INTEGER PRIMARY KEY,
        key_mr          BLOB NOT NULL UNIQUE,
        db_height       INTEGER NOT NULL UNIQUE,
        db_key_mr       BLOB NOT NULL UNIQUE,
        timestamp       INTEGER NOT NULL,
        data            BLOB NOT NULL
);
`

	createTableEntries = `CREATE TABLE entries (
        id              INTEGER PRIMARY KEY,
        eb_seq          INTEGER NOT NULL,
        timestamp       INTEGER NOT NULL,
        valid           BOOL NOT NULL DEFAULT FALSE,
        hash            BLOB NOT NULL UNIQUE,
        data            BLOB NOT NULL,

        FOREIGN KEY(eb_seq) REFERENCES eblocks
);
CREATE INDEX idx_entries_eb_seq ON entries(eb_seq);
`

	createTableAddresses = `CREATE TABLE addresses (
        id              INTEGER PRIMARY KEY,
        address         BLOB NOT NULL UNIQUE,
        balance         INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK (balance >= 0)
);
`

	createTableAddressTransactions = `CREATE TABLE address_transactions (
        entry_id        INTEGER NOT NULL,
        address_id      INTEGER NOT NULL,
        sent_to         BOOL NOT NULL,

        PRIMARY KEY(entry_id, address_id),

        FOREIGN KEY(entry_id)   REFERENCES entries,
        FOREIGN KEY(address_id) REFERENCES addresses
);
CREATE INDEX idx_address_transactions_address_id ON address_transactions(address_id);
`

	createTableNFTokens = `CREATE TABLE nf_tokens (
        id                      INTEGER PRIMARY KEY,
        metadata                BLOB,
        creation_entry_id       INTEGER NOT NULL,
        owner_id                INTEGER NOT NULL,

        FOREIGN KEY(creation_entry_id) REFERENCES entries,
        FOREIGN KEY(owner_id) REFERENCES addresses
);
CREATE INDEX idx_nf_tokens_metadata ON nf_tokens(metadata);
CREATE INDEX idx_nf_tokens_owner_id ON nf_tokens(owner_id);
`

	createTableNFTokensTransactions = `CREATE TABLE nf_token_transactions (
        entry_id        INTEGER NOT NULL,
        nf_token_id     INTEGER NOT NULL,
        owner_id        INTEGER NOT NULL,

        PRIMARY KEY(entry_id, nf_token_id),

        FOREIGN KEY(entry_id)    REFERENCES entries,
        FOREIGN KEY(nf_token_id) REFERENCES nf_tokens,
        FOREIGN KEY(owner_id)    REFERENCES addresses
);
CREATE INDEX idx_nf_token_transactions_owner_id ON nf_token_transactions(owner_id);
CREATE INDEX idx_nf_token_transactions_nf_token_id ON nf_token_transactions(nf_token_id);
`

	createTableMetadata = `CREATE TABLE metadata (
        id              INTEGER PRIMARY KEY,
        sync_height     INTEGER,
        sync_db_key_mr  BLOB,
        network_id      BLOB,
        init_entry_id   INTEGER,

        FOREIGN KEY(init_entry_id) REFERENCES entries
);`

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
		return fmt.Errorf("invalid schema: '%v'\n expected: '%#v'",
			fullSchema, schema)
	}
	return nil
}
func getFullSchema(conn *sqlite.Conn) (string, error) {
	const selectSchema = `SELECT sql FROM sqlite_master;`
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
