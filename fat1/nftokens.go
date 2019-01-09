package fat1

import (
	"encoding/json"
	"fmt"
	"sort"
)

// NFTokens are a set of unique NFTokenIDs. A map[NFTokenID]struct{} is used to
// guarantee uniqueness of NFTokenIDs.
type NFTokens map[NFTokenID]struct{}

// NFTokensSetter is an interface implemented by types that can set the
// NFTokenIDs they represent in a given NFTokens.
type NFTokensSetter interface {
	// Set the NFTokenIDs in nfTkns. Return an error if nfTkns already
	// contains one of the NFTokenIDs.
	Set(nfTkns NFTokens) error
	// Len returns number of NFTokenIDs that will be set.
	Len() int
}

// NewNFTokens returns an NFTokens initialized with ids. If ids contains any
// duplicate NFTokenIDs.
func NewNFTokens(ids ...NFTokensSetter) (NFTokens, error) {
	var capacity int
	for _, id := range ids {
		capacity += id.Len()
	}
	nfTkns := make(NFTokens, capacity)
	if err := nfTkns.Set(ids...); err != nil {
		return nil, err
	}
	return nfTkns, nil
}

// Set all ids in nfTkns. Return an error if ids contains any duplicate or
// previously set NFTokenIDs.
func (nfTkns NFTokens) Set(ids ...NFTokensSetter) error {
	for _, id := range ids {
		if err := id.Set(nfTkns); err != nil {
			return err
		}
	}
	return nil
}

func (nfTkns NFTokens) Intersects(nfTknsCmp NFTokens) error {
	small, large := nfTkns, nfTknsCmp
	if len(small) > len(large) {
		small, large = large, small
	}
	for nfTknID := range small {
		if _, ok := large[nfTknID]; ok {
			return fmt.Errorf("duplicate NFTokenID: %v", nfTknID)
		}
	}
	return nil
}

// Slice returns a sorted slice of nfTkns' NFTokenIDs.
func (nfTkns NFTokens) Slice() []NFTokenID {
	nfTknsAry := make([]NFTokenID, len(nfTkns))
	i := 0
	for nfTknID := range nfTkns {
		nfTknsAry[i] = nfTknID
		i++
	}
	sort.Slice(nfTknsAry, func(i, j int) bool {
		return nfTknsAry[i] < nfTknsAry[j]
	})
	return nfTknsAry
}

// MarshalJSON implements the json.Marshaler interface. MarshalJSON will always
// produce the most efficient representation of nfTkns using NFTokenIDRanges
// over individual NFTokenIDs where appropriate. MarshalJSON will return an
// error if nfTkns is empty.
func (nfTkns NFTokens) MarshalJSON() ([]byte, error) {
	if len(nfTkns) == 0 {
		return nil, fmt.Errorf("%T: empty", nfTkns)
	}

	nfTknsFullAry := nfTkns.Slice()

	// Compress the nfTknsAry by replacing contiguous id ranges with an
	// NFTokenIDRange.
	nfTknsAry := make([]interface{}, len(nfTkns))
	idRange := NewNFTokenIDRange(nfTknsFullAry[0])
	i := 0
	for _, id := range append(nfTknsFullAry[1:], 0) {
		// If this id is contiguous with idRange, expand the range to
		// include this id and check the next id.
		if id == idRange.Max+1 {
			idRange.Max = id
			continue
		}
		// Otherwise, the id is not contiguous with the range, so
		// append the idRange and set up a new idRange to start at id.

		// Use the most efficient JSON representation for the idRange.
		if idRange.IsEfficient() {
			nfTknsAry[i] = idRange
			i++
		} else {
			for _, id := range idRange.Slice() {
				nfTknsAry[i] = id
				i++
			}
		}
		idRange = NewNFTokenIDRange(id)
	}

	return json.Marshal(nfTknsAry[:i])
}

func (nfTkns *NFTokens) UnmarshalJSON(data []byte) error {
	var nfTknsJSONAry []json.RawMessage
	if err := json.Unmarshal(data, &nfTknsJSONAry); err != nil {
		return fmt.Errorf("%T: %v", nfTkns, err)
	}
	if len(nfTknsJSONAry) == 0 {
		return fmt.Errorf("%T: empty", nfTkns)
	}
	*nfTkns = make(NFTokens, len(nfTknsJSONAry))
	for _, data := range nfTknsJSONAry {
		var ids NFTokensSetter
		if data[0] == '{' {
			var idRange NFTokenIDRange
			if err := idRange.UnmarshalJSON(data); err != nil {
				return err
			}
			ids = idRange
		} else {
			var id NFTokenID
			if err := json.Unmarshal(data, &id); err != nil {
				return err
			}
			ids = id
		}
		if err := ids.Set(*nfTkns); err != nil {
			return fmt.Errorf("%T: %v", nfTkns, err)
		}
	}
	return nil

}
