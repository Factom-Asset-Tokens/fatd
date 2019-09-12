package db

import (
	"fmt"
	"strings"
	"time"

	"crawshaw.io/sqlite"
	"crawshaw.io/sqlite/sqlitex"
	"github.com/Factom-Asset-Tokens/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat/fat1"
)

func (chain *Chain) InsertEntry(e factom.Entry, ebSeq uint32) (int64, error) {
	data, err := e.MarshalBinary()
	if err != nil {
		panic(fmt.Errorf("factom.Entry{}.MarshalBinary(): %v", err))
	}

	stmt := chain.Conn.Prep(`INSERT INTO "entries"
                ("eb_seq", "timestamp", "hash", "data")
                VALUES (?, ?, ?, ?);`)
	stmt.BindInt64(1, int64(int32(ebSeq))) // Preserve uint32(-1) as -1
	stmt.BindInt64(2, int64(e.Timestamp.Unix()))
	stmt.BindBytes(3, e.Hash[:])
	stmt.BindBytes(4, data)

	if _, err := stmt.Step(); err != nil {
		return -1, err
	}
	return chain.Conn.LastInsertRowID(), nil
}

func (chain *Chain) setEntryValid(id int64) error {
	stmt := chain.Conn.Prep(`UPDATE "entries" SET "valid" = 1 WHERE "id" = ?;`)
	stmt.BindInt64(1, id)
	_, err := stmt.Step()
	if err != nil {
		return err
	}
	if chain.Conn.Changes() == 0 {
		panic("no entries updated")
	}
	return nil
}

const SelectEntryWhere = `SELECT "hash", "data", "timestamp" FROM "entries" WHERE `

func SelectEntry(stmt *sqlite.Stmt) (factom.Entry, error) {
	var e factom.Entry
	hasRow, err := stmt.Step()
	if err != nil {
		return e, err
	}
	if !hasRow {
		return e, nil
	}

	e.Hash = new(factom.Bytes32)
	if stmt.ColumnBytes(0, e.Hash[:]) != len(e.Hash) {
		panic("invalid hash length")
	}

	data := make([]byte, stmt.ColumnLen(1))
	stmt.ColumnBytes(1, data)
	if err := e.UnmarshalBinary(data); err != nil {
		panic(fmt.Errorf("factom.Entry{}.UnmarshalBinary(%x): %v",
			data, err))
	}

	e.Timestamp = time.Unix(stmt.ColumnInt64(2), 0)

	return e, nil
}

func SelectEntryByID(conn *sqlite.Conn, id int64) (factom.Entry, error) {
	stmt := conn.Prep(SelectEntryWhere + `"id" = ?;`)
	stmt.BindInt64(1, id)
	defer stmt.Reset()
	return SelectEntry(stmt)
}

func SelectEntryByHash(conn *sqlite.Conn, hash *factom.Bytes32) (factom.Entry, error) {
	stmt := conn.Prep(SelectEntryWhere + `"hash" = ?;`)
	stmt.BindBytes(1, hash[:])
	defer stmt.Reset()
	return SelectEntry(stmt)
}

func SelectEntryByHashValid(conn *sqlite.Conn, hash *factom.Bytes32) (factom.Entry, error) {
	stmt := conn.Prep(SelectEntryWhere + `"hash" = ? AND "valid" = true;`)
	stmt.BindBytes(1, hash[:])
	defer stmt.Reset()
	return SelectEntry(stmt)
}

func SelectEntryCount(conn *sqlite.Conn, validOnly bool) (int64, error) {
	stmt := conn.Prep(`SELECT count(*) FROM "entries" WHERE (? OR "valid" = true);`)
	stmt.BindBool(1, !validOnly)
	return sqlitex.ResultInt64(stmt)
}

