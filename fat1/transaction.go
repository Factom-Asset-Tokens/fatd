package fat1

import "github.com/Factom-Asset-Tokens/fatd/fat0"

// Transaction represents a fat1 transaction, which can be a normal account
// transaction or a coinbase transaction depending on the Inputs and the
// RCD/signature pair.
type Transaction struct {
	Inputs  AddressNFTokensMap `json:"inputs"`
	Outputs AddressNFTokensMap `json:"outputs"`
	fat0.Entry
}
