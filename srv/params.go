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
	"github.com/Factom-Asset-Tokens/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat"
	"github.com/Factom-Asset-Tokens/fatd/fat/fat1"
)

type Params interface {
	IsValid() error
	ValidChainID() *factom.Bytes32
	HasIncludePending() bool
}

// ParamsToken scopes a request down to a single FAT token using either the
// ChainID or both the TokenID and the IssuerChainID.
type ParamsToken struct {
	ChainID       *factom.Bytes32 `json:"chainid,omitempty"`
	TokenID       string          `json:"tokenid,omitempty"`
	IssuerChainID *factom.Bytes32 `json:"issuerid,omitempty"`

	IncludePending bool `json:"includepending,omitempty"`
}

func (p ParamsToken) IsValid() error {
	if p.ChainID != nil {
		if len(p.TokenID) > 0 || p.IssuerChainID != nil {
			return jrpc.InvalidParams(
				`cannot use "chainid" with "tokenid" or "issuerid"`)
		}
		return nil
	}
	if len(p.TokenID) > 0 || p.IssuerChainID != nil {
		if len(p.TokenID) == 0 {
			return jrpc.InvalidParams(
				`"tokenid" is required with "issuerid"`)
		}
		if p.IssuerChainID == nil {
			return jrpc.InvalidParams(
				`"issuerid" is required with "tokenid"`)
		}
		return nil
	}
	return jrpc.InvalidParams(
		`required: either "chainid" or both "tokenid" and "issuerid"`)
}

func (p ParamsToken) HasIncludePending() bool { return p.IncludePending }

func (p ParamsToken) ValidChainID() *factom.Bytes32 {
	if p.ChainID != nil {
		return p.ChainID
	}
	chainID := fat.ComputeChainID(p.TokenID, p.IssuerChainID)
	p.ChainID = &chainID
	return p.ChainID
}

type ParamsPagination struct {
	Page  *uint  `json:"page,omitempty"`
	Limit uint   `json:"limit,omitempty"`
	Order string `json:"order,omitempty"`
}

func (p *ParamsPagination) IsValid() error {
	if p.Page == nil {
		p.Page = new(uint)
		*p.Page = 1
	} else if *p.Page == 0 {
		return jrpc.InvalidParams(
			`"order" value must be either "asc" or "desc"`)
	}
	if p.Limit == 0 {
		p.Limit = 25
	}

	p.Order = strings.ToLower(p.Order)
	switch p.Order {
	case "", "asc", "desc":
		// ok
	default:
		return jrpc.InvalidParams(
			`"order" value must be either "asc" or "desc"`)
	}

	return nil
}

// ParamsGetTransaction is used to query for a single particular transaction
// with the given Entry Hash.
type ParamsGetTransaction struct {
	ParamsToken
	Hash *factom.Bytes32 `json:"entryhash"`
}

func (p ParamsGetTransaction) IsValid() error {
	if err := p.ParamsToken.IsValid(); err != nil {
		return err
	}
	if p.Hash == nil {
		return jrpc.InvalidParams(`required: "entryhash"`)
	}
	return nil
}

type ParamsGetTransactions struct {
	ParamsToken
	ParamsPagination
	// Transaction filters
	NFTokenID *fat1.NFTokenID    `json:"nftokenid,omitempty"`
	Addresses []factom.FAAddress `json:"addresses,omitempty"`
	StartHash *factom.Bytes32    `json:"entryhash,omitempty"`
	ToFrom    string             `json:"tofrom,omitempty"`
}

func (p *ParamsGetTransactions) IsValid() error {
	if err := p.ParamsToken.IsValid(); err != nil {
		return err
	}
	if err := p.ParamsPagination.IsValid(); err != nil {
		return err
	}

	p.ToFrom = strings.ToLower(p.ToFrom)
	switch p.ToFrom {
	case "to", "from":
		if len(p.Addresses) == 0 {
			return jrpc.InvalidParams(
				`"addresses" may not be empty when "tofrom" is set`)
		}
	case "":
		// empty is ok
	default:
		return jrpc.InvalidParams(
			`"tofrom" value must be either "to" or "from"`)
	}
	return nil
}