func SelectEntryByAddress(conn *sqlite.Conn, startHash *factom.Bytes32,
	adrs []factom.FAAddress, nfTkns fat1.NFTokens,
	toFrom, order string,
	page, limit uint64) ([]factom.Entry, error) {
	if page == 0 {
		return nil, fmt.Errorf("invalid page")
	}
	var sql sql
	sql.Append(SelectEntryWhere + `"valid" = true`)
	if startHash != nil {
		sql.Append(` AND "id" >= (SELECT "id" FROM "entries" WHERE "hash" = ?)`,
			func(s *sqlite.Stmt, p int) int {
				s.BindBytes(p, startHash[:])
				return 1
			})
	}
	var to bool
	switch strings.ToLower(toFrom) {
	case "to":
		to = true
	case "from", "":
	default:
		panic(fmt.Errorf("invalid toFrom: %v", toFrom))
	}
	if len(nfTkns) > 0 {
		sql.Append(` AND "id" IN (
                                SELECT "entry_id" FROM "nf_token_address_transactions"
                                        WHERE "nf_tkn_id" IN (`) // 2 open (
		sql.Bind(len(nfTkns), func(s *sqlite.Stmt, p int) int {
			i := 0
			for nfTkn := range nfTkns {
				s.BindInt64(p+i, int64(nfTkn))
				i++
			}
			return len(nfTkns)
		})
		sql.Append(`)`) // 1 open (
		if len(adrs) > 0 {
			sql.Append(` AND "address_id" IN (
                                SELECT "id" FROM "addresses"
                                        WHERE "address" IN (`) // 3 open (
			sql.Bind(len(adrs), func(s *sqlite.Stmt, p int) int {
				for i, adr := range adrs {
					s.BindBytes(p+i, adr[:])
				}
				return len(adrs)
			})
			sql.Append(`))`) // 1 open (
		}
		if len(toFrom) > 0 {
			sql.Append(` AND "to" = ?`, func(s *sqlite.Stmt, p int) int {
				s.BindBool(p, to)
				return 1
			})
		}
		sql.Append(`)`) // 0 open {
	} else if len(adrs) > 0 {
		sql.Append(` AND "id" IN (
                                SELECT "entry_id" FROM "address_transactions"
                                        WHERE "address_id" IN (
                                                SELECT "id" FROM "addresses"
                                                        WHERE "address" IN (`) // 3 open (

		sql.Bind(len(adrs), func(s *sqlite.Stmt, p int) int {
			for i, adr := range adrs {
				s.BindBytes(p+i, adr[:])
			}
			return len(adrs)
		})
		sql.Append(`))`) // 1 open (
		if len(toFrom) > 0 {
			sql.Append(` AND "to" = ?`, func(s *sqlite.Stmt, p int) int {
				s.BindBool(p, to)
				return 1
			})
		}
		sql.Append(`)`) // 0 open (
	}

	sql.OrderPaginate(order, page, limit)

	stmt := sql.Prep(conn)
	defer stmt.Reset()

	var entries []factom.Entry
	for {
		e, err := SelectEntry(stmt)
		if err != nil {
			return nil, err
		}
		if !e.IsPopulated() {
			break
		}
		entries = append(entries, e)
	}

	return entries, nil
}

func CheckEntryUniquelyValid(conn *sqlite.Conn,
	id int64, hash *factom.Bytes32) (bool, error) {
	stmt := conn.Prep(`SELECT count(*) FROM "entries" WHERE
                "valid" = true AND (? OR "id" < ?) AND "hash" = ?;`)
	stmt.BindBool(1, id > 0)
	stmt.BindInt64(2, id)
	stmt.BindBytes(3, hash[:])
	val, err := sqlitex.ResultInt(stmt)
	if err != nil {
		return false, err
	}
	return val == 0, nil
}

func SelectEntryLatestValid(conn *sqlite.Conn) (factom.Entry, error) {
	stmt := conn.Prep(SelectEntryWhere +
		`"id" = (SELECT max("id") FROM "entries" WHERE "valid" = true);`)
	e, err := SelectEntry(stmt)
	defer stmt.Reset()
	if err != nil {
		return e, err
	}
	return e, nil
}
