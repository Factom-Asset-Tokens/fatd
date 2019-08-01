// +build ignore

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/Factom-Asset-Tokens/fatd/db"
	. "github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat"
	fflag "github.com/Factom-Asset-Tokens/fatd/flag"
)

func init() {
	log.SetFlags(log.Lshortfile)
	fflag.DBPath = "./test-fatd.db"
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

	eblocks, err := EBlock{ChainID: chainID}.GetAllPrev(c)
	if err != nil {
		log.Fatal(err)
	}

	first := eblocks[len(eblocks)-1]
	var dblock DBlock
	dblock.Header.Height = first.Height
	if err := dblock.Get(c); err != nil {
		log.Fatal(err)
	}
	timestamp := dblock.Header.Timestamp
	for i := range first.Entries {
		e := &first.Entries[i]
		if err := e.Get(c); err != nil {
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
	identity := NewIdentity(identityChainID)
	if err := identity.Get(c); err != nil {
		log.Fatal(err)
	}

	// We don't need the actual dbKeyMR
	chain, err := db.OpenNew(first, dblock.KeyMR, Mainnet(), identity)
	if err != nil {
		log.Println(err)
		return
	}
	defer chain.Close()

	eblocks = eblocks[:len(eblocks)-1] // skip first eblock
	for i := range eblocks {
		eb := eblocks[len(eblocks)-i-1]
		var dblock DBlock
		dblock.Header.Height = eb.Height
		if err := dblock.Get(c); err != nil {
			log.Fatal(err)
		}
		timestamp := dblock.Header.Timestamp
		for i := range eb.Entries {
			e := &eb.Entries[i]
			if err := e.Get(c); err != nil {
				log.Fatal(err)
			}
			e.Timestamp = timestamp.Add(e.Timestamp.Sub(eb.Timestamp))
		}
		eb.Timestamp = timestamp
		if err := chain.Apply(eb, dblock.KeyMR); err != nil {
			log.Fatal(err)
		}
	}
}
