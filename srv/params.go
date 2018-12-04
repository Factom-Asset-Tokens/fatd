package srv

import (
	jrpc "github.com/AdamSLevy/jsonrpc2/v9"
	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat0"
)

var (
	NoParamsError      = jrpc.NewInvalidParamsError(`no "params" accepted`)
	TokenNotFoundError = jrpc.NewError(-32800, "Token Not Found",
		"not yet issued or not tracked by this instance of fatd")
	TransactionNotFoundError = jrpc.NewError(-32800, "Token Not Found",
		"not yet issued or not tracked by this instance of fatd")
)

type Params interface {
	IsValid() bool
	ValidChainID() *factom.Bytes32
	Error() jrpc.Error
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

var TokenParamsError = jrpc.NewInvalidParamsError(
	`"params" required: either "chain-id" or both "token-id" and "issuer-id"`)

func (p TokenParams) Error() jrpc.Error {
	return TokenParamsError
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

var GetTransactionParamsError = jrpc.NewInvalidParamsError(
	`"params" required: "hash" and either "chain-id" or both "token-id" and "issuer-id"`)

func (p GetTransactionParams) Error() jrpc.Error {
	return GetTransactionParamsError
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

var GetTransactionsParamsError = jrpc.NewInvalidParamsError(
	`"params" required: "hash" or "start" and either "chain-id" or both "token-id" and "issuer-id", "limit" must be greater than 0 if provided`)

func (p GetTransactionsParams) Error() jrpc.Error {
	return GetTransactionsParamsError
}

type GetNFTokenParams struct {
	TokenParams
	NonFungibleTokenID *string `json:"nf-token-id,omitempty"`
}

func (p GetNFTokenParams) IsValid() bool {
	return p.NonFungibleTokenID != nil
}

var GetNFTokenParamsError = jrpc.NewInvalidParamsError(
	`"params" required: "nf-token-id" and either "chain-id" or both "token-id" and "issuer-id"`)

func (p GetNFTokenParams) Error() jrpc.Error {
	return GetNFTokenParamsError
}

type GetBalanceParams struct {
	TokenParams
	Address *factom.Address `json:"fa-address,omitempty"`
}

func (p GetBalanceParams) IsValid() bool {
	return p.Address != nil
}

var GetBalanceParamsError = jrpc.NewInvalidParamsError(
	`"params" required: "fa-address" and either "chain-id" or both "token-id" and "issuer-id"`)

func (p GetBalanceParams) Error() jrpc.Error {
	return GetBalanceParamsError
}

type SendTransactionParams struct {
	TokenParams
	ExtIDs  []factom.Bytes `json:"rcd-sigs"`
	Content factom.Bytes   `json:"tx"`
}

func (p SendTransactionParams) IsValid() bool {
	return len(p.Content) > 0 && len(p.ExtIDs) > 0
}

var SendTransactionParamsError = jrpc.NewInvalidParamsError(
	`"params" required: "rcd-sigs" and "tx" and either "chain-id" or both "token-id" and "issuer-id"`)

func (p SendTransactionParams) Error() jrpc.Error {
	return SendTransactionParamsError
}
