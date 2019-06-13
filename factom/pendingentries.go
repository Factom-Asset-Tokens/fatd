package factom

import (
	"bytes"
	"sort"
)

// PendingEntries is a list of pending entries which may or may not be
// revealed. If the entry's ChainID is not nil, then its data has been revealed
// and can be queried from factomd.
type PendingEntries []Entry

// Get returns all pending entries sorted by ChainID, and then order they were
// originally returned.
func (pe *PendingEntries) Get(c *Client) error {
	if err := c.FactomdRequest("pending-entries", nil, pe); err != nil {
		return err
	}
	sort.SliceStable(*pe, func(i, j int) bool {
		pe := *pe
		var ci, cj []byte
		ei, ej := pe[i], pe[j]
		if ei.ChainID != nil {
			ci = ei.ChainID[:]
		}
		if ej.ChainID != nil {
			cj = ej.ChainID[:]
		}
		return bytes.Compare(ci, cj) < 0
	})
	return nil
}

// Entries efficiently finds and returns all entries in pe for the given
// chainID, if any exist. Otherwise, Entries returns nil.
func (pe PendingEntries) Entries(chainID Bytes32) []Entry {
	// Find the first index of the entry with this chainID.
	ei := sort.Search(len(pe), func(i int) bool {
		var c []byte
		e := pe[i]
		if e.ChainID != nil {
			c = e.ChainID[:]
		}
		return bytes.Compare(c, chainID[:]) >= 0
	})
	if ei < len(pe) && *pe[ei].ChainID == chainID {
		// Find all remaining entries with the chainID.
		for i, e := range pe[ei:] {
			if *e.ChainID != chainID {
				return pe[ei : ei+i]
			}
		}
		return pe[ei:]
	}
	// There are no entries for this ChainID.
	return nil
}
