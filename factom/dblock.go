package factom

import "fmt"

// DBlock unmarshals a Directory Block response from the factomd API for a
// given Height.
type DBlock struct {
	Height  uint64   `json:"-"`
	EBlocks []EBlock `json:"dbentries"`
}

// EBlock unmarshals an Entry Block response from the factomd API for a given
// KeyMR or ChainID.
type EBlock struct {
	// Populated when DBlock.Get is called
	ChainID *Bytes32 `json:"chainid"`
	KeyMR   *Bytes32 `json:"keymr,omitempty"`

	// Populated when EBlock.Get is called
	Entries      []Entry `json:"entrylist,omitempty"`
	EBlockHeader `json:"header,omitempty"`

	// Link back to the DBlock containing this EBlock
	DBlock `json:"-"`
}

// EBlockHeader unmarshals a nested structure of the Entry Block response from
// the factomd API.
type EBlockHeader struct {
	PrevKeyMR *Bytes32 `json:"prevkeymr"`
}

// Entry unmarshals an Entry in an Entry Block response from the factomd API
// and an Entry response for a given Hash.
type Entry struct {
	// Entry Block response fields
	Hash      Bytes32 `json:"entryhash,omitempty"`
	Timestamp Time    `json:"timestamp,omitempty"`

	// Entry response fields
	Content Bytes   `json:"content"`
	ExtIDs  []Bytes `json:"extids"`

	// Link back to the EBlock containing this entry
	EBlock
}

// Returns true if db has already been populated by a successful call to Get.
// Returns false if db or db.EBlocks is nil.
func (db *DBlock) Populated() bool {
	return db != nil && db.EBlocks != nil
}

// Get queries factomd for the directory block at db.Height. Returns any
// networking or marshalling errors. To check if the db has been successfully
// populated after calling Get(), call Populated().
func (db *DBlock) Get() error {
	if db.Populated() {
		return nil
	}
	params := map[string]interface{}{"height": db.Height}
	// We need the following anonymous struct to accomodate the way the
	// idiosyncratic way that the JSON response is returned.
	result := &struct {
		*DBlock `json:"dblock"`
	}{DBlock: db}
	if err := request("dblock-by-height", params, result); err != nil {
		return err
	}
	// Set DBlock pointer
	for i, _ := range db.EBlocks {
		db.EBlocks[i].DBlock = *db
	}
	return nil
}

// Returns true if eb has already been populated by a successful call to Get.
// Returns false if eb or eb.PrevKeyMR is nil.
func (eb *EBlock) Populated() bool {
	return eb != nil && eb.PrevKeyMR != nil
}

// Get queries factomd for the entry block corresponding to eb.KeyMR, if not
// nil, and otherwise the entry block chain head for eb.ChainID. Either
// eb.KeyMR or eb.ChainID must be allocated or else Get will panic. Returns any
// networking or marshalling errors. To check if the db has been successfully
// populated after calling Get(), call Populated().
func (eb *EBlock) Get() error {
	if eb.Populated() {
		return nil
	}
	var method string
	var params = make(map[string]interface{})
	// Fetch the specific entry-block if we have the KeyMR.
	// Otherwise fetch the chain head.
	if eb.KeyMR != nil {
		params["keymr"] = eb.KeyMR
		method = "entry-block"
	} else {
		params["chainid"] = eb.ChainID
		method = "chain-head"
	}
	if err := request(method, params, eb); err != nil {
		return err
	}
	// Populate the link in each entry back to its entry block
	for i, _ := range eb.Entries {
		eb.Entries[i].EBlock = *eb
	}
	return nil
}

// First returns true if this is the first EBlock in the chain. Returns false
// if eb has not yet populated by a successful call to Get.
func (eb *EBlock) First() bool {
	return eb.Populated() && *eb.PrevKeyMR == zeroBytes32
}

// NewPrev returns a pointer to a newly allocated EBlock with the ChainID set
// to eb.ChainID and KeyMR set to eb.PrevKeyMR. Get can then be called on this
// new EBlock to populate its data.
func (eb *EBlock) NewPrev() *EBlock {
	return &EBlock{ChainID: eb.ChainID, KeyMR: eb.PrevKeyMR}
}

// GetAllPrev returns a slice of EBlock pointers pointing to allocated and
// populated EBlocks of all preceding entry blocks up to and including eb. They
// are in order from earliest to latest. So eb is always equal to the last
// element of the returned slice. If eb is the first entry block, then it is
// the only element in the slice. If you are only interested in obtaining the
// first entry block, then consider using GetFirst to avoid unneeded memory
// allocations. Like Get, it returns any networking or marshalling errors. Call
// Populated() to check if the returned EBlock pointers have been successfully
// populated.
func (eb *EBlock) GetAllPrev() ([]*EBlock, error) {
	ebs := []*EBlock{}
	for eb := ebs[0]; !eb.First(); eb = ebs[0].NewPrev() {
		if err := eb.Get(); err != nil {
			return nil, fmt.Errorf("EBlock%+v.Get(): %v", eb, err)
		}
		if !eb.Populated() {
			return nil, nil
		}
		ebs = append([]*EBlock{eb}, ebs...)
	}
	return ebs, nil
}

// GetFirst finds and populates eb as the first EBlock in eb's chain. If eb is
// the first, then GetFirst does nothing. Otherwise, eb is reused to traverse
// up to the first entry block. GetFirst differs from GetAllPrev in that it
// does not allocate any additional EBlocks. Like Get, it returns any
// networking or marshalling errors. Call Populated() to check if eb has been
// successfully populated.
func (eb *EBlock) GetFirst() error {
	for ; !eb.First(); eb.KeyMR = eb.PrevKeyMR {
		if err := eb.Get(); err != nil {
			return fmt.Errorf("EBlock%+v.Get(): %v", eb, err)
		}
		if !eb.Populated() {
			return nil
		}
	}
	return nil
}

func (e *Entry) Populated() bool {
	return e != nil && e.ExtIDs != nil && e.Content != nil
}

// Get queries factomd for the entry corresponding to e.Hash, which must be
// populated.
func (e *Entry) Get() error {
	if e.Populated() {
		return nil
	}
	params := map[string]*Bytes32{"hash": &e.Hash}
	if err := request("entry", params, e); err != nil {
		return err
	}
	return nil
}

var zeroBytes32 Bytes32

// ZeroBytes32 returns an all zero Byte32.
func ZeroBytes32() Bytes32 {
	return zeroBytes32
}
