package state

import (
	"fmt"

	"crawshaw.io/sqlite/sqlitex"
)

// Save is implemented in the same way for all types but without generics this
// code just has to be duplicated. Putting them in the same file ensures
// consistency.

// Save the current state of the chain and the database that can be rolled back
// to if the returned closure is called with a non-nil error.
func (chain *FactomChain) Save() func(err *error) {
	rollback := sqlitex.Save(chain.Conn)
	chainCopy := *chain
	return func(err *error) {
		if *err != nil {
			// Reset chain on any error
			var alwaysRollbackErr = fmt.Errorf("always rollback")
			rollback(&alwaysRollbackErr)
			*chain = chainCopy
			return
		}
		// *err == nil, commit all changes
		rollback(err)
	}
}

// Save the current state of the chain and the database that can be rolled back
// to if the returned closure is called with a non-nil error.
func (chain *FATChain) Save() func(err *error) {
	rollback := sqlitex.Save(chain.Conn)
	chainCopy := *chain
	return func(err *error) {
		if *err != nil {
			// Reset chain on any error
			var alwaysRollbackErr = fmt.Errorf("always rollback")
			rollback(&alwaysRollbackErr)
			*chain = chainCopy
			return
		}
		// *err == nil, commit all changes
		rollback(err)
	}
}
