package main

import (
	"fmt"

	jrpc "github.com/AdamSLevy/jsonrpc2/v10"
	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat0"
)

func issue() error {
	eb := factom.EBlock{ChainID: chainID}
	if err := eb.GetFirst(); err != nil {
		if _, ok := err.(jrpc.Error); !ok {
			return err
		}
	}
	if flagIsSet["chainid"] {
		if !eb.IsPopulated() {
			// The chain must already exist if the user specifies
			// -chainid.
			return fmt.Errorf("The specified chainid does not exist.\n" +
				"Use -tokenid and -identity to attempt to create it.")
		}
		// Get NameIDs for chain to check if this chain is valid.
		first := eb.Entries[0]
		if err := first.Get(); err != nil {
			return err
		}
		if !fat0.ValidTokenNameIDs(first.ExtIDs) {
			return fmt.Errorf("Not a valid token chain")
		}
		issuance.TokenID = string(first.ExtIDs[1])
		copy(identity.ChainID[:], first.ExtIDs[3])
	} else if !eb.IsPopulated() {
		// Create the chain
		e := factom.Entry{ExtIDs: fat0.NameIDs(issuance.TokenID, identity.ChainID)}
		txID, err := e.Create(ECPub)
		if err != nil {
			return err
		}
		fmt.Println("Created Token Chain")
		fmt.Println("Token Chain ID: ", e.ChainID)
		fmt.Println("First Entry Hash: ", e.Hash)
		fmt.Println("Factom TxID: ", txID)
		fmt.Println("You must wait until the Token Chain is created " +
			"before issuing the token. \nThis can take up to 10 minutes.")
		return nil
	}
	if err := identity.Get(); err != nil {
		return err
	}
	if *identity.IDKey != *sk1.RCDHash() {
		return fmt.Errorf("Invalid SK1 key for Identity%+v", identity)
	}

	// Create issuance entry
	if err := issuance.MarshalEntry(); err != nil {
		return err
	}
	issuance.Sign(sk1)
	if err := issuance.Valid(identity.IDKey); err != nil {
		return err
	}
	txID, err := issuance.Create(ECPub)
	if err != nil {
		return err
	}
	fmt.Println("Created Issuance Entry")
	fmt.Println("Token Chain ID: ", chainID)
	fmt.Println("Issuance Entry Hash: ", issuance.Hash)
	fmt.Println("Factom TxID: ", txID)
	return nil
}
