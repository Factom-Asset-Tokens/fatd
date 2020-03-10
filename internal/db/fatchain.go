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

package db

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Factom-Asset-Tokens/factom"
	"github.com/Factom-Asset-Tokens/factom/fat"
	"github.com/Factom-Asset-Tokens/fatd/internal/db/address"
	"github.com/Factom-Asset-Tokens/fatd/internal/db/entry"
	"github.com/Factom-Asset-Tokens/fatd/internal/db/metadata"
	_log "github.com/Factom-Asset-Tokens/fatd/internal/log"
)

type FATChain struct {
	// FAT Specific Data
	TokenID       string
	IssuerChainID *factom.Bytes32
	Identity      factom.Identity
	Issuance      fat.Issuance
	NumIssued     uint64

	// General Factom Blockchain Data
	FactomChain
}

func NewFATChain(ctx context.Context, dbPath, tokenID string,
	chainID, issuerChainID *factom.Bytes32,
	networkID factom.NetworkID) (_ FATChain, err error) {

	chain, err := NewFactomChain(ctx, dbPath, chainID, networkID)
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			chain.Close()
			if err := os.Remove(dbPath + chain.DBFile); err != nil &&
				!os.IsNotExist(err) {
				chain.Log.Errorf("os.Remove(): %w", err)
			}
		}
	}()

	if err = metadata.InsertFATChain(chain.Conn,
		tokenID, issuerChainID); err != nil {
		return
	}

	// Ensure that the coinbase address has rowid = 1.
	coinbase := fat.Coinbase()
	if _, err = address.Add(chain.Conn, &coinbase, 0); err != nil {
		return
	}

	return FATChain{
		FactomChain: chain,

		TokenID:       tokenID,
		IssuerChainID: issuerChainID,
		Identity: factom.Identity{
			Entry: factom.Entry{ChainID: issuerChainID}},
	}, nil
}

func OpenFATChain(ctx context.Context,
	dbPath, fname string) (chain FATChain, err error) {
	chain.Log = _log.New("chain", strings.TrimRight(fname, dbFileExtension))
	chain.Log.Info("Opening...")
	chain.Conn, chain.Pool, err = OpenConnPool(ctx, dbPath+fname)
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			chain.Close()
		}
	}()
	chain.DBFile = fname

	err = chain.load()
	return
}

func (chain *FATChain) load() error {
	if err := chain.FactomChain.load(); err != nil {
		return err
	}

	// Load NameIDs
	first, err := entry.SelectByID(chain.Conn, 1)
	if err != nil {
		return fmt.Errorf("entry.SelectByID(): %w", err)
	}
	if !first.IsPopulated() {
		return fmt.Errorf("no first entry")
	}

	nameIDs := first.ExtIDs
	if !fat.ValidNameIDs(nameIDs) {
		return fmt.Errorf("invalid FAT Token Chain Name IDs: %v", nameIDs)
	}
	tokenID, issuerChainID := fat.ParseTokenIssuer(nameIDs)
	chain.TokenID = tokenID
	chain.IssuerChainID = &issuerChainID
	chain.Identity.ChainID = &issuerChainID

	chain.NumIssued, chain.TokenID, chain.Identity,
		chain.Issuance, err = metadata.SelectFATChain(chain.Conn)

	return err
}

func (chain *FATChain) AddNumIssued(add uint64) error {
	if err := metadata.AddNumIssued(chain.Conn, add); err != nil {
		return err
	}
	chain.NumIssued += add
	return nil
}
