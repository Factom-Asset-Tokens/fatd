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

// Package nftokens provides functions and SQL framents for working with the
// "nf_tokens" table, which stores fat.NFToken with owner, creation id, and
// metadata.

package pegnet

import (
	"fmt"

	"crawshaw.io/sqlite"
	"github.com/Factom-Asset-Tokens/factom"
)

const CreateTableRate = `CREATE TABLE "pn_rate" (
        "eb_seq" INTEGER NOT NULL,
        "token" TEXT,
        "value" INTEGER,
        
        UNIQUE("eb_seq", "token")

        FOREIGN KEY("eb_seq") REFERENCES "eblocks"
);
`

func InsertRate(conn *sqlite.Conn, eb factom.EBlock, token string, value uint64) error {
	stmt := conn.Prep(`INSERT INTO "pn_rate"
                ("eb_seq", "token", "value") VALUES (?, ?, ?);`)
	stmt.BindInt64(1, int64(eb.Sequence))
	stmt.BindText(2, token)
	stmt.BindInt64(3, int64(value))
	if _, err := stmt.Step(); err != nil {
		if sqlite.ErrCode(err) == sqlite.SQLITE_CONSTRAINT_UNIQUE {
			return fmt.Errorf("Rate{%d-%s} already exists", eb.Sequence, token)
		}
		return err
	}

	return nil
}
