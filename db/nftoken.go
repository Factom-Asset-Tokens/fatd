package db

import (
	"encoding/json"

	"crawshaw.io/sqlite"
	"github.com/Factom-Asset-Tokens/fatd/fat/fat1"
)

func SetNFTokenOwner(conn *sqlite.Conn, nfID fat1.NFTokenID, aID, eID int64) error {
	stmt := conn.Prep(`INSERT INTO nf_tokens
                (id, owner_id, creation_entry_id) VALUES (?, ?, ?)
                ON CONFLICT(id) DO
                UPDATE SET owner_id = excluded.owner_id;`)
	stmt.BindInt64(1, int64(nfID))
	stmt.BindInt64(2, aID)
	stmt.BindInt64(3, eID)
	_, err := stmt.Step()
	return err
}

func AttachNFTokenMetadata(conn *sqlite.Conn,
	nfID fat1.NFTokenID, metadata json.RawMessage) error {
	stmt := conn.Prep(`UPDATE nf_tokens
                SET metadata = ? WHERE id = ?;`)
	stmt.BindBytes(1, metadata)
	stmt.BindInt64(2, int64(nfID))
	_, err := stmt.Step()
	if conn.Changes() == 0 {
		panic("no NFTokenID updated")
	}
	return err
}

func InsertNFTokenTransaction(conn *sqlite.Conn,
	nfID fat1.NFTokenID, eID, aID int64) error {
	stmt := conn.Prep(`INSERT INTO nf_token_transactions
                (entry_id, nf_token_id, owner_id) VALUES (?, ?, ?);`)
	stmt.BindInt64(1, eID)
	stmt.BindInt64(2, int64(nfID))
	stmt.BindInt64(3, aID)
	_, err := stmt.Step()
	return err
}
