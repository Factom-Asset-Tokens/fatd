package state

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"crawshaw.io/sqlite"
	"github.com/Factom-Asset-Tokens/factom"
	"github.com/Factom-Asset-Tokens/fatd/internal/db"
)

type PendingChain struct {
	Chain

	ctx context.Context
	c   *factom.Client

	OfficialState    Chain
	OfficialSnapshot *sqlite.Snapshot

	Session *sqlite.Session

	Entries map[factom.Bytes32]factom.Entry
}

func ToPendingChain(chain Chain) (pending *PendingChain, ok bool) {
	pending, ok = chain.(*PendingChain)
	return
}

func NewPendingChain(ctx context.Context, c *factom.Client, chain Chain) (
	_ *PendingChain, err error) {

	factomChain := chain.ToFactomChain()
	factomChain.Log.Debug("Initializing pending...")

	if ok := factomChain.CloseMtx.RTryLock(ctx); !ok {
		return nil, ctx.Err()
	}
	s, err := factomChain.Pool.GetSnapshot(ctx, "")
	if err != nil {
		return
	}

	// Start a new session so we can track all changes and later rollback
	// all pending entries.
	session, err := factomChain.Conn.CreateSession("")
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			session.Delete()
		}
	}()
	if err := session.Attach(""); err != nil {
		return nil, err
	}

	if err := chain.UpdateSidechainData(ctx, c); err != nil {
		return nil, err
	}

	return &PendingChain{
		Chain:            chain,
		Entries:          make(map[factom.Bytes32]factom.Entry),
		OfficialState:    chain.Copy(),
		Session:          session,
		OfficialSnapshot: s,

		ctx: ctx, c: c,
	}, nil
}

func (pending *PendingChain) ApplyPendingEntries(es []factom.Entry) error {

	// Apply any new pending entries.
	for _, e := range es {
		// Ignore entries we have seen before.
		if _, ok := pending.Entries[*e.Hash]; ok {
			continue
		}

		// Load the Entry data.
		if err := e.Get(pending.ctx, pending.c); err != nil {
			return fmt.Errorf("factom.Entry.Get(): %w", err)
		}

		// The timestamp won't be established until the next EBlock so
		// use the current time for now.
		e.Timestamp = time.Now()

		if _, err := pending.Chain.ApplyEntry(e); err != nil {
			return fmt.Errorf("state.Chain.ApplyEntry(): %w", err)
		}

		// Cache the entry.
		pending.Entries[*e.Hash] = e
	}

	return nil
}
func (pending *PendingChain) LoadFromCache(eb *factom.EBlock) {
	// Load any cached entries that are pending and add them to eb.
	for i := range eb.Entries {
		e := &eb.Entries[i]

		// Check if this entry is cached.
		cachedE, ok := pending.Entries[*e.Hash]
		if !ok {
			continue
		}

		// Use official Timestamp established by EBlock.
		cachedE.Timestamp = e.Timestamp
		*e = cachedE
	}
}

func (pending *PendingChain) Revert() (Chain, error) {
	factomChain := pending.ToFactomChain()
	factomChain.Log.Debug("Cleaning up pending state...")
	// We must clear the interrupt to prevent from panicking or being
	// interrupted while reverting.
	oldDone := factomChain.Conn.SetInterrupt(nil)
	defer func() {
		pending.Session.Delete()
		factomChain.Conn.SetInterrupt(oldDone)
		factomChain.CloseMtx.RUnlock()
	}()
	// Revert all of the pending transactions by applying the inverse of
	// the changeset tracked by the session.
	var changeset bytes.Buffer
	if err := pending.Session.Changeset(&changeset); err != nil {
		return nil, fmt.Errorf("sqlite.Session.Changeset(): %w", err)
	}
	inverse := bytes.NewBuffer(make([]byte, 0, changeset.Len()))
	if err := sqlite.ChangesetInvert(inverse, &changeset); err != nil {
		return nil, fmt.Errorf("sqlite.ChangesetInvert(): %w", err)
	}
	if err := factomChain.Conn.ChangesetApply(
		inverse, nil, conflictFn(factomChain)); err != nil {
		return nil, fmt.Errorf("sqlite.Conn.ChangesetApply(): %w", err)
	}

	return pending.OfficialState, nil
}
func conflictFn(chain *db.FactomChain) func(
	sqlite.ConflictType, sqlite.ChangesetIter) sqlite.ConflictAction {

	return func(cType sqlite.ConflictType,
		_ sqlite.ChangesetIter) sqlite.ConflictAction {

		chain.Log.Errorf("ChangesetApply Conflict: %v", cType)
		return sqlite.SQLITE_CHANGESET_ABORT
	}
}

func (pending *PendingChain) Close() error {
	_, err := pending.Revert()
	if err != nil {
		pending.ToFactomChain().Log.Errorf(
			"state.PendingChain.Revert(): %v", err)
	}
	return pending.Chain.Close()
}
