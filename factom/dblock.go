package factom

// DBlock represents a Factom Directory Block.
type DBlock struct {
	Height uint64 `json:"height"`

	// DBlock.Get populates EBlocks with their ChainID and KeyMR.
	EBlocks []EBlock `json:"dbentries,omitempty"`
}

// IsPopulated returns true if db has already been successfully populated by a
// call to Get. IsPopulated returns false if db.EBlocks is nil.
func (db DBlock) IsPopulated() bool {
	return db.EBlocks != nil
}

// Get queries factomd for the Directory Block at db.Height. After a successful
// call, the EBlocks will all have their ChainID and KeyMR, but not their
// Entries. Call Get on the EBlocks individually to populate their Entries.
func (db *DBlock) Get(c *Client) error {
	if db.IsPopulated() {
		return nil
	}

	result := struct{ DBlock *DBlock }{DBlock: db}
	if err := c.FactomdRequest("dblock-by-height", db, &result); err != nil {
		return err
	}

	return nil
}
