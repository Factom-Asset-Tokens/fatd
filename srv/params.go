package srv

import (
	jrpc "github.com/AdamSLevy/jsonrpc2/v9"
	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat0"
)

type Params interface {
	IsValid() bool
	ValidChainID() *factom.Bytes32
	Error() jrpc.Error
}

// ParamsToken scopes a request down to a single FAT token using either the
// ChainID or both the TokenID and the IssuerChainID.
type ParamsToken struct {
	ChainID       *factom.Bytes32 `json:"chain-id,omitempty"`
	TokenID       *string         `json:"token-id,omitempty"`
	IssuerChainID *factom.Bytes32 `json:"issuer-id,omitempty"`
}

func (p ParamsToken) IsValid() bool {
	if (p.ChainID != nil && p.TokenID == nil && p.IssuerChainID == nil) ||
		(p.ChainID == nil && p.TokenID != nil && p.IssuerChainID != nil) {
		return true
	}
	return false
}

func (p ParamsToken) ValidChainID() *factom.Bytes32 {
	if !p.IsValid() {
		return nil
	}
	if p.ChainID != nil {
		return p.ChainID
	}
	chainID := fat0.ChainID(*p.TokenID, p.IssuerChainID)
	return &chainID
}

func (p ParamsToken) Error() jrpc.Error {
	return ParamsErrorToken
}

// ParamsGetTransaction is used to query for a single particular transaction
// with the given Entry Hash.
type ParamsGetTransaction struct {
	ParamsToken
	Hash *factom.Bytes32 `json:"entryhash"`
}

func (p ParamsGetTransaction) IsValid() bool {
	return p.Hash != nil
}

func (p ParamsGetTransaction) Error() jrpc.Error {
	return ParamsErrorGetTransaction
}

type ParamsGetTransactions struct {
	ParamsToken
	NonFungibleTokenID *string         `json:"nf-token-id,omitempty"`
	FactoidAddress     *factom.Address `json:"fa-address,omitempty"`

	// Pagination
	Hash  *factom.Bytes32 `json:"entryhash,omitempty"`
	Start *uint           `json:"start,omitempty"`
	Limit *uint           `json:"limit,omitempty"`
}

func (p *ParamsGetTransactions) IsValid() bool {
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

func (p ParamsGetTransactions) Error() jrpc.Error {
	return ParamsErrorGetTransactions
}

type ParamsGetNFToken struct {
	ParamsToken
	NonFungibleTokenID *string `json:"nf-token-id,omitempty"`
}

func (p ParamsGetNFToken) IsValid() bool {
	return p.NonFungibleTokenID != nil
}

func (p ParamsGetNFToken) Error() jrpc.Error {
	return ParamsErrorGetNFToken
}

type ParamsGetBalance struct {
	ParamsToken
	Address *factom.Address `json:"fa-address,omitempty"`
}

func (p ParamsGetBalance) IsValid() bool {
	return p.Address != nil
}

func (p ParamsGetBalance) Error() jrpc.Error {
	return ParamsErrorGetBalance
}

type ParamsSendTransaction struct {
	ParamsToken
	ExtIDs  []factom.Bytes `json:"rcd-sigs"`
	Content factom.Bytes   `json:"tx"`
}

func (p ParamsSendTransaction) IsValid() bool {
	return len(p.Content) > 0 && len(p.ExtIDs) > 0
}

func (p ParamsSendTransaction) Error() jrpc.Error {
	return ParamsErrorSendTransaction
}