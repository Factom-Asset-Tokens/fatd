package fat1

import (
	"encoding/json"
	"fmt"
	"sort"
)

const MaxCapacity = 4e5

var maxCapacity int = MaxCapacity

var ErrorCapacity = fmt.Errorf("NFTokenID max capacity (%v) exceeded", maxCapacity)

// NFTokens are a set of unique NFTokenIDs. A map[NFTokenID]struct{} is used to
// guarantee uniqueness of NFTokenIDs.
type NFTokens map[NFTokenID]struct{}

// NFTokensSetter is an interface implemented by types that can set the
// NFTokenIDs they represent in a given NFTokens.
type NFTokensSetter interface {
	// Set the NFTokenIDs in tkns. Return an error if tkns already
	// contains one of the NFTokenIDs.
	Set(tkns NFTokens) error
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
	tkns := make(NFTokens, capacity)
	if err := tkns.Set(ids...); err != nil {
		return nil, err
	}
	return tkns, nil
}

func (tkns NFTokens) Append(newTkns NFTokens) error {
	if len(tkns)+len(newTkns) > maxCapacity {
		return ErrorCapacity
	}
	if err := tkns.NoIntersection(newTkns); err != nil {
		return err
	}
	for tknID := range newTkns {
		tkns[tknID] = struct{}{}
	}
	return nil
}

// Set all ids in tkns. Return an error if ids contains any duplicate or
// previously set NFTokenIDs.
func (tkns NFTokens) Set(ids ...NFTokensSetter) error {
	for _, id := range ids {
		if err := id.Set(tkns); err != nil {
			return err
		}
	}
	return nil
}

type ErrorNFTokenIDIntersection NFTokenID

func (id ErrorNFTokenIDIntersection) Error() string {
	return fmt.Sprintf("duplicate NFTokenID: %v", NFTokenID(id))
}

func (tkns NFTokens) NoIntersection(tknsCmp NFTokens) error {
	small, large := tkns, tknsCmp
	if len(small) > len(large) {
		small, large = large, small
	}
	for tknID := range small {
		if _, ok := large[tknID]; ok {
			return ErrorNFTokenIDIntersection(tknID)
		}
	}
	return nil
}

type ErrorMissingNFTokenID NFTokenID

func (id ErrorMissingNFTokenID) Error() string {
	return fmt.Sprintf("missing NFTokenID: %v", NFTokenID(id))
}

func (tkns NFTokens) ContainsAll(tknsSub NFTokens) error {
	if len(tknsSub) > len(tkns) {
		return fmt.Errorf("cannot contain a bigger NFTokens set")
	}
	for tknID := range tknsSub {
		if _, ok := tkns[tknID]; !ok {
			return ErrorMissingNFTokenID(tknID)
		}
	}
	return nil
}

// Slice returns a sorted slice of tkns' NFTokenIDs.
func (tkns NFTokens) Slice() []NFTokenID {
	tknsAry := make([]NFTokenID, len(tkns))
	i := 0
	for tknID := range tkns {
		tknsAry[i] = tknID
		i++
	}
	sort.Slice(tknsAry, func(i, j int) bool {
		return tknsAry[i] < tknsAry[j]
	})
	return tknsAry
}

// MarshalJSON implements the json.Marshaler interface. MarshalJSON will always
// produce the most efficient representation of tkns using NFTokenIDRanges
// over individual NFTokenIDs where appropriate. MarshalJSON will return an
// error if tkns is empty.
func (tkns NFTokens) MarshalJSON() ([]byte, error) {
	if len(tkns) == 0 {
		return nil, fmt.Errorf("%T: empty", tkns)
	}

	tknsFullAry := tkns.Slice()

	// Compress the tknsAry by replacing contiguous id ranges with an
	// NFTokenIDRange.
	tknsAry := make([]interface{}, len(tkns))
	idRange := NewNFTokenIDRange(tknsFullAry[0])
	i := 0
	for _, id := range append(tknsFullAry[1:], 0) {
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
			tknsAry[i] = idRange
			i++
		} else {
			for id := idRange.Min; id <= idRange.Max; id++ {
				tknsAry[i] = id
				i++
			}
		}
		idRange = NewNFTokenIDRange(id)
	}

	return json.Marshal(tknsAry[:i])
}

func (tkns *NFTokens) UnmarshalJSON(data []byte) error {
	var tknsJSONAry []json.RawMessage
	if err := json.Unmarshal(data, &tknsJSONAry); err != nil {
		return fmt.Errorf("%T: %v", tkns, err)
	}
	if len(tknsJSONAry) == 0 {
		return fmt.Errorf("%T: empty", tkns)
	}
	*tkns = make(NFTokens, len(tknsJSONAry))
	for _, data := range tknsJSONAry {
		var ids NFTokensSetter
		if data[0] == '{' {
			var idRange NFTokenIDRange
			if err := idRange.UnmarshalJSON(data); err != nil {
				return fmt.Errorf("%T: %v", tkns, err)
			}
			ids = idRange
		} else {
			var id NFTokenID
			if err := json.Unmarshal(data, &id); err != nil {
				return fmt.Errorf("%T: %v", tkns, err)
			}
			ids = id
		}
		if err := ids.Set(*tkns); err != nil {
			return fmt.Errorf("%T: %v", tkns, err)
		}
	}
	return nil

}
