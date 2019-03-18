package state

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"

	"github.com/gocraft/dbr"
	"github.com/gocraft/dbr/dialect"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat"
	"github.com/Factom-Asset-Tokens/fatd/fat/fat1"
	"github.com/Factom-Asset-Tokens/fatd/flag"
	_log "github.com/Factom-Asset-Tokens/fatd/log"
)

var (
	SavedHeight uint64 = 163180
	log         _log.Log
	c           = flag.FactomClient
)

// Load state from all existing databases
func Load() error {
	log = _log.New("state")
	// Try to create the database directory in case it doesn't already
	// exist.
	if err := os.Mkdir(flag.DBPath, 0755); err != nil && !os.IsExist(err) {
		return fmt.Errorf("os.Mkdir(%#v)", flag.DBPath)
	}

	minHeight := uint64(math.MaxUint64)

	// Scan through all files within the database directory. Ignore invalid
	// file names.
	files, err := ioutil.ReadDir(flag.DBPath)
	if err != nil {
		return fmt.Errorf("ioutil.ReadDir(%#v): %v", flag.DBPath, err)
	}
	for _, f := range files {
		fname := f.Name()
		chain := Chain{ChainStatus: ChainStatusTracked}

		if chain.ID = fnameToChainID(fname); chain.ID == nil {
			continue
		}
		log.Debugf("loading chain: %v", chain.ID)
		var err error
		if err = chain.open(fname); err != nil {
			return err
		}
		if err := chain.loadMetadata(); err != nil {
			return err
		}
		if err := chain.loadIssuance(); err != nil {
			return err
		}

		Chains.set(chain.ID, &chain)
		if chain.Metadata.Height == 0 {
			continue
		}
		if chain.Metadata.Height < minHeight {
			minHeight = chain.Metadata.Height
		}
	}

	if minHeight < math.MaxUint64 {
		SavedHeight = minHeight
	}
	if flag.StartScanHeight > -1 {
		if uint64(flag.StartScanHeight-1) > SavedHeight {
			log.Warnf("-startscanheight (%v) is higher than the last saved block height (%v) which will very likely result in a corrupted database.",
				flag.StartScanHeight, SavedHeight)
		}
		SavedHeight = uint64(flag.StartScanHeight - 1)
	}
	return nil
}
func fnameToChainID(fname string) *factom.Bytes32 {
	if len(fname) != dbFileNameLen ||
		fname[dbFileNameLen-len(dbFileExtension):dbFileNameLen] != dbFileExtension {
		return nil
	}
	var chainID factom.Bytes32
	if err := json.Unmarshal(
		[]byte(fmt.Sprintf("%#v", fname[0:64])), &chainID); err != nil {
		return nil
	}
	return &chainID
}

func Close() {
	defer Chains.Unlock()
	Chains.Lock()
	for _, chain := range Chains.m {
		if chain.DB == nil {
			continue
		}
		if err := chain.Close(); err != nil {
			log.Errorf(err.Error())
		}
	}
}

func SaveHeight(height uint64) error {
	Chains.Lock()
	defer Chains.Unlock()

	for _, chain := range Chains.m {
		if !chain.IsTracked() ||
			chain.Metadata.Height >= height ||
			chain.DB.Error != nil {
			continue
		}
		if err := chain.saveHeight(height); err != nil {
			return err
		}
		Chains.m[*chain.ID] = chain
	}
	SavedHeight = height
	return nil
}

const (
	dbDriver        = "sqlite3"
	dbFileExtension = ".sqlite3"
	dbFileNameLen   = len(factom.Bytes32{})*2 + len(dbFileExtension)
)

