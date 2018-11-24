package state

import (
	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat0"
)

func GetBalance(chainID *factom.Bytes32, address *factom.Address) uint64 {
	return 0
}

func GetIssuance(chainID *factom.Bytes32) *fat0.Issuance {
	chain := chains.Get(chainID)
	if !chain.Issued() {
		return nil
	}
	return &chain.Issuance
}
