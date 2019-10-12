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

// +build ignore

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	. "github.com/Factom-Asset-Tokens/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat"
	"github.com/Factom-Asset-Tokens/fatd/internal/db"
	fflag "github.com/Factom-Asset-Tokens/fatd/internal/flag"
)

func init() {
	log.SetFlags(log.Lshortfile)
	fflag.DBPath = "./test-fatd.db/"
	fflag.LogDebug = true
}

func main() {
	if err := os.Mkdir(fflag.DBPath, 0755); err != nil {
		if !os.IsExist(err) {
			log.Fatalf("os.Mkdir(%#v): %v", fflag.DBPath, err)
		}
	}
	c := NewClient()
	chainID := NewBytes32FromString(
		"b54c4310530dc4dd361101644fa55cb10aec561e7874a7b786ea3b66f2c6fdfb")
	flag.Var(chainID, "chainid", "Chain ID to use for the test database")
	flag.StringVar(&c.FactomdServer, "factomd", c.FactomdServer, "factomd endpoint")
	flag.Parse()

	log.SetPrefix(fmt.Sprintf("ChainID: %v ", chainID.String()))

	eblocks, err := EBlock{ChainID: chainID}.GetPrevAll(context.Background(), c)
	if err != nil {
		log.Fatal(err)
	}

	first := eblocks[len(eblocks)-1]
	var dblock DBlock
	dblock.Height = first.Height
	if err := dblock.Get(context.Background(), c); err != nil {
		log.Fatal(err)
	}
	timestamp := dblock.Timestamp
	for i := range first.Entries {
		e := &first.Entries[i]
		if err := e.Get(context.Background(), c); err != nil {
			log.Fatal(err)
		}
		e.Timestamp = timestamp.Add(e.Timestamp.Sub(first.Timestamp))
	}
	first.Timestamp = timestamp

	nameIDs := first.Entries[0].ExtIDs

	if !fat.ValidTokenNameIDs(nameIDs) {
		log.Fatalf("invalid token chain")
	}
	_, identityChainID := fat.TokenIssuer(nameIDs)
	identity := NewIdentity(&identityChainID)
	if err := identity.Get(context.Background(), c); err != nil {
		log.Fatal(err)
	}

	// We don't need the actual dbKeyMR
	chain, err := db.OpenNew(context.Background(),
		fflag.DBPath, dblock.KeyMR, first, MainnetID(), identity)
	if err != nil {
		log.Println(err)
		return
	}
	defer chain.Close()

	eblocks = eblocks[:len(eblocks)-1] // skip first eblock
	for i := range eblocks {
		eb := eblocks[len(eblocks)-i-1]
		if err := eb.GetEntries(context.Background(), c); err != nil {
			log.Fatal(err)
		}
		var dblock DBlock
		dblock.Height = eb.Height
		if err := dblock.Get(context.Background(), c); err != nil {
			log.Fatal(err)
		}
		eb.SetTimestamp(dblock.Timestamp)

		if err := chain.Apply(dblock.KeyMR, eb); err != nil {
			log.Fatal(err)
		}
	}
}
