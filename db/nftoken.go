package db

import (
	"encoding/json"
	"fmt"

	"crawshaw.io/sqlite"
	"crawshaw.io/sqlite/sqlitex"
	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat/fat1"
)

func (chain *Chain) setNFTokenOwner(nfID fat1.NFTokenID, aID, eID int64) error {
	stmt := chain.Conn.Prep(`INSERT INTO "nf_tokens"
                ("id", "owner_id", "creation_entry_id") VALUES (?, ?, ?)
                ON CONFLICT("id") DO
                UPDATE SET "owner_id" = "excluded"."owner_id";`)
	stmt.BindInt64(1, int64(nfID))
	stmt.BindInt64(2, aID)
	stmt.BindInt64(3, eID)
	_, err := stmt.Step()
	return err
}

func (chain *Chain) setNFTokenMetadata(
	nfID fat1.NFTokenID, metadata json.RawMessage) error {
	stmt := chain.Conn.Prep(`UPDATE "nf_tokens"
                SET "metadata" = ? WHERE "id" = ?;`)
	stmt.BindBytes(1, metadata)
	stmt.BindInt64(2, int64(nfID))
	_, err := stmt.Step()
	if chain.Conn.Changes() == 0 {
		panic("no NFTokenID updated")
	}
	return err
}

func (chain *Chain) insertNFTokenTransaction(nfID fat1.NFTokenID, adrTxID int64) error {
	stmt := chain.Conn.Prep(`INSERT INTO "nf_token_transactions"
                ("nf_tkn_id", "adr_tx_id") VALUES (?, ?);`)
	stmt.BindInt64(1, int64(nfID))
	stmt.BindInt64(2, adrTxID)
	_, err := stmt.Step()
	return err
}

func SelectNFTokenOwnerID(conn *sqlite.Conn, nfTkn fat1.NFTokenID) (int64, error) {
	stmt := conn.Prep(`SELECT "owner_id" FROM "nf_tokens" WHERE "id" = ?;`)
	stmt.BindInt64(1, int64(nfTkn))
	ownerID, err := sqlitex.ResultInt64(stmt)
	if err != nil && err.Error() == "sqlite: statement has no results" {
		return -1, nil
	}
	if err != nil {
		return -1, err
	}
	return ownerID, nil
}

func SelectNFToken(conn *sqlite.Conn,
	nfTkn fat1.NFTokenID) (factom.FAAddress, factom.Bytes32, []byte, error) {
	var owner factom.FAAddress
	var creationHash factom.Bytes32
	stmt := conn.Prep(`SELECT "owner", "metadata", "creation_hash"
                        FROM "nf_tokens_addresses" WHERE "id" = ?;`)
	stmt.BindInt64(1, int64(nfTkn))
	hasRow, err := stmt.Step()
	if err != nil {
		stmt.Reset()
		return owner, creationHash, nil, err
	}
	if !hasRow {
		return owner, creationHash, nil, nil
	}
	defer stmt.Reset()
	if stmt.ColumnBytes(0, owner[:]) != len(owner) {
		panic("invalid address length")
	}
	metadata := make([]byte, stmt.ColumnLen(1))
	stmt.ColumnBytes(1, metadata)

	if stmt.ColumnBytes(2, creationHash[:]) != len(creationHash) {
		panic("invalid hash length")
	}
	return owner, creationHash, metadata, nil
}

func SelectNFTokens(conn *sqlite.Conn,
	order string, page, limit uint64) ([]fat1.NFTokenID,
	[]factom.FAAddress, []factom.Bytes32, [][]byte, error) {
	if page == 0 {
		return nil, nil, nil, nil, fmt.Errorf("invalid page")
	}
	stmt := conn.Prep(`SELECT "id", "owner", "creation_hash", "metadata"
                        FROM "nf_tokens_addresses";`)
	defer stmt.Reset()

	var tkns []fat1.NFTokenID
	var owners []factom.FAAddress
	var creationHashes []factom.Bytes32
	var metadata [][]byte
	for {
		hasRow, err := stmt.Step()
		if err != nil {
			return nil, nil, nil, nil, err
		}
		if !hasRow {
			break
		}
		tkns = append(tkns, fat1.NFTokenID(stmt.ColumnInt64(0)))

		var owner factom.FAAddress
		if stmt.ColumnBytes(1, owner[:]) != len(owner) {
			stmt.Reset()
			panic("invalid address length")
		}
		owners = append(owners, owner)

		var creationHash factom.Bytes32
		if stmt.ColumnBytes(2, creationHash[:]) != len(creationHash) {
			stmt.Reset()
			panic("invalid hash length")
		}
		creationHashes = append(creationHashes, creationHash)

		data := make([]byte, stmt.ColumnLen(3))
		stmt.ColumnBytes(3, data)
		metadata = append(metadata, data)
	}
	return tkns, owners, creationHashes, metadata, nil
}

func SelectNFTokensByOwner(conn *sqlite.Conn, adr *factom.FAAddress,
	page, limit uint64, order string) ([]fat1.NFTokenID, error) {
	if page == 0 {
		return nil, fmt.Errorf("invalid page")
	}
	var sql sql
	sql.Append(`SELECT "id" FROM "nf_tokens" WHERE "owner_id" = (
                SELECT "id" FROM "addresses" WHERE "address" = ?)`,
		func(s *sqlite.Stmt, c int) int {
			s.BindBytes(c, adr[:])
			return 1
		})
	sql.OrderPaginate(order, page, limit)

	stmt := sql.Prep(conn)
	defer stmt.Reset()
	var nfTkns []fat1.NFTokenID
	for {
		hasRow, err := stmt.Step()
		if err != nil {
			return nil, err
		}
		if !hasRow {
			break
		}
		nfTkns = append(nfTkns, fat1.NFTokenID(stmt.ColumnInt64(0)))
	}
	return nfTkns, nil
}
