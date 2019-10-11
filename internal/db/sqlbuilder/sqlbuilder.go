// MIT License
//
// Copyright 2018 Canonical Ledgers, LLC
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
// IN THE SOFTWARE.

package sqlbuilder

import (
	"fmt"
	"strings"

	"crawshaw.io/sqlite"
)

// BindFunc is a function that binds one or more values to a sqlite.Stmt. A
// valid BindFunc must use startParam as the first param index for any binds
// and must return the total number of parameters bound.
type BindFunc func(stmt *sqlite.Stmt, startParam int) int

// SQLBuilder prepares a sqlite.Stmt that has variable numbers of Binds.
type SQLBuilder struct {
	strings.Builder
	Binds []BindFunc

	LimitMax uint // Max limit allowed by Paginate()
}

// Prep prepares a sqlite.Stmt on conn using s.String() with a trailing `;` as
// the SQL, and then sequentially calls all s.Binds using the Stmt and the
// appropriate startParam. The Stmt is returned ready for its first Step().
//
// If the total number of binds reported by the Binds differs from the total
// number of params reported by the Stmt, then Prep panics.
func (s *SQLBuilder) Prep(conn *sqlite.Conn) *sqlite.Stmt {
	s.WriteString(`;`)
	stmt := conn.Prep(s.String())
	param := 1
	for _, bind := range s.Binds {
		param += bind(stmt, param)
	}
	if param-1 != stmt.BindParamCount() {
		panic(fmt.Errorf(
			"reported bind count (%v) does not match bind param count (%v)"+
				"\nSQL:%q",
			s.String()+`;`, param-1, stmt.BindParamCount()))
	}

	return stmt
}

// Append sql and any associated binds.
//
// Do not include a `;` in sql.
//
// The sum of the binds return values must equal the number of params (e.g.
// "?") in sql or else s.Prep will panic.
func (s *SQLBuilder) Append(sql string, binds ...BindFunc) {
	s.WriteString(sql)
	s.Binds = append(s.Binds, binds...)
}

// BindNParams appends n comma separated params placeholders (e.g. "?, ?, ... ,
// ?") and append the binds.
//
// Do not include a `;` in sql.
//
// The sum of the binds return values must equal n or else s.Prep will panic.
func (s *SQLBuilder) BindNParams(n int, binds ...BindFunc) {
	str := strings.TrimRight(strings.Repeat("?, ", n), ", ")
	s.Append(str, binds...)
}

// LimitMaxDefault is used in SQLBuilder.Paginate() if SQLBuilder.LimitMax
// equals 0.
var LimitMaxDefault uint = 600

// Paginate appends ` LIMIT ?, ?` and the appropriate page and limit binds.
func (s *SQLBuilder) Paginate(page, limit uint) {
	if s.LimitMax == 0 {
		s.LimitMax = LimitMaxDefault
	}
	if limit == 0 || limit > s.LimitMax {
		limit = s.LimitMax
	}
	s.Append(` LIMIT ?, ?`, func(s *sqlite.Stmt, p int) int {
		s.BindInt64(p, int64((page-1)*limit))
		s.BindInt64(p+1, int64(limit))
		return 2
	})
}

// OrderBy append fmt.Sprintf(` ORDER BY %q %s`, col, order). No binds are
// added.
func (s *SQLBuilder) OrderBy(col, ascDesc string) {
	ascDesc = strings.ToUpper(ascDesc)
	switch ascDesc {
	case "ASC", "DESC":
	case "":
		ascDesc = "ASC"
	default:
		panic(fmt.Errorf("invalid order: %v", ascDesc))
	}
	s.WriteString(fmt.Sprintf(` ORDER BY %q %v`, col, ascDesc))
}

// OrderByPaginate calls s.OrderBy() and then s.Paginate().
func (s *SQLBuilder) OrderByPaginate(col, order string, page, limit uint) {
	s.OrderBy(col, order)
	s.Paginate(page, limit)
}
