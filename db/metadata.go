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
	stmt := conn.Prep(`UPDATE metadata SET
                (id, sync_height, sync_db_key_mr) = (?, ?, ?);`)
	stmt.BindInt64(1, 0)
	stmt.BindInt64(2, int64(height))
	stmt.BindBytes(3, dbKeyMR[:])
	_, err := stmt.Step()
	return err
}

func SaveInitEntryID(conn *sqlite.Conn, entryID int64) error {
	stmt := conn.Prep(`UPDATE metadata SET init_entry_id = ?;`)
	stmt.BindInt64(1, entryID)
	_, err := stmt.Step()
	return err

}

func SelectMetadata(conn *sqlite.Conn) (int64, uint32, factom.Bytes32, [4]byte, error) {
	var dbKeyMR factom.Bytes32
	var networkID [4]byte
	stmt := conn.Prep(`SELECT sync_height, sync_db_key_mr, network_id, init_entry_id
                FROM metadata;`)
	hasRow, err := stmt.Step()
	if err != nil {
		return -1, 0, dbKeyMR, networkID, err
	}
	if !hasRow {
		return -1, 0, dbKeyMR, networkID, fmt.Errorf("no saved metadata")
	}

	if stmt.ColumnBytes(1, dbKeyMR[:]) != len(dbKeyMR) {
		return -1, 0, dbKeyMR, networkID,
			fmt.Errorf("invalid sync_db_key_mr length")
	}

	if stmt.ColumnBytes(2, networkID[:]) != len(networkID) {
		return -1, 0, dbKeyMR, networkID, fmt.Errorf("invalid network_id length")
	}

	var initEntryID int64 = -1
	if stmt.ColumnType(3) != sqlite.SQLITE_NULL {
		initEntryID = stmt.ColumnInt64(3)
	}

	return initEntryID, uint32(stmt.ColumnInt64(0)), dbKeyMR, networkID, nil
}
