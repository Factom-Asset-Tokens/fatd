package srv

import (
	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat0"
)

type Params interface {
	IsValid() bool
	ValidChainID() *factom.Bytes32
}

// TokenParams scopes a request down to a single FAT token using either the
// ChainID or both the TokenID and the IssuerChainID.
type TokenParams struct {
	ChainID       *factom.Bytes32 `json:"chain-id,omitempty"`
	TokenID       *string         `json:"token-id,omitempty"`
	IssuerChainID *factom.Bytes32 `json:"issuer-id,omitempty"`
}

func (p TokenParams) IsValid() bool {
	if (p.ChainID != nil && p.TokenID == nil && p.IssuerChainID == nil) ||
		(p.ChainID == nil && p.TokenID != nil && p.IssuerChainID != nil) {
		return true
	}
	return false
}

func (p TokenParams) ValidChainID() *factom.Bytes32 {
	if !p.IsValid() {
		return nil
	}
	if p.ChainID != nil {
		return p.ChainID
	}
	return fat0.ChainID(*p.TokenID, p.IssuerChainID)
}

// GetTransactionParams is used to query for a single particular transaction
// with the given Entry Hash.
type GetTransactionParams struct {
	TokenParams
	Hash *factom.Bytes32 `json:"entryhash"`
}

func (p GetTransactionParams) IsValid() bool {
	return p.Hash != nil
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

func (p *GetTransactionsParams) IsValid() bool {
	if p.Hash != nil {
		if p.Start != nil {
			return false
		}
	} else if p.Start == nil {
		p.Start = new(uint)
	}
	if p.Limit == nil {
		p.Limit = new(uint)
		*p.Limit = 25
	} else {
		if *p.Limit == 0 {
			return false
		}
	}
	return true
}

type GetNFTokenParams struct {
	TokenParams
	NonFungibleTokenID *string `json:"nf-token-id,omitempty"`
}

func (p GetNFTokenParams) IsValid() bool {
	return p.NonFungibleTokenID != nil
}

type GetBalanceParams struct {
	TokenParams
	Address *factom.Address `json:"fa-address,omitempty"`
}

func (p GetBalanceParams) IsValid() bool {
	return p.Address != nil
}

type SendTransactionParams struct {
	ExtIDs  []factom.Bytes `json:"rcd-sigs"`
	Content factom.Bytes   `json:"tx"`
}
