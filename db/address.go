package db

import (
	"fmt"

	"crawshaw.io/sqlite"
	"crawshaw.io/sqlite/sqlitex"
	"github.com/Factom-Asset-Tokens/factom"
)

func (chain *Chain) addressAdd(adr *factom.FAAddress, add uint64) (int64, error) {
	stmt := chain.Conn.Prep(`INSERT INTO "addresses"
                ("address", "balance") VALUES (?, ?)
                ON CONFLICT("address") DO
                UPDATE SET "balance" = "balance" + "excluded"."balance";`)
	stmt.BindBytes(1, adr[:])
	stmt.BindInt64(2, int64(add))
	_, err := stmt.Step()
	if err != nil {
		return -1, err
	}
	return SelectAddressID(chain.Conn, adr)
}

func (chain *Chain) addressSub(adr *factom.FAAddress, sub uint64) (int64, error, error) {
	if sub == 0 {
		// Allow tx's with zeros to result in an INSERT.
		id, err := chain.addressAdd(adr, 0)
		return id, nil, err
	}
	id, err := SelectAddressID(chain.Conn, adr)
	if err != nil {
		if err.Error() == sqlitexNoResultsErr {
			return id, fmt.Errorf("insufficient balance: %v", adr), nil
		}
		return id, nil, err
	}
	if id < 0 {
		return id, fmt.Errorf("insufficient balance: %v", adr), nil
	}
	stmt := chain.Conn.Prep(
		`UPDATE addresses SET balance = balance - ? WHERE rowid = ?;`)
	stmt.BindInt64(1, int64(sub))
	stmt.BindInt64(2, id)
	if _, err := stmt.Step(); err != nil {
		if sqlite.ErrCode(err) == sqlite.SQLITE_CONSTRAINT_CHECK {
			return id, fmt.Errorf("insufficient balance: %v", adr), nil
		}
		return id, nil, err
	}
	if chain.Conn.Changes() == 0 {
		panic("no balances updated")
	}
	return id, nil, nil
}

var sqlitexNoResultsErr = "sqlite: statement has no results"

func SelectAddressIDBalance(conn *sqlite.Conn,
	adr *factom.FAAddress) (adrID int64, bal uint64, err error) {
	adrID = -1
	stmt := conn.Prep(`SELECT "id", "balance" FROM "addresses" WHERE "address" = ?;`)
	defer stmt.Reset()
	stmt.BindBytes(1, adr[:])
	hasRow, err := stmt.Step()
	if err != nil {
		return
	}
	if !hasRow {
		return
	}
	adrID = stmt.ColumnInt64(0)
	bal = uint64(stmt.ColumnInt64(1))
	return
}

func SelectAddressID(conn *sqlite.Conn, adr *factom.FAAddress) (int64, error) {
	stmt := conn.Prep(`SELECT "id" FROM "addresses" WHERE "address" = ?;`)
	stmt.BindBytes(1, adr[:])
	return sqlitex.ResultInt64(stmt)
}

func SelectAddressCount(conn *sqlite.Conn, nonZeroOnly bool) (int64, error) {
	stmt := conn.Prep(`SELECT count(*) FROM "addresses" WHERE "id" != 1
                AND (? OR "balance" > 0);`)
	stmt.BindBool(1, !nonZeroOnly)
	return sqlitex.ResultInt64(stmt)
}

func (chain *Chain) insertAddressTransaction(
	adrID int64, entryID int64, to bool) (int64, error) {
	stmt := chain.Conn.Prep(`INSERT INTO "address_transactions"
                ("address_id", "entry_id", "to") VALUES
                (?, ?, ?)`)
	stmt.BindInt64(1, adrID)
	stmt.BindInt64(2, entryID)
	stmt.BindBool(3, to)
	_, err := stmt.Step()
	if err != nil {
		return -1, err
	}
	return chain.Conn.LastInsertRowID(), nil
}
