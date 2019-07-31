package db

import (
	"fmt"

	"crawshaw.io/sqlite"
	"github.com/Factom-Asset-Tokens/fatd/factom"
)

// AddressAdd adds add to the balance of adr, if it exists, otherwise it
// inserts the address with balance set to add.
func AddressAdd(conn *sqlite.Conn, adr *factom.FAAddress, add uint64) (int64, error) {
	stmt := conn.Prep(`INSERT INTO addresses
                (address, balance) VALUES (?, ?)
                ON CONFLICT(address) DO
                UPDATE SET balance = balance + excluded.balance;`)
	stmt.BindBytes(1, adr[:])
	stmt.BindInt64(2, int64(add))
	_, err := stmt.Step()
	if err != nil {
		return -1, err
	}
	return SelectAddressID(conn, adr)
}

// AddressSub subtracts sub from the balance of adr, only if this does not
// cause balance to be < 0, otherwise an error is returned.
func AddressSub(conn *sqlite.Conn, adr *factom.FAAddress, sub uint64) (int64, error) {
	id, err := SelectAddressID(conn, adr)
	if err != nil {
		return id, err
	}
	if id < 0 {
		return id, fmt.Errorf("insufficient balance: %v", adr)
	}
	stmt := conn.Prep(`UPDATE addresses
                SET balance = balance - ?
                WHERE rowid = ?;`)
	stmt.BindInt64(1, int64(sub))
	stmt.BindInt64(2, id)
	if _, err := stmt.Step(); err != nil {
		if sqlite.ErrCode(err) == sqlite.SQLITE_CONSTRAINT_CHECK {
			return id, fmt.Errorf("insufficient balance: %v", adr)
		}
		return id, err
	}
	if conn.Changes() == 0 {
		panic("no balances updated")
	}
	return id, nil
}

func SelectAddressBalance(conn *sqlite.Conn, adr *factom.FAAddress) (uint64, error) {
	stmt := conn.Prep(`SELECT balance FROM addresses WHERE address = ?;`)
	stmt.BindBytes(1, adr[:])
	hasRow, err := stmt.Step()
	if err != nil {
		return 0, err
	}
	if !hasRow {
		return 0, nil
	}
	return uint64(stmt.ColumnInt64(0)), nil
}
func SelectAddressID(conn *sqlite.Conn, adr *factom.FAAddress) (int64, error) {
	stmt := conn.Prep(`SELECT rowid FROM addresses WHERE address = ?;`)
	stmt.BindBytes(1, adr[:])
	hasRow, err := stmt.Step()
	if err != nil {
		return -1, err
	}
	if !hasRow {
		return -1, nil
	}
	return stmt.ColumnInt64(0), nil
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
