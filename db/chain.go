package db

import (
	"fmt"
	"os"
	"strings"

	"crawshaw.io/sqlite"
	"crawshaw.io/sqlite/sqlitex"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat"
	"github.com/Factom-Asset-Tokens/fatd/fat/fat0"
	"github.com/Factom-Asset-Tokens/fatd/fat/fat1"
	"github.com/Factom-Asset-Tokens/fatd/flag"
	_log "github.com/Factom-Asset-Tokens/fatd/log"
)

// Chain combines a READWRITE sqlite.Conn and a READ_ONLY sqlitex.Pool with the
// corresponding ChainID. Only a single thread may use the Read/Write Conn at a
// time. Multi-threaded readers must Pool.Get a read-only Conn from the and
// must Pool.Put the Conn back when finished.
type Chain struct {
	ID            *factom.Bytes32
	TokenID       string
	IssuerChainID *factom.Bytes32
	Head          factom.EBlock
	DBKeyMR       *factom.Bytes32
	factom.Identity
	NetworkID factom.NetworkID

	SyncHeight  uint32
	SyncDBKeyMR *factom.Bytes32

	fat.Issuance
	NumIssued uint64

	*sqlite.Conn  // Read/Write
	*sqlitex.Pool // Read Only Pool
	Log           _log.Log

	apply ApplyFunc
}

type ApplyFunc func(factom.Entry) error

// Open a new Chain for the given chainID within flag.DBPath. Validate or apply
// chainDBSchema.
func OpenNew(eb factom.EBlock, dbKeyMR *factom.Bytes32, networkID factom.NetworkID,
	identity factom.Identity) (*Chain, error) {
	fname := eb.ChainID.String() + dbFileExtension
	path := flag.DBPath + "/" + fname
	// Ensure that the database file doesn't already exist.
	_, err := os.Stat(path)
	if err == nil {
		return nil, fmt.Errorf("already exists: %v", path)
	}
	if !os.IsNotExist(err) { // Any other error is unexpected.
		return nil, err
	}

	chain, err := open(fname)
	if err != nil {
		return nil, err
	}

	chain.Head = eb
	chain.ID = eb.ChainID
	nameIDs := eb.Entries[0].ExtIDs
	if !fat.ValidTokenNameIDs(nameIDs) {
		return nil, fmt.Errorf("invalid token chain Name IDs")
	}
	chain.TokenID, chain.IssuerChainID = fat.TokenIssuer(nameIDs)
	chain.DBKeyMR = dbKeyMR
	chain.Identity = identity
	chain.SyncHeight = eb.Height
	chain.SyncDBKeyMR = dbKeyMR
	chain.NetworkID = networkID

	if err := chain.InsertMetadata(); err != nil {
		return nil, err
	}
	coinbase := fat.Coinbase()
	if _, err := AddressAdd(chain.Conn, &coinbase, 0); err != nil {
		return nil, err
	}

	chain.apply = chain.ApplyIssuance
	if err := chain.Apply(eb, dbKeyMR); err != nil {
		return nil, err
	}

	return chain, nil
}

// Open an existing chain database.
func Open(fname string) (*Chain, error) {
	chain, err := open(fname)
	if err != nil {
		return nil, err
	}

	// Load NameIDs, so load the first entry.
	first, err := SelectEntryByID(chain.Conn, 1)
	if err != nil {
		return nil, err
	}
	if !first.IsPopulated() {
		// A database must always have at least one EBlock.
		return nil, fmt.Errorf("no first entry")
	}

	nameIDs := first.ExtIDs
	if !fat.ValidTokenNameIDs(nameIDs) {
		return nil, fmt.Errorf("invalid token chain Name IDs")
	}
	chain.TokenID, chain.IssuerChainID = fat.TokenIssuer(nameIDs)

	// Load Chain Head
	eb, dbKeyMR, err := SelectLatestEBlock(chain.Conn)
	if err != nil {
		return nil, err
	}
	if !eb.IsPopulated() {
		// A database must always have at least one EBlock.
		return nil, fmt.Errorf("no eblock in database")
	}
	chain.Head = eb
	chain.DBKeyMR = &dbKeyMR
	chain.ID = eb.ChainID

	if err := chain.LoadMetadata(); err != nil {
		return nil, err
	}

	return chain, nil
}

