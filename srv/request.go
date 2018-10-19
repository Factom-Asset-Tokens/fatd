package srv

import (
	"github.com/Factom-Asset-Tokens/fatd/factom"
)

type TokenParams struct {

	//base token params
	ChainID       *factom.Bytes32 `json:"chain-id,omitempty"`
	TokenID       *string         `json:"token-id,omitempty"`
	IssuerChainID *factom.Bytes32 `json:"issuer-id,omitempty"`

	//query params
	TransactionID      *factom.Bytes32 `json:"tx-id,omitempty"`
	FactoidAddress     *factom.Address `json:"fa-address,omitempty"`
	NonFungibleTokenID *string         `json:"nf-token-address,omitempty"`

	//pagination
	Page  *int `json:"page,omitempty"`
	Limit *int `json:"limit,omitempty"`
}
