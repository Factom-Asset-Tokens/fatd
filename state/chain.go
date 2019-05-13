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

package state

import (
	"fmt"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat"
	"github.com/gocraft/dbr"
	"github.com/jinzhu/gorm"
)

type Chain struct {
	ID *factom.Bytes32
	ChainStatus
	factom.Identity
	fat.Issuance
	Metadata
	*gorm.DB
	DBR *dbr.Connection
}

func (chain Chain) String() string {
	return fmt.Sprintf("{ChainStatus:%v, ID:%v, Metadata:%+v, "+
		"fat.Identity:%+v, fat.Issuance:%+v}",
		chain.ChainStatus, chain.ID, chain.Metadata,
		chain.Identity, chain.Issuance)
}

func (chain *Chain) ignore() {
	chain.ID = nil
	chain.ChainStatus = ChainStatusIgnored
}
func (chain *Chain) track(first factom.Entry) error {
	chain.ChainStatus = ChainStatusTracked
	chain.Identity.ChainID = factom.NewBytes32(first.ExtIDs[3])
	chain.Metadata.Token = string(first.ExtIDs[1])
	chain.Metadata.Issuer = chain.Identity.ChainID
	chain.Metadata.Height = first.Height

	if err := chain.setupDB(); err != nil {
		return err
	}
	log.Debugf("Tracked: %v", chain)
	return nil
}
func (chain *Chain) issue(issuance fat.Issuance) error {
	chain.ChainStatus = ChainStatusIssued
	chain.Issuance = issuance

	if err := chain.saveIssuance(); err != nil {
		return err
	}
	log.Debugf("Issued: %v", chain)
	return nil
}
