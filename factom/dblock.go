package factom

// DBlock represents a Factom Directory Block.
type DBlock struct {
	Height uint64 `json:"-"`

	// DBlock.Get populates EBlocks with their ChainID and KeyMR.
	EBlocks []EBlock `json:"dbentries,omitempty"`
}

// IsPopulated returns true if db has already been successfully populated by a
// call to Get. IsPopulated returns false if db.EBlocks is nil.
func (db DBlock) IsPopulated() bool {
	return db.EBlocks != nil
}

// Get queries factomd for the Directory Block at db.Height.
//
// Get returns any networking or marshaling errors, but not JSON RPC errors. To
// check if the DBlock has been successfully populated, call IsPopulated().
func (db *DBlock) Get() error {
	if db.IsPopulated() {
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

	// Populate the Height for all EBlocks.
	for i := range db.EBlocks {
		db.EBlocks[i].Height = db.Height
	}
	return nil
}
