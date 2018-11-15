package factom

// Entry represents a Factom Entry.
type Entry struct {
	// EBlock.Get populates the Hash, Timestamp, ChainID, and Height.
	Hash      *Bytes32 `json:"entryhash,omitempty"`
	Timestamp *Time    `json:"timestamp,omitempty"`
	ChainID   *Bytes32 `json:"chainid,omitempty"`
	Height    uint64   `json:"-"`

	// Entry.Get populates the Content and ExtIDs.
	Content Bytes   `json:"content"`
	ExtIDs  []Bytes `json:"extids"`
}

// IsPopulated returns true if e has already been successfully populated by a
// call to Get. IsPopulated returns false if both e.ExtIDs and e.Content are
// nil.
func (e Entry) IsPopulated() bool {
	return e.ExtIDs != nil || e.Content != nil
}

// Get queries factomd for the entry corresponding to e.Hash.
//
// Get returns any networking or marshaling errors, but not JSON RPC errors. To
// check if the Entry has been successfully populated, call IsPopulated().
func (e *Entry) Get() error {
	// If the Entry is already populated then there is nothing to do. If
	// the Hash is nil, we cannot populate it anyway.
	if e.IsPopulated() || e.Hash == nil {
		return nil
	}
	params := map[string]*Bytes32{"hash": e.Hash}
	if err := request("entry", params, e); err != nil {
		return err
	}
	return nil
}
