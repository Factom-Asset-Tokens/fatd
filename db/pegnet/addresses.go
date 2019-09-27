package pegnet

import (
	"fmt"
	"strings"

	"crawshaw.io/sqlite"
	"crawshaw.io/sqlite/sqlitex"
	"github.com/Factom-Asset-Tokens/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat/fat2"
)

// CreateTable is a SQL string that creates the "pn_addresses" table.
const CreateTableAddresses = `CREATE TABLE "pn_addresses" (
        "id"            INTEGER PRIMARY KEY,
        "address"       BLOB NOT NULL UNIQUE,
        "peg_balance"   INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("peg_balance" >= 0),
        "pusd_balance"  INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("pusd_balance" >= 0),
        "peur_balance"  INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("peur_balance" >= 0),
        "pjpy_balance"  INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("pjpy_balance" >= 0),
        "pgbp_balance"  INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("pgbp_balance" >= 0),
        "pcad_balance"  INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("pcad_balance" >= 0),
        "pchf_balance"  INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("pchf_balance" >= 0),
        "pinr_balance"  INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("pinr_balance" >= 0),
        "psgd_balance"  INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("psgd_balance" >= 0),
        "pcny_balance"  INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("pcny_balance" >= 0),
        "phkd_balance"  INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("phkd_balance" >= 0),
        "pkrw_balance"  INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("pkrw_balance" >= 0),
        "pbrl_balance"  INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("pbrl_balance" >= 0),
        "pphp_balance"  INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("pphp_balance" >= 0),
        "pmxn_balance"  INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("pmxn_balance" >= 0),
        "pxau_balance"  INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("pxau_balance" >= 0),
        "pxag_balance"  INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("pxag_balance" >= 0),
        "pxbt_balance"  INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("pxbt_balance" >= 0),
        "peth_balance"  INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("peth_balance" >= 0),
        "pltc_balance"  INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("pltc_balance" >= 0),
        "prvn_balance"  INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("prvn_balance" >= 0),
        "pxbc_balance"  INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("pxbc_balance" >= 0),
        "pfct_balance"  INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("pfct_balance" >= 0),
        "pbnb_balance"  INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("pbnb_balance" >= 0),
        "pxlm_balance"  INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("pxlm_balance" >= 0),
        "pada_balance"  INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("pada_balance" >= 0),
        "pxmr_balance"  INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("pxmr_balance" >= 0),
        "pdas_balance"  INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("pdas_balance" >= 0),
        "pzec_balance"  INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("pzec_balance" >= 0),
        "pdcr_balance"  INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("pdcr_balance" >= 0),
);
`

// Add adds add to the typed balance of adr, creating a new row in "pn_addresses" if it
// does not exist. If successful, the row id of adr is returned.
func Add(conn *sqlite.Conn, adr *factom.FAAddress, ticker fat2.PTicker, add uint64) (int64, error) {
	stmtStringFmt := `INSERT INTO "pn_addresses"
                ("address", "%[1]s_balance") VALUES (?, ?)
                ON CONFLICT("address") DO
                UPDATE SET "%[1]s_balance" = "%[1]s_balance" + "excluded"."%[1]s_balance";`
	tickerLower := strings.ToLower(ticker.String())
	stmt := conn.Prep(fmt.Sprintf(stmtStringFmt, tickerLower))
	stmt.BindBytes(1, adr[:])
	stmt.BindInt64(2, int64(add))
	_, err := stmt.Step()
	if err != nil {
		return -1, err
	}
	return SelectID(conn, adr)
}

const sqlitexNoResultsErr = "sqlite: statement has no results"

// Sub subtracts sub from the typed balance of adr creating the row in "addresses" if
// it does not exist and sub is 0. If successful, the row id of adr is
// returned. If subtracting sub would result in a negative balance, txErr is
// not nil and starts with "insufficient balance".
func Sub(conn *sqlite.Conn,
	adr *factom.FAAddress, ticker fat2.PTicker, sub uint64) (id int64, txErr, err error) {
	if sub == 0 {
		// Allow tx's with zeros to result in an INSERT.
		id, err = Add(conn, adr, ticker, 0)
		return id, nil, err
	}
	id, err = SelectID(conn, adr)
	if err != nil {
		if err.Error() == sqlitexNoResultsErr {
			return id, fmt.Errorf("insufficient balance: %v", adr), nil
		}
		return id, nil, err
	}
	if id < 0 {
		return id, fmt.Errorf("insufficient balance: %v", adr), nil
	}

	stmtStringFmt := `UPDATE pn_addresses SET %[1]s_balance = %[1]s_balance - ? WHERE rowid = ?;`
	tickerLower := strings.ToLower(ticker.String())
	stmt := conn.Prep(fmt.Sprintf(stmtStringFmt, tickerLower))
	stmt.BindInt64(1, int64(sub))
	stmt.BindInt64(2, id)
	if _, err := stmt.Step(); err != nil {
		if sqlite.ErrCode(err) == sqlite.SQLITE_CONSTRAINT_CHECK {
			return id, fmt.Errorf("insufficient balance: %v", adr), nil
		}
		return id, nil, err
	}
	if conn.Changes() == 0 {
		panic("no balances updated")
	}
	return id, nil, nil
}

// SelectIDTypedBalance returns the row id and balance for the given adr and ticker.
func SelectIDTypedBalance(conn *sqlite.Conn,
	adr *factom.FAAddress, ticker fat2.PTicker) (adrID int64, bal uint64, err error) {
	adrID = -1
	stmtStringFmt := `SELECT "id", "%s_balance" FROM "pn_addresses" WHERE "address" = ?;`
	tickerLower := strings.ToLower(ticker.String())
	stmt := conn.Prep(fmt.Sprintf(stmtStringFmt, tickerLower))
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

// SelectID returns the row id for the given adr.
func SelectID(conn *sqlite.Conn, adr *factom.FAAddress) (int64, error) {
	stmt := conn.Prep(`SELECT "id" FROM "pn_addresses" WHERE "address" = ?;`)
	stmt.BindBytes(1, adr[:])
	return sqlitex.ResultInt64(stmt)
}

// SelectCount returns the number of rows in "addresses". If nonZeroOnly is
// true, then only count the addresses with a non zero balance.
func SelectCount(conn *sqlite.Conn, nonZeroOnly bool) (int64, error) {
	stmt := conn.Prep(`SELECT count(*) FROM "pn_addresses" WHERE "id" != 1
                AND (? OR "balance" > 0);`)
	stmt.BindBool(1, !nonZeroOnly)
	return sqlitex.ResultInt64(stmt)
}
