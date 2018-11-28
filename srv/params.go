package srv

import (
	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat0"
)

// TokenParams scopes a request down to a single FAT token using either the
// ChainID or both the TokenID and the IssuerChainID.
type TokenParams struct {
	ChainID       *factom.Bytes32 `json:"chain-id,omitempty"`
	TokenID       *string         `json:"token-id,omitempty"`
	IssuerChainID *factom.Bytes32 `json:"issuer-id,omitempty"`
}

func (t TokenParams) IsValid() bool {
	if (t.ChainID != nil && t.TokenID == nil && t.IssuerChainID == nil) ||
		(t.ChainID == nil && t.TokenID != nil && t.IssuerChainID != nil) {
		return true
	}
	return false
}

func (t TokenParams) ValidChainID() *factom.Bytes32 {
	if !t.IsValid() {
		return nil
	}
	if t.ChainID != nil {
		return t.ChainID
	}
	return fat0.ChainID(*t.TokenID, t.IssuerChainID)
}

// GetTransactionParams is used to query for a single particular transaction
// with the given Entry Hash.
type GetTransactionParams struct {
	TokenParams
	Hash *factom.Bytes32 `json:"entryhash"`
}

type GetTransactionsParams struct {
	TokenParams
	NonFungibleTokenID *string         `json:"nf-token-id,omitempty"`
	FactoidAddress     *factom.Address `json:"fa-address,omitempty"`

	// Pagination
	Hash  *factom.Bytes32 `json:"entryhash,omitempty"`
	Start *uint           `json:"start,omitempty"`
	Limit *uint           `json:"limit,omitempty"`
}

type GetNFTokenParams struct {
	TokenParams
	FactoidAddress *factom.Address `json:"fa-address,omitempty"`
}

type GetBalanceParams struct {
	TokenParams
	NonFungibleTokenID *string `json:"nf-token-id,omitempty"`
}

type SendTransactionParams struct {
	ExtIDs  []factom.Bytes `json:"rcd-sigs"`
	Content factom.Bytes   `json:"tx"`
}