type ParamsGetNFToken struct {
	ParamsToken
	NFTokenID *fat1.NFTokenID `json:"nftokenid"`
}

func (p ParamsGetNFToken) IsValid() error {
	if err := p.ParamsToken.IsValid(); err != nil {
		return err
	}
	if p.NFTokenID == nil {
		return jrpc.InvalidParams(`required: "nftokenid"`)
	}
	return nil
}

type ParamsGetBalance struct {
	ParamsToken
	Address *factom.FAAddress `json:"address,omitempty"`
}

func (p ParamsGetBalance) IsValid() error {
	if err := p.ParamsToken.IsValid(); err != nil {
		return err
	}
	if p.Address == nil {
		return jrpc.InvalidParams(`required: "address"`)
	}
	return nil
}

type ParamsGetBalances struct {
	Address        *factom.FAAddress `json:"address,omitempty"`
	IncludePending bool              `json:"includepending,omitempty"`
}

func (p ParamsGetBalances) HasIncludePending() bool { return p.IncludePending }

func (p ParamsGetBalances) IsValid() error {
	if p.Address == nil {
		return jrpc.InvalidParams(`required: "address"`)
	}
	return nil
}
func (p ParamsGetBalances) ValidChainID() *factom.Bytes32 {
	return nil
}

type ParamsGetNFBalance struct {
	ParamsToken
	ParamsPagination
	Address *factom.FAAddress `json:"address,omitempty"`
}

func (p *ParamsGetNFBalance) IsValid() error {
	if err := p.ParamsToken.IsValid(); err != nil {
		return err
	}
	if err := p.ParamsPagination.IsValid(); err != nil {
		return err
	}
	if p.Address == nil {
		return jrpc.InvalidParams(`required: "address"`)
	}
	return nil
}

type ParamsGetAllNFTokens struct {
	ParamsToken
	ParamsPagination
}

func (p *ParamsGetAllNFTokens) IsValid() error {
	if err := p.ParamsToken.IsValid(); err != nil {
		return err
	}
	if err := p.ParamsPagination.IsValid(); err != nil {
		return err
	}
	return nil
}

type ParamsSendTransaction struct {
	ParamsToken
	ExtIDs  []factom.Bytes `json:"extids,omitempty"`
	Content factom.Bytes   `json:"content,omitempty"`
	Raw     factom.Bytes   `json:"raw,omitempty"`
	DryRun  bool           `json:"dryrun,omitempty"`
	entry   factom.Entry
}

func (p *ParamsSendTransaction) IsValid() error {
	if p.Raw != nil {
		if p.ExtIDs != nil || p.Content != nil ||
			p.ParamsToken != (ParamsToken{}) {
			return jrpc.InvalidParams(
				`"raw cannot be used with "content" or "extids"`)
		}
		if err := p.entry.UnmarshalBinary(p.Raw); err != nil {
			return jrpc.InvalidParams(err.Error())
		}
		p.entry.Timestamp = time.Now()
		p.ChainID = p.entry.ChainID
		return nil
	}
	if err := p.ParamsToken.IsValid(); err != nil {
		return err
	}
	if len(p.Content) == 0 || len(p.ExtIDs) == 0 {
		return jrpc.InvalidParams(`required: "raw" or "content" and "extids"`)
	}
	p.entry = factom.Entry{
		ExtIDs:    p.ExtIDs,
		Content:   p.Content,
		Timestamp: time.Now(),
		ChainID:   p.ChainID,
	}

	data, err := p.entry.MarshalBinary()
	if err != nil {
		return jrpc.InvalidParams(err)
	}
	hash := factom.ComputeEntryHash(data)
	p.entry.Hash = &hash
	p.Raw = data

	return nil
}

func (p ParamsSendTransaction) Entry() factom.Entry {
	return p.entry
}
