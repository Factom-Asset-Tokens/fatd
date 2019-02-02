package srv

import (
	"strings"
	"time"

	jrpc "github.com/AdamSLevy/jsonrpc2/v10"
	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat"
	"github.com/Factom-Asset-Tokens/fatd/fat/fat1"
)

type Params interface {
	IsValid() bool
	ValidChainID() *factom.Bytes32
	Error() jrpc.Error
}

// ParamsToken scopes a request down to a single FAT token using either the
// ChainID or both the TokenID and the IssuerChainID.
type ParamsToken struct {
	ChainID       *factom.Bytes32 `json:"chainid,omitempty"`
	TokenID       string          `json:"tokenid,omitempty"`
	IssuerChainID *factom.Bytes32 `json:"issuerid,omitempty"`
}

func (p ParamsToken) IsValid() bool {
	if (p.ChainID != nil && len(p.TokenID) == 0 && p.IssuerChainID == nil) ||
		(p.ChainID == nil && len(p.TokenID) != 0 && p.IssuerChainID != nil) {
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
	chainID := fat.ChainID(p.TokenID, p.IssuerChainID)
	p.ChainID = &chainID
	return p.ChainID
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
	// Transaction filters
	NonFungibleTokenID string          `json:"nftokenid,omitempty"`
	StartHash          *factom.Bytes32 `json:"entryhash,omitempty"`
	FactoidAddress     *factom.RCDHash `json:"address,omitempty"`
	ToFrom             string          `json:"tofrom"`

	// Pagination
	Page  *uint `json:"page,omitempty"`
	Limit *uint `json:"limit,omitempty"`
}

func (p *ParamsGetTransactions) IsValid() bool {
	if p.Page == nil {
		p.Page = new(uint)
	}
	if p.Limit == nil {
		p.Limit = new(uint)
		*p.Limit = 25
	}
	if *p.Limit == 0 {
		return false
	}
	p.ToFrom = strings.ToLower(p.ToFrom)
	switch p.ToFrom {
	case "to":
	case "from":
	case "":
	default:
		return false
	}
	return true
}

func (p ParamsGetTransactions) Error() jrpc.Error {
	return ParamsErrorGetTransactions
}

type ParamsGetNFToken struct {
	ParamsToken
	NFTokenID *fat1.NFTokenID `json:"nftokenid"`
}

func (p ParamsGetNFToken) IsValid() bool {
	return p.NFTokenID != nil
}

func (p ParamsGetNFToken) Error() jrpc.Error {
	return ParamsErrorGetNFToken
}

type ParamsGetBalance struct {
	ParamsToken
	Address *factom.RCDHash `json:"address,omitempty"`
}

func (p ParamsGetBalance) IsValid() bool {
	return p.Address != nil
}

func (p ParamsGetBalance) Error() jrpc.Error {
	return ParamsErrorGetBalance
}

type ParamsGetNFBalance struct {
	ParamsToken
	Address *factom.RCDHash `json:"address,omitempty"`

	// Pagination
	Page  *uint  `json:"page,omitempty"`
	Limit *uint  `json:"limit,omitempty"`
	Order string `json:"order,omitempty"`
}

func (p *ParamsGetNFBalance) IsValid() bool {
	if p.Page == nil {
		p.Page = new(uint)
	}
	if p.Limit == nil {
		p.Limit = new(uint)
		*p.Limit = 25
	}
	if *p.Limit == 0 {
		return false
	}
	p.Order = strings.ToLower(p.Order)
	switch p.Order {
	case "asc":
	case "desc":
	case "":
	default:
		return false
	}
	return p.Address != nil
}

func (p ParamsGetNFBalance) Error() jrpc.Error {
	return ParamsErrorGetBalance
}

type ParamsGetAllNFTokens struct {
	ParamsToken

	// Pagination
	Page  *uint  `json:"page,omitempty"`
	Limit *uint  `json:"limit,omitempty"`
	Order string `json:"order,omitempty"`
}

func (p *ParamsGetAllNFTokens) IsValid() bool {
	if p.Page == nil {
		p.Page = new(uint)
	}
	if p.Limit == nil {
		p.Limit = new(uint)
		*p.Limit = 25
	}
	if *p.Limit == 0 {
		return false
	}
	p.Order = strings.ToLower(p.Order)
	switch p.Order {
	case "asc":
	case "desc":
	case "":
	default:
		return false
	}
	return true
}

func (p ParamsGetAllNFTokens) Error() jrpc.Error {
	return ParamsErrorGetBalance
}

type ParamsSendTransaction struct {
	ParamsToken
	ExtIDs  []factom.Bytes `json:"extids"`
	Content factom.Bytes   `json:"content"`
}

func (p ParamsSendTransaction) IsValid() bool {
	return len(p.Content) > 0 && len(p.ExtIDs) > 0
}

func (p ParamsSendTransaction) Error() jrpc.Error {
	return ParamsErrorSendTransaction
}

func (p ParamsSendTransaction) Entry() factom.Entry {
	return factom.Entry{
		ExtIDs:    p.ExtIDs,
		Content:   p.Content,
		Timestamp: &factom.Time{Time: time.Now()},
		ChainID:   p.ChainID,
	}
}