// open a database
func (c *Chain) open(fname string) error {
	fpath := flag.DBPath + "/" + fname
	db, err := gorm.Open(dbDriver, fpath)
	if err != nil {
		return err
	}
	// Ensure the db gets closed if there are any issues.
	defer func() {
		if err != nil {
			db.Close()
		}
	}()
	db.LogMode(false)
	if err = autoMigrate(db); err != nil {
		return err
	}
	c.DB = db
	c.DBR = &dbr.Connection{
		DB: db.DB(), Dialect: dialect.SQLite3,
		EventReceiver: &dbr.NullEventReceiver{},
	}
	return nil
}
func autoMigrate(db *gorm.DB) error {
	if err := deleteEmptyTables(db); err != nil {
		return fmt.Errorf("deleteEmptyTables(): %v", err)
	}
	if err := db.AutoMigrate(&entry{}).Error; err != nil {
		return fmt.Errorf("db.AutoMigrate(&Entry{}): %v", err)
	}
	if err := db.AutoMigrate(&Address{}).Error; err != nil {
		return fmt.Errorf("db.AutoMigrate(&Address{}): %v", err)
	}
	if err := db.AutoMigrate(&Metadata{}).Error; err != nil {
		return fmt.Errorf("db.AutoMigrate(&Metadata{}): %v", err)
	}
	if err := db.AutoMigrate(&NFToken{}).Error; err != nil {
		return fmt.Errorf("db.AutoMigrate(&Metadata{}): %v", err)
	}
	return nil
}

func deleteEmptyTables(db *gorm.DB) error {
	var tables []struct{ Name string }
	var selectQry = "SELECT name FROM sqlite_master "
	qry := selectQry + "WHERE type = 'table';"
	if err := db.Raw(qry).Find(&tables).Error; err != nil {
		return fmt.Errorf("%#v: %v", qry, err)
	}
	for _, table := range tables {
		table := table.Name
		if table == "sqlite_sequence" {
			continue
		}
		var count int
		if err := db.Table(table).Count(&count).Error; err != nil {
			return fmt.Errorf("db.Table(%v).Count(): %v", table, err)
		}
		if count > 0 {
			continue
		}
		qry = fmt.Sprintf("DROP TABLE %v;", table)
		if err := db.Exec(qry).Error; err != nil {
			return fmt.Errorf("%#v: %v", qry, err)
		}
		var indexes []struct{ Name string }
		qry = selectQry + "WHERE type = 'index' AND tbl_name = ?;"
		if err := db.Raw(qry, table).
			Scan(&indexes).Error; err != nil {
			return fmt.Errorf("%#v: %v", qry, err)
		}
		for _, index := range indexes {
			index := index.Name
			qry = fmt.Sprintf("DROP INDEX %v;", index)
			if err := db.Exec(qry, index).Error; err != nil {
				return fmt.Errorf("%#v: %v", qry, err)
			}
		}
	}
	return nil
}

// setupDB a database for a given token chain.
func (chain *Chain) setupDB() error {
	fname := fmt.Sprintf("%v%v", chain.ID, dbFileExtension)
	var err error
	if err = chain.open(fname); err != nil {
		return err
	}
	// Ensure the db gets closed if there are any issues.
	defer func() {
		if err != nil {
			chain.Close()
			chain.DB = nil
		}
	}()
	if err := chain.Create(&chain.Metadata).Error; err != nil {
		return err
	}
	coinbase := newAddress(factom.Address{})
	if err := chain.Create(&coinbase).Error; err != nil {
		return err
	}
	return nil
}

func (chain *Chain) loadMetadata() error {
	var MetadataTableCount int
	if err := chain.DB.Model(&Metadata{}).
		Count(&MetadataTableCount).Error; err != nil {
		return err
	}
	if MetadataTableCount != 1 {
		return fmt.Errorf(`table "metadata" must have exactly one row`)
	}
	if err := chain.First(&chain.Metadata).Error; err != nil {
		return err
	}
	if !fat.ValidTokenNameIDs(fat.NameIDs(chain.Token, chain.Issuer)) ||
		*chain.ID != fat.ChainID(chain.Token, chain.Issuer) {
		return fmt.Errorf(`corrupted "metadata" table for chain %v`, chain.ID)
	}
	chain.Identity.ChainID = chain.Metadata.Issuer
	return nil
}

func (chain *Chain) loadIssuance() error {
	e := entry{}
	if err := chain.First(&e).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		return err
	}
	if !e.IsValid() {
		return fmt.Errorf("corrupted entry hash")
	}
	chain.Issuance = fat.NewIssuance(e.Entry())
	if err := chain.Issuance.UnmarshalEntry(); err != nil {
		return err
	}
	chain.ChainStatus = ChainStatusIssued
	if err := chain.Identity.Get(c); err != nil {
		return err
	}
	return nil
}

