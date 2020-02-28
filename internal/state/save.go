package state

import (
	"crawshaw.io/sqlite/sqlitex"
	"github.com/Factom-Asset-Tokens/fatd/internal/db/address"
	"github.com/Factom-Asset-Tokens/fatd/internal/db/entry"
	"github.com/Factom-Asset-Tokens/fatd/internal/db/metadata"
)

// Save is implemented in the same way for all types but without generics this
// code just has to be duplicated. Putting them in the same file ensures
// consistency.

// Save the current state of the chain and the database that can be rolled back
// to if the returned closure is called with a non-nil error.
func (chain *FactomChain) Save() func(err *error) {
	rollback := sqlitex.Save(chain.Conn)
	chainCopy := *chain
	saveDepth := chain.SaveDepth
	chain.SaveDepth++
	return func(err *error) {
		ch := chain.Conn.SetInterrupt(nil)
		defer rollback(err)
		defer func() {
			if *err != nil {
				*chain = chainCopy
			}
			chain.Conn.SetInterrupt(ch)
		}()
		chain.SaveDepth--
		if saveDepth > 0 || *err != nil {
			return
		}
		if *err = entry.Commit(chain.Conn); *err != nil {
			return
		}
		// *err == nil, commit all changes
	}
}

// Save the current state of the chain and the database that can be rolled back
// to if the returned closure is called with a non-nil error.
func (chain *FATChain) Save() func(err *error) {
	rollback := sqlitex.Save(chain.Conn)
	chainCopy := *chain
	saveDepth := chain.SaveDepth
	chain.SaveDepth++
	return func(err *error) {
		ch := chain.Conn.SetInterrupt(nil)
		defer rollback(err)
		defer func() {
			if *err != nil {
				*chain = chainCopy
			}
			chain.Conn.SetInterrupt(ch)
		}()
		chain.SaveDepth--
		if saveDepth > 0 || *err != nil {
			return
		}
		if *err = entry.Commit(chain.Conn); *err != nil {
			return
		}
		if *err = address.Commit(chain.Conn); *err != nil {
			return
		}
		if *err = metadata.Commit(chain.Conn); *err != nil {
			return
		}
	}
}
