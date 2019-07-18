package db

import (
	"fmt"

	"crawshaw.io/sqlite"
	"github.com/Factom-Asset-Tokens/fatd/factom"
)

func InsertMetadata(conn *sqlite.Conn,
	height uint32, dbKeyMR *factom.Bytes32, networkID [4]byte) error {
	stmt := conn.Prep(`INSERT INTO metadata
                (id, sync_height, sync_db_key_mr, network_id)
                VALUES (?, ?, ?, ?);`)
	stmt.BindInt64(1, 0)
	stmt.BindInt64(2, int64(height))
	stmt.BindBytes(3, dbKeyMR[:])
	stmt.BindBytes(4, networkID[:])

	_, err := stmt.Step()
	return err

}

func SaveSync(conn *sqlite.Conn, height uint32, dbKeyMR *factom.Bytes32) error {
	stmt := conn.Prep(`UPDATE metadata
                (id, sync_height, sync_db_key_mr)
                VALUES (?, ?, ?);`)
	stmt.BindInt64(1, 0)
	stmt.BindInt64(2, int64(height))
	stmt.BindBytes(3, dbKeyMR[:])
	_, err := stmt.Step()
	return err
}

func SelectMetadata(conn *sqlite.Conn) (uint32, factom.Bytes32, [4]byte, error) {
	var dbKeyMR factom.Bytes32
	var networkID [4]byte
	stmt := conn.Prep(`SELECT sync_height, sync_db_key_mr, network_id
                FROM metadata;`)
	hasRow, err := stmt.Step()
	if err != nil {
		return 0, dbKeyMR, networkID, err
	}
	if !hasRow {
		return 0, dbKeyMR, networkID, fmt.Errorf("no saved metadata")
	}

	if stmt.ColumnBytes(1, dbKeyMR[:]) != len(dbKeyMR) {
		return 0, dbKeyMR, networkID, fmt.Errorf("invalid sync_db_key_mr length")
	}

	if stmt.ColumnBytes(2, networkID[:]) != len(networkID) {
		return 0, dbKeyMR, networkID, fmt.Errorf("invalid network_id length")
	}

	return uint32(stmt.ColumnInt64(0)), dbKeyMR, networkID, nil
}
