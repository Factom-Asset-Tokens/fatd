package fat1

import (
	"fmt"

	"github.com/Factom-Asset-Tokens/fatd/fat/jsonlen"
)

// NFTokenID is a Non-Fungible Token ID.
type NFTokenID uint64

// Set id in nfTkns and return an error if it is already set.
func (id NFTokenID) Set(tkns NFTokens) error {
	if len(tkns)+id.Len() > maxCapacity {
		return fmt.Errorf("%T(len:%v): %T(%v): %v",
			tkns, len(tkns), id, id, ErrorCapacity)
	}
	if _, ok := tkns[id]; ok {
		return fmt.Errorf("duplicate NFTokenID: %v", id)
	}
	tkns[id] = struct{}{}
	return nil
}

// Len returns 1.
func (id NFTokenID) Len() int {
	return 1
}

func (id NFTokenID) jsonLen() int {
	return jsonlen.Uint64(uint64(id))
}
