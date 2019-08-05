package db

import (
	"fmt"

	"crawshaw.io/sqlite"
	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat"
)

func (chain *Chain) insertMetadata() error {
	stmt := chain.Conn.Prep(`INSERT INTO metadata
                (id, sync_height, sync_db_key_mr, network_id, id_key_entry, id_key_height)
                VALUES (0, ?, ?, ?, ?, ?);`)
	stmt.BindInt64(1, int64(chain.SyncHeight))
	stmt.BindBytes(2, chain.SyncDBKeyMR[:])
	stmt.BindBytes(3, chain.NetworkID[:])
	if chain.Identity.IsPopulated() {
		data, err := chain.Identity.MarshalBinary()
		if err != nil {
			return err
		}
		fmt.Printf("bind bytes %x\n", data)
		stmt.BindBytes(4, data)
		stmt.BindInt64(5, int64(chain.Identity.Height))
	} else {
		stmt.BindNull(4)
		stmt.BindNull(5)
	}
	_, err := stmt.Step()
	return err
}

func (chain *Chain) SetSync(height uint32, dbKeyMR *factom.Bytes32) error {
	if height <= chain.SyncHeight {
		return nil
	}
	stmt := chain.Conn.Prep(`UPDATE metadata SET
                (sync_height, sync_db_key_mr) = (?, ?) WHERE id = 0;`)
	stmt.BindInt64(1, int64(height))
	stmt.BindBytes(2, dbKeyMR[:])
	_, err := stmt.Step()
	if chain.Conn.Changes() == 0 {
		panic("nothing updated")
	}
	chain.SyncHeight = height
	chain.SyncDBKeyMR = dbKeyMR
	return err
}

func (chain *Chain) setInitEntryID(id int64) error {
	stmt := chain.Conn.Prep(`UPDATE metadata SET
                (init_entry_id, num_issued) = (?, 0) WHERE id = 0;`)
	stmt.BindInt64(1, id)
	_, err := stmt.Step()
	if chain.Conn.Changes() == 0 {
		panic("nothing updated")
	}
	return err
}

func (chain *Chain) numIssuedAdd(add uint64) error {
	stmt := chain.Conn.Prep(`UPDATE metadata SET
                num_issued = num_issued + ? WHERE id = 0;`)
	stmt.BindInt64(1, int64(add))
	_, err := stmt.Step()
	if chain.Conn.Changes() == 0 {
		panic("nothing updated")
	}
	chain.NumIssued += add
	return err
}

func (chain *Chain) loadMetadata() error {
	// Load NameIDs
	first, err := SelectEntryByID(chain.Conn, 1)
	if err != nil {
		return err
	}
	if !first.IsPopulated() {
		return fmt.Errorf("no first entry")
	}

	nameIDs := first.ExtIDs
	if !fat.ValidTokenNameIDs(nameIDs) {
		return fmt.Errorf("invalid token chain Name IDs")
	}
	chain.TokenID, chain.IssuerChainID = fat.TokenIssuer(nameIDs)

	// Load Chain Head
	eb, dbKeyMR, err := SelectLatestEBlock(chain.Conn)
	if err != nil {
		return err
	}
	if !eb.IsPopulated() {
		// A database must always have at least one EBlock.
		return fmt.Errorf("no eblock in database")
	}
	chain.Head = eb
	chain.DBKeyMR = &dbKeyMR
	chain.ID = eb.ChainID

	stmt := chain.Conn.Prep(`SELECT sync_height, sync_db_key_mr, network_id,
                id_key_entry, id_key_height, init_entry_id, num_issued FROM metadata;`)
	hasRow, err := stmt.Step()
	if err != nil {
		return err
	}
	if !hasRow {
		return fmt.Errorf("no saved metadata")
	}

	chain.SyncHeight = uint32(stmt.ColumnInt64(0))

	chain.SyncDBKeyMR = new(factom.Bytes32)
	if stmt.ColumnBytes(1, chain.SyncDBKeyMR[:]) != len(chain.SyncDBKeyMR) {
		return fmt.Errorf("invalid sync_db_key_mr length")
	}

	if stmt.ColumnBytes(2, chain.NetworkID[:]) != len(chain.NetworkID) {
		return fmt.Errorf("invalid network_id length")
	}

	// Load chain.Identity...
	if stmt.ColumnType(3) == sqlite.SQLITE_NULL {
		// No Identity, therefore no Issuance.
		return nil
	}
	idKeyEntryData := make(factom.Bytes, stmt.ColumnLen(3))
	stmt.ColumnBytes(3, idKeyEntryData)
	if err := chain.Identity.UnmarshalBinary(idKeyEntryData); err != nil {
		return fmt.Errorf("chain.Identity.UnmarshalBinary(): %v", err)
	}
	chain.Identity.Height = uint32(stmt.ColumnInt64(4))
	if *chain.Identity.ChainID != *chain.IssuerChainID {
		return fmt.Errorf("invalid chain.Identity.ChainID")
	}
	chain.Identity.ChainID = chain.IssuerChainID // free mem from duplicates

	// Load chain.Issuance...
	if stmt.ColumnType(5) == sqlite.SQLITE_NULL {
		// No issuance entry so far...
		return nil
	}
	initEntryID := stmt.ColumnInt64(5)
	chain.Issuance.Entry.Entry, err = SelectEntryByID(chain.Conn, initEntryID)
	if err != nil {
		return err
	}
	if err := chain.Issuance.Validate(chain.ID1); err != nil {
		return err
	}
	chain.setApplyFunc()

	chain.NumIssued = uint64(stmt.ColumnInt64(6))

	return nil
}
