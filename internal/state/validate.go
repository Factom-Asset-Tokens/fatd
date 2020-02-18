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

package state

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"time"

	"crawshaw.io/sqlite"
	"crawshaw.io/sqlite/sqlitex"
	"github.com/AdamSLevy/sqlitechangeset"
	"github.com/Factom-Asset-Tokens/factom"
	"github.com/Factom-Asset-Tokens/factom/fat"
	"github.com/Factom-Asset-Tokens/fatd/internal/db/eblock"
	"github.com/Factom-Asset-Tokens/fatd/internal/db/entry"
	"github.com/Factom-Asset-Tokens/fatd/internal/flag"
)

func init() {
	sqlitechangeset.AlwaysUseBlob = true
}

// Validate all Entry Hashes and EBlock KeyMRs, as well as the continuity of
// all stored EBlocks and Entries.
//
// This does not validate the validity of the saved DBlock KeyMRs.
func (chain *FATChain) Validate(ctx context.Context, repair bool) (err error) {
	// Validate ChainID...
	chain.Log.Info("Validating...")
	read := chain.Pool.Get(ctx)
	defer chain.Pool.Put(read)
	write := chain.Conn
	write.SetInterrupt(ctx.Done())
	first, err := entry.SelectByID(read, 1)
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
	sess.Attach("eblock")
	sess.Attach("entry")
	sess.Attach("address")
	sess.Attach("nftoken")
	sess.Attach("metadata")
	defer sess.Delete()

	// In case there are any changes, we want to roll back everything. We
	// don't fix corrupted databases, at least not yet.
	defer chain.Save()(&err)

	chain.Head = factom.EBlock{}
	chain.SyncHeight = 0
	chain.SyncDBKeyMR = nil

	// Completely clear the state, while preserving all chain data.
	err = sqlitex.ExecScript(write, `
                UPDATE "address" SET "balance" = 0;
                DELETE FROM "address_tx";
                DELETE FROM "nftoken";
                DELETE FROM "nftoken_tx";
                DELETE FROM "eblock";
                DELETE FROM "entry";
                UPDATE "fat_chain" SET ("init_entry_id", "num_issued") = (NULL, NULL);
                `)
	if err != nil {
		return err
	}
	chain.NumIssued = 0
	chain.Issuance = fat.Issuance{}

	eBlockStmt, _, err := read.PrepareTransient(
		eblock.SelectWhere + `true;`) // SELECT all EBlocks.
	if err != nil {
		panic(err)
	}
	defer eBlockStmt.Finalize()
	entryStmt, _, err := read.PrepareTransient(
		entry.SelectWhere + `true;`) // SELECT all Entries.
	if err != nil {
		panic(err)
	}
	defer entryStmt.Finalize()

	var eID int = 1     // Entry ID
	var sequence uint32 // EBlock Sequence
	var prevKeyMR, prevFullHash *factom.Bytes32
	for {
		eb, err := eblock.Select(eBlockStmt)
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
			e, err := entry.Select(entryStmt)
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
				return fmt.Errorf(
					"invalid Entry{%v}: invalid Timestamp: %v, expected: %v",
					e.Hash, e.Timestamp, ebe.Timestamp)
			}

			eb.Entries[i] = e
			eID++
		}
		dbKeyMR, err := eblock.SelectDBKeyMR(read, eb.Sequence)
		if err != nil {
			return err
		}
		if err := Apply(chain, &dbKeyMR, eb); err != nil {
			return err
		}
	}
	if sequence == 0 {
		return fmt.Errorf("no eblocks")
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}

	var changeset bytes.Buffer
	if err := sess.Changeset(&changeset); err != nil {
		return fmt.Errorf("sqlite.Session.Changeset(): %w", err)
	}

	iter, err := sqlite.ChangesetIterStart(&changeset)
	if err != nil {
		return fmt.Errorf("sqlite.ChangesetIterStart(): %w", err)
	}
	defer func() {
		if err := iter.Finalize(); err != nil {
			chain.Log.Errorf("sqlite.ChangesetIter.Finalize(): %w", err)
		}
	}()
	hasRow, err := iter.Next()
	if err != nil {
		return fmt.Errorf("sqlite.ChangesetIter.Next(): %w", err)
	}
	if hasRow {
		if repair {
			chain.Log.Warnf("Corrupted state repaired!")
			return nil
		}
		chain.Log.Error("Corrupted state!")
		// Write the changeset to a file for later analysis...
		path := fmt.Sprintf("%v%v-corrupt-%v.changeset",
			flag.DBPath, chain.ID.String(), time.Now().Unix())
		chain.Log.Warnf("writing corrupted state changeset to %v", path)
		f, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("os.Create(): %w", err)
		}
		if err := sess.Changeset(f); err != nil {
			return fmt.Errorf("sqlite.Session.Changeset(): %w", err)
		}
		if err := f.Close(); err != nil {
			return fmt.Errorf("os.File.Close(): %w", err)
		}
		return fmt.Errorf("Corrupted state!")
	}
	return nil
}
