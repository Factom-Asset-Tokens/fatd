package factom

import "fmt"

// EBlock represents an Factom Entry Block.
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
	return eb.PrevKeyMR != nil
}

// Get queries factomd for the Entry Block corresponding to eb.KeyMR, if not
// nil, and otherwise the Entry Block chain head for eb.ChainID. Either
// eb.KeyMR or eb.ChainID must be allocated or else Get will fail to populate
// the EBlock.
//
// Get returns any networking or marshaling errors, but not JSON RPC errors. To
// check if the EBlock has been successfully populated, call IsPopulated().
func (eb *EBlock) Get() error {
	// If the EBlock is already populated then there is nothing to do.
	if eb.IsPopulated() {
		return nil
	}
	// If the KeyMR and ChainID are both nil we have nothing to query for.
	if eb.KeyMR == nil && eb.ChainID == nil {
		return fmt.Errorf("KeyMR and ChainID are both nil")
	}

	// If we don't have a KeyMR, fetch the chain head's KeyMR.
	if eb.KeyMR == nil {
		if err := eb.GetChainHead(); err != nil {
			return err
		}
		// If we don't get a KeyMR back for the chain head then we just
		// return nil because the chain ID wasn't found and so we can't
		// populate the entry block.
		if eb.KeyMR == nil {
			return nil
		}
	}

	// Make RPC request for this Entry Block.
	params := map[string]interface{}{"keymr": eb.KeyMR}
	method := "entry-block"
	if err := factomdRequest(method, params, eb); err != nil {
		return err
	}

	// Populate the ChainID and Height for all Entries.
	for i := range eb.Entries {
		eb.Entries[i].ChainID = eb.ChainID
		eb.Entries[i].Height = eb.Height
	}
	return nil
}

func (eb *EBlock) GetChainHead() error {
	params := eb
	method := "chain-head"
	result := struct {
		KeyMR *Bytes32 `json:"chainhead"`
	}{}
	if err := factomdRequest(method, params, &result); err != nil {
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

// Prev returns the previous EBlock in eb's chain, without populating it with a
// call to Get. In other words, the KeyMR will be equal to eb.PrevKeyMR and the
// ChainID will be equal to eb.ChainID. If eb is the first Entry Block in the
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
// and not all of the intermediary ones, then use GetFirst to reduce memory
// usage.
//
// Like Get, GetAllPrev returns any networking or marshaling errors, but not
// JSON RPC errors. However, failing to populate any EBlock in the chain will
// result in returning a nil slice, thus it is unneccessary to call IsPopulated
// on any of the EBlocks in the returned slice.
func (eb EBlock) GetAllPrev() ([]EBlock, error) {
	ebs := []EBlock{eb}
	for ; !ebs[0].IsFirst(); ebs = append([]EBlock{ebs[0].Prev()}, ebs...) {
		if err := ebs[0].Get(); err != nil {
			return nil, err
		}
		if !ebs[0].IsPopulated() {
			return nil, nil
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
//
// Like Get, GetFirst returns any networking or marshaling errors, but not JSON
// RPC errors. To check if the EBlock has been successfully populated, call
// IsPopulated().
func (eb *EBlock) GetFirst() error {
	for ; !eb.IsFirst(); *eb = eb.Prev() {
		if err := eb.Get(); err != nil {
			return err
		}
		if !eb.IsPopulated() {
			return nil
		}
	}
	return nil
}
