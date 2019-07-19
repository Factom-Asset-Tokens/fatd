package db

import (
	"fmt"

	"crawshaw.io/sqlite"
	"github.com/Factom-Asset-Tokens/fatd/factom"
)

// ValidateChain validates all Entry Hashes and EBlock KeyMRs, as well as the
// continuity of all stored EBlocks and Entries. It does not validate the
// validity of the saved DBlock KeyMRs.
func ValidateChain(conn *sqlite.Conn, chainID *factom.Bytes32) error {
	eBlockStmt := conn.Prep(SelectEBlockWhere + `true;`)
	entryStmt := conn.Prep(SelectEntryWhere + `true;`)
	var prevKeyMR, prevFullHash factom.Bytes32
	var sequence uint32
	var eID int = 1
	for {
		eb, err := SelectEBlock(eBlockStmt)
		if err != nil {
			return err
		}
		if !eb.IsPopulated() {
			// No more EBlocks.
			return nil
		}

		if *eb.ChainID != *chainID {
			return fmt.Errorf("invalid EBlock{%v, %v}: invalid ChainID",
				eb.Sequence, eb.KeyMR)
		}

		if sequence != eb.Sequence {
			return fmt.Errorf("invalid EBlock{%v, %v}: invalid Sequence",
				eb.Sequence, eb.KeyMR)
		}
		sequence++

		if *eb.PrevKeyMR != prevKeyMR {
			return fmt.Errorf("invalid EBlock{%v, %v}: broken PrevKeyMR link",
				eb.Sequence, eb.KeyMR)
		}

		if *eb.PrevFullHash != prevFullHash {
			return fmt.Errorf("invalid EBlock{%v, %v}: broken FullHash link",
				eb.Sequence, eb.KeyMR)
		}

		keyMR, err := eb.ComputeKeyMR()
		if err != nil {
			return err
		}
		if keyMR != *eb.KeyMR {
			return fmt.Errorf("invalid EBlock%+v: invalid KeyMR: %v",
				eb, keyMR)
		}

		prevFullHash, err = eb.ComputeFullHash()
		if err != nil {
			return err
		}
		prevKeyMR = keyMR

		for _, ebe := range eb.Entries {
			e, valid, err := SelectEntry(entryStmt)

			if *e.Hash != *ebe.Hash {
				return fmt.Errorf("invalid Entry{%v}: broken EBlock link",
					e.Hash)
			}

			hash, err := e.ComputeHash()
			if err != nil {
				return err
			}
			if hash != *e.Hash {
				return fmt.Errorf("invalid Entry{%v}: invalid Hash",
					e.Hash)
			}

			if *e.ChainID != *chainID {
				return fmt.Errorf("invalid Entry{%v}: invalid ChainID",
					e.Hash)
			}

			if e.Timestamp != ebe.Timestamp {
				return fmt.Errorf(
					"invalid Entry{%v, %v}: invalid Timestamp ebe %v e %v",
					eID, e.Hash, ebe.Timestamp, e.Timestamp)
			}

			// Attempt apply entry as fat entry.
			if valid != false {
				return fmt.Errorf("invalid Entry{%v}: marked as valid",
					e.Hash)
			}
			eID++
		}
	}
}
