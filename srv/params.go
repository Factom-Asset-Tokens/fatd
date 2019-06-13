// MIT License
//
// Copyright 2018 Canonical Ledgers, LLC
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
// IN THE SOFTWARE.

package srv

import (
	"strings"
	"time"

	jrpc "github.com/AdamSLevy/jsonrpc2/v11"
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
	chainID := fat.ChainID(p.TokenID, *p.IssuerChainID)
	p.ChainID = &chainID
	return p.ChainID
}

func (p ParamsToken) Error() jrpc.Error {
	return *ParamsErrorToken
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
	return *ParamsErrorGetTransaction
}

type ParamsGetTransactions struct {
	ParamsToken
	// Transaction filters
	NFTokenID *fat1.NFTokenID    `json:"nftokenid,omitempty"`
	Addresses []factom.FAAddress `json:"addresses,omitempty"`
	StartHash *factom.Bytes32    `json:"entryhash,omitempty"`
	ToFrom    string             `json:"tofrom,omitempty"`
	Order     string             `json:"order,omitempty"`

	// Pagination
	Page  *uint64 `json:"page,omitempty"`
	Limit *uint64 `json:"limit,omitempty"`
}

func (p *ParamsGetTransactions) IsValid() bool {
	if p.Page == nil {
		p.Page = new(uint64)
	}
	if p.Limit == nil {
		p.Limit = new(uint64)
		*p.Limit = 25
	}
	if *p.Limit == 0 {
		return false
	}
	if p.Addresses != nil && len(p.Addresses) == 0 {
		return false
	}
	p.ToFrom = strings.ToLower(p.ToFrom)
	switch p.ToFrom {
	case "to", "from":
		if p.Addresses == nil {
			return false
		}
	case "":
	default:
		return false
	}
	p.Order = strings.ToLower(p.Order)
	switch p.Order {
	case "", "asc", "desc":
	default:
		return false
	}
	return true
}

func (p ParamsGetTransactions) Error() jrpc.Error {
	return *ParamsErrorGetTransactions
}

type ParamsGetNFToken struct {
	ParamsToken
	NFTokenID *fat1.NFTokenID `json:"nftokenid"`
}

func (p ParamsGetNFToken) IsValid() bool {
	return p.NFTokenID != nil
}

func (p ParamsGetNFToken) Error() jrpc.Error {
	return *ParamsErrorGetNFToken
}

type ParamsGetBalance struct {
	ParamsToken
	Address *factom.FAAddress `json:"address,omitempty"`
}

func (p ParamsGetBalance) IsValid() bool {
	return p.Address != nil
}

func (p ParamsGetBalance) Error() jrpc.Error {
	return *ParamsErrorGetBalance
}

type ParamsGetNFBalance struct {
	ParamsToken
	Address *factom.FAAddress `json:"address,omitempty"`

	// Pagination
	Page  *uint64 `json:"page,omitempty"`
	Limit *uint64 `json:"limit,omitempty"`
	Order string  `json:"order,omitempty"`
}

func (p *ParamsGetNFBalance) IsValid() bool {
	if p.Page == nil {
		p.Page = new(uint64)
	}
	if p.Limit == nil {
		p.Limit = new(uint64)
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
	return *ParamsErrorGetBalance
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
	return *ParamsErrorGetBalance
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
	return *ParamsErrorSendTransaction
}

func (p ParamsSendTransaction) Entry() factom.Entry {
	return factom.Entry{
		ExtIDs:    p.ExtIDs,
		Content:   p.Content,
		Timestamp: time.Now(),
		ChainID:   p.ChainID,
	}
}
