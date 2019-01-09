package fat1

import "fmt"

// NFTokenID is a Non-Fungible Token ID.
type NFTokenID uint64

// Set id in nfTkns and return an error if it is already set.
func (id NFTokenID) Set(nfTkns NFTokens) error {
	if _, ok := nfTkns[id]; ok {
		return fmt.Errorf("duplicate NFTokenID: %v", id)
	}
	nfTkns[id] = struct{}{}
	return nil
}

// Len returns 1.
func (id NFTokenID) Len() int {
	return 1
}

// JSONLen returns the expected JSON encoded length of id.
func (id NFTokenID) JSONLen() int {
	l := 1
	for pow := NFTokenID(10); id/pow > 0; pow *= 10 {
		l++
	}
	return l
}
