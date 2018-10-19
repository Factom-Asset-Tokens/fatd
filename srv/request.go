package srv

import (
	"github.com/Factom-Asset-Tokens/fatd/factom"
)

type Token struct {
	ChainID *factom.Bytes32 `json:"chain-id,omitempty"`

	ID            *factom.Bytes32 `json:"token-id,omitempty"`
	IssuerChainID *factom.Bytes32 `json:"issuer-id,omitempty"`
}