func (chain *Chain) saveIssuance() error {
	var entriesTableCount int
	if err := chain.DB.Model(&entry{}).Count(&entriesTableCount).Error; err != nil {
		return err
	}
	if entriesTableCount != 0 {
		return fmt.Errorf(`table "entries" must be empty prior to issuance`)
	}

	if _, err := chain.createEntry(chain.Issuance.Entry.Entry); err != nil {
		return err
	}
	chain.ChainStatus = ChainStatusIssued
	return nil
}
func (chain *Chain) saveMetadata() error {
	if err := chain.Save(&chain.Metadata).Error; err != nil {
		return err
	}
	return nil
}
func (chain *Chain) createEntry(fe factom.Entry) (*entry, error) {
	e := newEntry(fe)
	if !e.IsValid() {
		return nil, fmt.Errorf("invalid hash: factom.Entry%+v", fe)
	}
	if err := chain.Where("hash = ?", e.Hash).First(&e).Error; err !=
		gorm.ErrRecordNotFound {
		return nil, err
	}
	if err := chain.Create(&e).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

func (chain *Chain) createNFToken(tknID fat1.NFTokenID,
	metadata json.RawMessage) (*NFToken, error) {
	tkn := NFToken{NFTokenID: tknID, Metadata: metadata}
	if err := chain.Where("nf_token_id = ?", tknID).First(&tkn).Error; err !=
		gorm.ErrRecordNotFound {
		return nil, err
	}
	if err := chain.Create(&tkn).Error; err != nil {
		return nil, err
	}
	return &tkn, nil
}

func (chain *Chain) saveHeight(height uint64) error {
	chain.Metadata.Height = height
	if err := chain.saveMetadata(); err != nil {
		return err
	}
	return nil
}
func (chain Chain) GetAddress(rcdHash *factom.RCDHash) (Address, error) {
	a := Address{RCDHash: rcdHash}
	if err := chain.Where(&a).First(&a).Error; err != nil &&
		err != gorm.ErrRecordNotFound {
		return a, err
	}
	return a, nil
}

func (chain Chain) GetNFToken(tkn *NFToken) error {
	qry := chain.Where("nf_token_id = ?", tkn.NFTokenID)
	if tkn.OwnerID != 0 {
		qry = chain.Where("nf_token_id = ? AND owner_id = ?",
			tkn.NFTokenID, tkn.OwnerID)
	}
	if err := qry.Preload("Owner").First(tkn).Error; err != nil {
		return err
	}
	return nil
}

func (chain Chain) GetNFTokensForOwner(rcdHash *factom.RCDHash,
	page, limit uint, order string) (fat1.NFTokens, error) {
	sess := chain.DBR.NewSession(nil)
	ownerID := dbr.Select("id").From("addresses").
		Where("rcd_hash = ?", rcdHash)
	stmt := sess.Select("nf_token_id").From("nf_tokens").
		Where("owner_id = ?", ownerID)

	switch order {
	case "", "asc":
		stmt.OrderAsc("nf_token_id")
	case "desc":
		stmt.OrderDesc("nf_token_id")
	default:
		panic(fmt.Sprintf("invalid order value: %#v", order))
	}

	var dbtkns []NFToken
	if _, err := stmt.Load(&dbtkns); err != nil {
		return nil, err
	}
	tkns := make(fat1.NFTokens, len(dbtkns))
	for _, tkn := range dbtkns {
		tkns[tkn.NFTokenID] = struct{}{}
	}
	return tkns, nil
}

func (chain Chain) GetAllNFTokens(page, limit uint, order string) ([]NFToken, error) {
	var tkns []NFToken
	if err := chain.Offset(page * limit).Limit(limit).
		Order("nf_token_id " + order).
		Preload("Owner").Find(&tkns).Error; err != nil {
		return nil, err
	}
	return tkns, nil
}
func (chain *Chain) rollbackUnlessCommitted(savedChain Chain, err *error) {
	// This rollback will silently fail if the db tx has already
	// been committed.
	rberr := chain.Rollback().Error
	chain.DB = savedChain.DB
	if rberr == sql.ErrTxDone {
		// already committed
		return
	}
	if rberr != nil && *err != nil {
		// Report other Rollback errors if there wasn't already
		// a returned error.
		*err = rberr
		return
	}
	// complete rollback
	chain.Issued = savedChain.Issued
}

func (chain Chain) GetEntry(hash *factom.Bytes32) (factom.Entry, error) {
	e, err := chain.getEntry(hash)
	if e == nil {
		return factom.Entry{}, err
	}
	return e.Entry(), nil
}

func (chain Chain) getEntry(hash *factom.Bytes32) (*entry, error) {
	e := entry{}
	if err := chain.Not("id = ?", 1).
		Where("hash = ?", hash).First(&e).Error; err != nil {
		return nil, err
	}
	e.Hash = hash
	return &e, nil
}

const LimitMax = 1000

func (chain Chain) GetEntries(hash *factom.Bytes32,
	rcdHashes []factom.RCDHash, tknID *fat1.NFTokenID,
	toFrom, order string,
	page, limit uint) ([]factom.Entry, error) {
	if limit == 0 || limit > LimitMax {
		limit = LimitMax
	}

	sess := chain.DBR.NewSession(nil)
	stmt := sess.Select("*").From("entries").Where("id != 1").
		Limit(uint64(limit)).
		Offset(uint64(page * limit))

	var sign string
	switch order {
	case "", "asc":
		stmt.OrderAsc("id")
		sign = ">"
	case "desc":
		stmt.OrderDesc("id")
		sign = "<"
	default:
		panic(fmt.Sprintf("invalid order value: %#v", order))
	}

	if hash != nil {
		entryID := dbr.Select("id").From("entries").Where("hash = ?", hash)
		stmt.Where(fmt.Sprintf("id %v= ?", sign), entryID)
	}

	if len(rcdHashes) > 0 {
		addressIDs := dbr.Select("id").From("addresses").
			Where("rcd_hash IN ?", rcdHashes)
		var entryIDs dbr.Builder
		switch toFrom {
		case "to", "from":
			entryIDs = dbr.Select("entry_id").
				From("address_transactions_"+toFrom).
				Where("address_id IN ?", addressIDs)
		case "":
			entryIDs = dbr.UnionAll(
				dbr.Select("entry_id").From("address_transactions_to").
					Where("address_id IN ?", addressIDs),
				dbr.Select("entry_id").From("address_transactions_from").
					Where("address_id IN ?", addressIDs))
		default:
			panic(fmt.Sprintf("invalid toFrom value: %#v", toFrom))
		}
		stmt.Where("id IN ?", entryIDs)
	}

	if tknID != nil {
		tokenIDStmt := dbr.Select("id").From("nf_tokens").
			Where("nf_token_id == ?", tknID)
		entryIDs := dbr.Select("entry_id").
			From("nf_token_transactions").
			Where("nf_token_id == ?", tokenIDStmt)
		stmt.Where("id IN ?", entryIDs)
	}

	var es []entry
	if _, err := stmt.Load(&es); err != nil {
		return nil, err
	}
	entries := make([]factom.Entry, len(es))
	for i, e := range es {
		entries[i] = e.Entry()
	}
	return entries, nil
}

type erlog struct{}

func (e erlog) Event(eventName string) {
	log.Debugf("Event: %#v", eventName)
}
func (e erlog) EventKv(eventName string, kvs map[string]string) {
	log.Debugf("Event: %#v Kv: %v", eventName, kvs)
}
func (e erlog) EventErr(eventName string, err error) error {
	log.Debugf("Event: %#v Err: %v", eventName, err)
	return err
}
func (e erlog) EventErrKv(eventName string, err error, kvs map[string]string) error {
	log.Debugf("Event: %#v Err: %v Kvs: %v", eventName, err, kvs)
	return err
}
func (e erlog) Timing(eventName string, nanoseconds int64) {
	log.Debugf("Event: %#v Timing: %v", eventName, nanoseconds)
}
func (e erlog) TimingKv(eventName string, nanoseconds int64, kvs map[string]string) {
	log.Debugf("Event: %#v Timing: %v Kvs: %v", eventName, nanoseconds, kvs)
}