// open a READWRITE Conn and a READ_ONLY Pool for flag.DBPath + "/" + fname.
// Validate or apply schema.
func open(fname string) (*Chain, error) {
	const baseFlags = sqlite.SQLITE_OPEN_WAL |
		sqlite.SQLITE_OPEN_URI |
		sqlite.SQLITE_OPEN_NOMUTEX
	path := flag.DBPath + "/" + fname
	flags := baseFlags | sqlite.SQLITE_OPEN_READWRITE | sqlite.SQLITE_OPEN_CREATE
	conn, err := sqlite.OpenConn(path, flags)
	if err != nil {
		return nil, fmt.Errorf("sqlite.OpenConn(%q, %x): %v",
			path, flags, err)
	}
	if err := validateOrApplySchema(conn, chainDBSchema); err != nil {
		return nil, err
	}
	flags = baseFlags | sqlite.SQLITE_OPEN_READONLY
	pool, err := sqlitex.Open(path, flags, PoolSize)
	if err != nil {
		return nil, fmt.Errorf("sqlitex.Open(%q, %x, %v): %v",
			path, flags, PoolSize, err)
	}
	return &Chain{Conn: conn, Pool: pool,
		Log: _log.New("chain", strings.TrimRight(fname, dbFileExtension)),
	}, nil
}

// Close the Conn and Pool. Log any errors.
func (chain *Chain) Close() {
	if err := chain.Conn.Close(); err != nil {
		chain.Log.Errorf("chain.Conn.Close(): %v", err)
	}
	if err := chain.Pool.Close(); err != nil {
		chain.Log.Errorf("chain.Pool.Close(): %v", err)
	}
}

// Apply save the EBlock and all Entries and updates the chain state according
// to the FAT-0 and FAT-1 protocols.
func (chain *Chain) Apply(eb factom.EBlock, dbKeyMR *factom.Bytes32) (err error) {
	defer sqlitex.Save(chain.Conn)(&err)
	if err := InsertEBlock(chain.Conn, eb, dbKeyMR); err != nil {
		return err
	}
	chain.Head = eb
	for _, e := range eb.Entries {
		if err := chain.apply(e); err != nil {
			return err
		}
	}
	return nil
}

func (chain *Chain) ApplyIssuance(e factom.Entry) error {
	eid, err := InsertEntry(chain.Conn, e, chain.Head.Sequence)
	if err != nil {
		return err
	}
	issuance := fat.NewIssuance(e)
	if err = issuance.Validate(chain.ID1); err != nil {
		chain.Log.Debugf("Entry{%v}: invalid issuance: %v", e.Hash, err)
		return nil
	}
	// check sig and is valid
	if err := SaveInitEntryID(chain.Conn, eid); err != nil {
		return err
	}
	chain.Issuance = issuance
	chain.SetApplyFunc()
	chain.Log.Debugf("Valid Issuance Entry: %v %+v", e.Hash, issuance)
	return nil
}

func (chain *Chain) ApplyFAT0Tx(e factom.Entry) (tx fat0.Transaction, err error) {
	tx = fat0.NewTransaction(e)
	valid, eID, err := chain.ApplyTx(e, &tx)
	if err != nil {
		return
	}
	if !valid {
		return
	}

	// Do not return, but log, any errors past this point as they are
	// related to being unable to apply a transaction.
	defer func() {
		if err != nil {
			chain.Log.Debugf("Entry{%v}: invalid %v transaction: %v",
				e.Hash, chain.Type, err)
			err = nil
		} else {
			var cbStr string
			if tx.IsCoinbase() {
				cbStr = "Coinbase "
			}
			chain.Log.Debugf("Valid %v %vTransaction: %v %+v",
				chain.Type, cbStr, tx.Hash, tx)
		}
	}()
	// But first rollback on any error.
	defer sqlitex.Save(chain.Conn)(&err)

	if err = MarkEntryValid(chain.Conn, eID); err != nil {
		return
	}

	for adr, amount := range tx.Outputs {
		var aID int64
		aID, err = AddressAdd(chain.Conn, &adr, amount)
		if err != nil {
			return
		}
		if err = InsertAddressTransaction(chain.Conn,
			aID, eID, true); err != nil {
			return
		}
	}

	if tx.IsCoinbase() {
		addIssued := tx.Inputs[fat.Coinbase()]
		if chain.Supply > 0 &&
			int64(chain.NumIssued+addIssued) > chain.Supply {
			err = fmt.Errorf("coinbase exceeds max supply")
			return
		}
		if err = InsertAddressTransaction(chain.Conn,
			1, eID, false); err != nil {
			return
		}
		err = chain.IncrementNumIssued(addIssued)
		return
	}

	for adr, amount := range tx.Inputs {
		var aID int64
		aID, err = AddressSub(chain.Conn, &adr, amount)
		if err != nil {
			return
		}
		if err = InsertAddressTransaction(chain.Conn,
			aID, eID, false); err != nil {
			return
		}
	}
	return
}

