package factom

import "fmt"

// EBlock represents a Factom Entry Block.
type EBlock struct {
	// DBlock.Get populates the ChainID, KeyMR, and Height.
	ChainID *Bytes32 `json:"chainid,omitempty"`
	KeyMR   *Bytes32 `json:"keymr,omitempty"`

	// EBlock.Get populates the EBlockHeader.PrevKeyMR and the Entries with
	// their Hash and Timestamp.
	EBlockHeader `json:"header,omitempty"`
	Entries      []Entry `json:"entrylist,omitempty"`
}

// EBlockHeader is required for unmashaling the nested structure of the Entry
// Block response from the factomd JSON RPC API.
type EBlockHeader struct {
	PrevKeyMR *Bytes32 `json:"prevkeymr,omitempty"`
	Height    uint64   `json:"dbheight"`
}

// IsPopulated returns true if eb has already been successfully populated by a
// call to Get. Returns false if eb.PrevKeyMR is nil.
func (eb EBlock) IsPopulated() bool {
	return eb.Entries != nil &&
		eb.ChainID != nil &&
		eb.KeyMR != nil &&
		eb.PrevKeyMR != nil
}

// Get queries factomd for the Entry Block corresponding to eb.KeyMR, if not
// nil, and otherwise the Entry Block chain head for eb.ChainID. Either
// eb.KeyMR or eb.ChainID must be not nil or else Get will fail to populate the
// EBlock. After a successful call, EBlockHeader and Entries will be populated.
// Each Entry will be populated with its Hash, Timestamp, ChainID, and Height,
// but not its Content or ExtIDs. Call Get on the individual Entries to
// populate their Content and ExtIDs.
func (eb *EBlock) Get(c *Client) error {
	// If the EBlock is already populated then there is nothing to do.
	if eb.IsPopulated() {
		return nil
	}

	// If we don't have a KeyMR, fetch the chain head KeyMR.
	if eb.KeyMR == nil {
		// If the KeyMR and ChainID are both nil we have nothing to
		// query for.
		if eb.ChainID == nil {
			return fmt.Errorf("no KeyMR or ChainID specified")
		}
		if err := eb.GetChainHead(c); err != nil {
			return err
		}
	}

	// Make RPC request for this Entry Block.
	params := struct {
		KeyMR *Bytes32 `json:"keymr"`
	}{KeyMR: eb.KeyMR}
	method := "entry-block"
	if err := c.FactomdRequest(method, params, eb); err != nil {
		return err
	}

	// Populate the ChainID and Height for all Entries.
	for i := range eb.Entries {
		eb.Entries[i].ChainID = eb.ChainID
		eb.Entries[i].Height = eb.Height
	}
	return nil
}

// GetChainHead queries factomd for the latest eb.KeyMR for chain eb.ChainID.
func (eb *EBlock) GetChainHead(c *Client) error {
	params := eb
	method := "chain-head"
	result := struct {
		KeyMR *Bytes32 `json:"chainhead"`
	}{}
	if err := c.FactomdRequest(method, params, &result); err != nil {
		return err
	}
	eb.KeyMR = result.KeyMR
	return nil
}

// IsFirst returns true if this is the first EBlock in its chain, indicated by
// the PrevKeyMR being all zeroes. IsFirst returns false if eb is not populated
// or if the PrevKeyMR is not all zeroes.
func (eb EBlock) IsFirst() bool {
	return eb.IsPopulated() && *eb.PrevKeyMR == zeroBytes32
}

// Prev returns the an EBlock with its KeyMR initialized to eb.PrevKeyMR and
// ChainID initialized to eb.ChainID. If eb is the first Entry Block in the
// chain, then eb is returned.
func (eb EBlock) Prev() EBlock {
	if eb.IsFirst() {
		return eb
	}
	return EBlock{ChainID: eb.ChainID, KeyMR: eb.PrevKeyMR}
}

// GetAllPrev returns a slice of all preceding EBlocks in eb's chain, in order
// from earliest to latest, up to and including eb. So the last element of the
// returned slice is always equal to eb. If eb is the first entry block in its
// chain, then it is the only element in the slice.
//
// If you are only interested in obtaining the first entry block in eb's chain,
// and not all of the intermediary ones, then use GetFirst to reduce network
// calls and memory usage.
func (eb EBlock) GetAllPrev(c *Client) ([]EBlock, error) {
	ebs := []EBlock{eb}
	for ; !ebs[0].IsFirst(); ebs = append([]EBlock{ebs[0].Prev()}, ebs...) {
		if err := ebs[0].Get(c); err != nil {
			return nil, err
		}
	}
	return ebs, nil
}

// GetFirst finds the first Entry Block in eb's chain, and populates eb as
// such.
//
// GetFirst differs from GetAllPrev in that it does not allocate any additional
// EBlocks. GetFirst avoids allocating any new EBlocks by reusing eb to
// traverse up to the first entry block.
func (eb *EBlock) GetFirst(c *Client) error {
	for ; !eb.IsFirst(); *eb = eb.Prev() {
		if err := eb.Get(c); err != nil {
			return err
		}
	}
	return nil
}
