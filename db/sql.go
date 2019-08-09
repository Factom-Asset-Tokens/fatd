package db

import (
	"fmt"
	"strings"

	"crawshaw.io/sqlite"
)

type bindFunc func(*sqlite.Stmt, int) int

type sql struct {
	sql   string
	binds []bindFunc
}

func (sql sql) Prep(conn *sqlite.Conn) *sqlite.Stmt {
	sql.sql += `;`
	stmt := conn.Prep(sql.sql)
	col := 1
	for _, bind := range sql.binds {
		col += bind(stmt, col)
	}
	return stmt
}

func (sql *sql) Append(str string, binds ...bindFunc) {
	sql.sql += str
	sql.binds = append(sql.binds, binds...)
}
func (sql *sql) Bind(n int, binds ...bindFunc) {
	str := strings.TrimRight(strings.Repeat("?, ", n), ", ")
	sql.Append(str, binds...)
}

const MaxLimit = 600

func (sql *sql) Paginate(page, limit int64) {
	if limit == 0 || limit > MaxLimit {
		limit = MaxLimit
	}
	sql.Append(` LIMIT ?, ?`, func(s *sqlite.Stmt, c int) int {
		s.BindInt64(c, page*limit)
		s.BindInt64(c+1, limit)
		return 2
	})
}

func (sql *sql) Order(order string) {
	switch strings.ToLower(order) {
	case "asc", "":
		sql.Append(` ORDER BY "id" ASC`)
	case "desc":
		sql.Append(` ORDER BY "id" DESC`)
	default:
		panic(fmt.Errorf("invalid order: %v", order))
	}
}

func (sql *sql) OrderPaginate(order string, page, limit int64) {
	sql.Order(order)
	sql.Paginate(page, limit)
}
