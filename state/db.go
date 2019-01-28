package state

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"

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
		if chain.DB, err = open(fname); err != nil {
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
		if !chain.IsTracked() || chain.Metadata.Height >= height || chain.DB.Error != nil {
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
func open(fname string) (*gorm.DB, error) {
	fpath := flag.DBPath + "/" + fname
	db, err := gorm.Open(dbDriver, fpath)
	if err != nil {
		return nil, err
	}
	// Ensure the db gets closed if there are any issues.
	defer func() {
		if err != nil {
			db.Close()
		}
	}()
	db.LogMode(false)
	if err = autoMigrate(db); err != nil {
		return nil, err
	}
	return db, nil
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
	if chain.DB, err = open(fname); err != nil {
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
	if err := chain.Identity.Get(); err != nil {
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
	page, limit uint) ([]fat1.NFTokenID, error) {
	a, err := chain.GetAddress(rcdHash)
	if err != nil {
		return nil, err
	}
	var tkns []NFToken
	if err := chain.Where(&NFToken{OwnerID: a.ID}).
		Offset(page * limit).Limit(limit).
		Find(&tkns).Error; err != nil {
		return nil, err
	}
	tknIDs := make([]fat1.NFTokenID, len(tkns))
	for i, tkn := range tkns {
		tknIDs[i] = tkn.NFTokenID
	}
	return tknIDs, nil
}

func (chain Chain) GetAllNFTokens(page, limit uint) ([]NFToken, error) {
	var tkns []NFToken
	if err := chain.Offset(page * limit).Limit(limit).
		Order("nf_token_id").
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

func (chain Chain) GetEntries(hash *factom.Bytes32,
	rcdHash *factom.RCDHash, toFrom string,
	page, limit uint) ([]factom.Entry, error) {
	if limit == 0 {
		limit = math.MaxUint32
	}
	var e *entry
	var es []entry
	if rcdHash != nil {
		a, err := chain.GetAddress(rcdHash)
		if err != nil {
			return nil, err
		}
		var to, from []entry
		if toFrom != "from" {
			if err := chain.Limit(limit).Model(&a).
				Association("To").Find(&to).Error; err != nil {
				return nil, err
			}
		}
		if toFrom != "to" {
			if err := chain.Limit(limit).Model(&a).
				Association("From").Find(&from).Error; err != nil {
				return nil, err
			}
		}
		if int(limit) > len(to)+len(from) {
			limit = uint(len(to) + len(from))
		}
		es = make([]entry, limit)
		var t, f int
		for i := range es {
			if t < len(to) && f < len(from) {
				var next entry
				if to[t].ID < from[f].ID {
					next = to[t]
					t++
				} else {
					next = from[f]
					f++
				}
				es[i] = next
			} else {
				var nexts []entry
				if t < len(to) {
					nexts = to
				} else {
					nexts = from
				}
				es = append(es[:i], nexts...)
				break
			}
		}
		if hash != nil {
			hashId := uint(len(es))
			for i, e := range es {
				if *e.Hash == *hash {
					hashId = uint(i)
					break
				}
			}
			page += hashId
			if page > uint(len(es)) {
				page = uint(len(es))
			}
		}
		es = es[page:]
	} else {
		if hash != nil {
			var err error
			e, err = chain.getEntry(hash)
			if e == nil {
				return nil, err
			}
			page = uint(e.ID)
		} else {
			page++
		}
		if err := chain.Offset(page).Limit(limit).Find(&es).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, nil
			}
			return nil, err
		}
	}
	entries := make([]factom.Entry, len(es))
	for i, e := range es {
		entries[i] = e.Entry()
	}
	return entries, nil
}