func (chain *Chain) ApplyTx(e factom.Entry, tx fat.Validator) (bool, int64, error) {
	eID, err := InsertEntry(chain.Conn, e, chain.Head.Sequence)
	if err != nil {
		return false, eID, err
	}
	if err := tx.Validate(chain.ID1); err != nil {
		chain.Log.Debugf("Entry{%v}: invalid %v transaction: %v",
			e.Hash, chain.Type, err)
		return false, eID, nil
	}
	valid, err := CheckEntryUniqueValid(chain.Conn, eID, e.Hash)
	if err != nil {
		return false, eID, err
	}
	if !valid {
		chain.Log.Debugf("Entry{%v}: invalid %v transaction: %v",
			e.Hash, chain.Type, "replay")
	}
	return valid, eID, nil
}

func (chain *Chain) SetApplyFunc() {
	// Adapt to match ApplyFunc.
	switch chain.Type {
	case fat0.Type:
		chain.apply = func(e factom.Entry) error {
			_, err := chain.ApplyFAT0Tx(e)
			return err
		}
	case fat1.Type:
		chain.apply = func(e factom.Entry) error {
			_, err := chain.ApplyFAT1Tx(e)
			return err
		}
	default:
		panic("invalid type")
	}
}

func (chain *Chain) ApplyFAT1Tx(e factom.Entry) (tx fat1.Transaction, err error) {
	tx = fat1.NewTransaction(e)
	valid, eID, err := chain.ApplyTx(e, &tx)
	if err != nil {
		return
	}
	if !valid {
		return
	}

	// Do not return, but log, any errors past this point as they are
	// related to being unable to apply a transaction.
	defer func() {
		if err != nil {
			chain.Log.Debugf("Entry{%v}: invalid %v transaction: %v",
				e.Hash, chain.Type, err)
			err = nil
		} else {
			var cbStr string
			if tx.IsCoinbase() {
				cbStr = "Coinbase "
			}
			chain.Log.Debugf("Valid %v %vTransaction: %v %+v",
				chain.Type, cbStr, tx.Hash, tx)
		}
	}()
	// But first rollback on any error.
	defer sqlitex.Save(chain.Conn)(&err)

	if err = MarkEntryValid(chain.Conn, eID); err != nil {
		return
	}

	for adr, nfTkns := range tx.Outputs {
		var aID int64
		aID, err = AddressAdd(chain.Conn, &adr, uint64(len(nfTkns)))
		if err != nil {
			return
		}
		if err = InsertAddressTransaction(chain.Conn,
			aID, eID, true); err != nil {
			return
		}
		for nfID := range nfTkns {
			if err = SetNFTokenOwner(chain.Conn, nfID, aID, eID); err != nil {
				return
			}
			if err = InsertNFTokenTransaction(chain.Conn,
				nfID, eID, aID); err != nil {
				return
			}
		}
	}

	if tx.IsCoinbase() {
		nfTkns := tx.Inputs[fat.Coinbase()]
		addIssued := uint64(len(nfTkns))
		if chain.Supply > 0 &&
			int64(chain.NumIssued+addIssued) > chain.Supply {
			err = fmt.Errorf("coinbase exceeds max supply")
			return
		}
		if err = InsertAddressTransaction(chain.Conn,
			1, eID, false); err != nil {
			return
		}
		for nfID := range nfTkns {
			metadata := tx.TokenMetadata[nfID]
			if len(metadata) == 0 {
				continue
			}
			if err = AttachNFTokenMetadata(chain.Conn,
				nfID, metadata); err != nil {
				return
			}
		}
		err = chain.IncrementNumIssued(addIssued)
		return
	}

	for adr, nfTkns := range tx.Inputs {
		var aID int64
		aID, err = AddressSub(chain.Conn, &adr, uint64(len(nfTkns)))
		if err != nil {
			return
		}
		if err = InsertAddressTransaction(chain.Conn,
			aID, eID, false); err != nil {
			return
		}
	}
	return
}
