package db

import (
	"fmt"

	"crawshaw.io/sqlite"
	"github.com/Factom-Asset-Tokens/fatd/factom"
)

// AddressAdd adds add to the balance of adr, if it exists, otherwise it
// inserts the address with balance set to add.
func AddressAdd(conn *sqlite.Conn, adr *factom.FAAddress, add uint64) error {
	stmt := conn.Prep(`INSERT INTO addresses
(address, balance) VALUES (?, ?)
ON CONFLICT(address) DO
UPDATE SET balance = balance + ?;`)
	stmt.BindBytes(1, adr[:])
	stmt.BindInt64(2, int64(add))
	stmt.BindInt64(3, int64(add))
	_, err := stmt.Step()
	return err
}

// AddressSub subtracts sub from the balance of adr, only if this does not
// cause balance to be < 0, otherwise an error is returned.
func AddressSub(conn *sqlite.Conn, adr *factom.FAAddress, sub uint64) error {
	stmt := conn.Prep(`UPDATE addresses
SET balance = balance - ?
WHERE address = ?;`)
	stmt.BindInt64(1, int64(sub))
	stmt.BindBytes(2, adr[:])
	if _, err := stmt.Step(); err != nil {
		return err
	}
	if conn.Changes() == 0 {
		return fmt.Errorf("CHECK constraint failed: insufficient balance")
	}
	return nil
}

// SelectAddress returns the id and balance for the given adr.
func SelectAddress(conn *sqlite.Conn, adr *factom.FAAddress) (int64, uint64, error) {
	stmt := conn.Prep(`SELECT id, balance FROM addresses WHERE address = ?;`)
	stmt.BindBytes(1, adr[:])
	if _, err := stmt.Step(); err != nil {
		return 0, 0, err
	}
	return stmt.ColumnInt64(0), uint64(stmt.ColumnInt64(1)), nil
}

func InsertAddressTransaction(conn *sqlite.Conn,
	adrID int64, entryID int64, to bool) error {
	stmt := conn.Prep(`INSERT INTO address_transactions
(address_id, entry_id, sent_to) VALUES
(?, ?, ?)`)
	stmt.BindInt64(1, adrID)
	stmt.BindInt64(2, entryID)
	stmt.BindBool(3, to)
	_, err := stmt.Step()
	return err
}
