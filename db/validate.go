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

package db

import (
	"fmt"
	"os"
	"time"

	"crawshaw.io/sqlite/sqlitex"
	"github.com/AdamSLevy/sqlitechangeset"
	"github.com/Factom-Asset-Tokens/factom"
	"github.com/Factom-Asset-Tokens/fatd/db/eblocks"
	"github.com/Factom-Asset-Tokens/fatd/db/entries"
	"github.com/Factom-Asset-Tokens/fatd/fat"
	"github.com/Factom-Asset-Tokens/fatd/flag"
)

func init() {
	sqlitechangeset.AlwaysUseBlob = true
}

// Validate all Entry Hashes and EBlock KeyMRs, as well as the continuity of
// all stored EBlocks and Entries.
//
// This does not validate the validity of the saved DBlock KeyMRs.
func (chain Chain) Validate() (err error) {
	chain.Log.Debug("Validating database...")
	// Validate ChainID...
	read := chain.Pool.Get(nil)
	defer chain.Pool.Put(read)
	write := chain.Conn
	first, err := entries.SelectByID(read, 1)
	if err != nil {
		return err
	}
	if !first.IsPopulated() {
		return fmt.Errorf("no entries")
	}
	if *chain.ID != factom.ComputeChainID(first.ExtIDs) {
		return fmt.Errorf("invalid NameIDs")
	}

	// We will use a session to determine if recomputing state results in
	// any changes. If the state is uncorrupted, the session should have an
	// empty patchset.
	sess, err := write.CreateSession("")
	if err != nil {
		return err
	}
	sess.Attach("entries")
	sess.Attach("addresses")
	sess.Attach("nf_tokens")
	sess.Attach("metadata")
	defer sess.Delete()

	// In case there are any changes, we want to roll back everything. We
	// don't fix corrupted databases, at least not yet.
	defer sqlitex.Save(write)(&err)

	// Completely clear the state, while preserving all chain data.
	sqlitex.ExecScript(write, `
                UPDATE "addresses" SET "balance" = 0;
                DELETE FROM "address_transactions";
                DELETE FROM "nf_tokens";
                DELETE FROM "nf_token_transactions";
                DELETE FROM "eblocks";
                DELETE FROM "entries";
                UPDATE "metadata" SET ("init_entry_id", "num_issued") = (NULL, NULL);
                `)
	chain.NumIssued = 0
	chain.Issuance = fat.Issuance{}
	chain.setApplyFunc()

	eBlockStmt := read.Prep(eblocks.SelectWhere + `true;`) // SELECT all EBlocks.
	defer eBlockStmt.Reset()
	entryStmt := read.Prep(entries.SelectWhere + `true;`) // SELECT all Entries.
	defer entryStmt.Reset()

	var eID int = 1     // Entry ID
	var sequence uint32 // EBlock Sequence
	var prevKeyMR, prevFullHash *factom.Bytes32
	for {
		eb, err := eblocks.Select(eBlockStmt)
		if err != nil {
			return err
		}
		if !eb.IsPopulated() {
			// No more EBlocks.
			break
		}

		if sequence != eb.Sequence {
			return fmt.Errorf("invalid EBlock{%v, %v}: invalid Sequence",
				eb.Sequence, eb.KeyMR)
		}
		sequence++

		if (prevKeyMR != nil && *eb.PrevKeyMR != *prevKeyMR) ||
			(prevKeyMR == nil && !eb.PrevKeyMR.IsZero()) {
			return fmt.Errorf("invalid EBlock{%v, %v}: broken PrevKeyMR link",
				eb.Sequence, eb.KeyMR)
		}
		prevKeyMR = eb.KeyMR

		if (prevFullHash != nil && *eb.PrevFullHash != *prevFullHash) ||
			(prevFullHash == nil && !eb.PrevFullHash.IsZero()) {
			return fmt.Errorf("invalid EBlock{%v, %v}: broken FullHash link",
				eb.Sequence, eb.KeyMR)
		}
		prevFullHash = eb.FullHash

		for i, ebe := range eb.Entries {
			e, err := entries.Select(entryStmt)
			if err != nil {
				return err
			}

			if *e.Hash != *ebe.Hash {
				return fmt.Errorf("invalid Entry{%v}: broken EBlock link",
					e.Hash)
			}

			if *e.ChainID != *chain.ID {
				return fmt.Errorf("invalid Entry{%v}: invalid ChainID",
					e.Hash)
			}

			if e.Timestamp != ebe.Timestamp {
				return fmt.Errorf("invalid Entry{%v}: invalid Timestamp",
					e.Hash)
			}

			eb.Entries[i] = e
			eID++
		}
		dbKeyMR, err := eblocks.SelectDBKeyMR(read, eb.Sequence)
		if err != nil {
			return err
		}
		if err := chain.Apply(&dbKeyMR, eb); err != nil {
			return err
		}
	}
	if sequence == 0 {
		return fmt.Errorf("no eblocks")
	}

	changesetSQL, err := sqlitechangeset.SessionToSQL(chain.Conn, sess)
	if err != nil {
		chain.Log.Debugf("sqlitechangeset.SessionToSQL(): %v", err)
		return
	}
	if len(changesetSQL) > 0 {
		defer func() {
			chain.Log.Warnf("invalid state changeset: %v", changesetSQL)
			// Write the changeset to a file for later analysis...
			path := fmt.Sprintf("%v/%v-corrupt-%v.changeset",
				flag.DBPath, chain.ID.String(), time.Now().Unix())
			chain.Log.Warnf("writing corrupted state changeset to %v", path)
			f, err := os.Create(path)
			if err != nil {
				chain.Log.Debug(err)
				return
			}
			if _, err := f.WriteString(changesetSQL); err != nil {
				chain.Log.Debug(err)
			}
			if err := f.Close(); err != nil {
				chain.Log.Debug(err)
			}
		}()
		return fmt.Errorf("could not recompute saved state")
	}
	return nil
}
