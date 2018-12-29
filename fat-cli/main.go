package main

import (
	"fmt"
	"os"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat0"
)

func main() { os.Exit(_main()) }
func _main() (ret int) {
	Parse()
	// Attempt to run the completion program.
	if Completion.Complete() {
		// The completion program ran, so just return.
		return 0
	}
	if err := Validate(); err != nil {
		fmt.Println(err)
		return 1
	}

	switch cmd {
	case "issue":
		if err := issue(); err != nil {
			fmt.Println(err)
			return 1
		}
	}

	return 0
}
func issue() error {
	eb := factom.EBlock{ChainID: chainID}
	if err := eb.GetFirst(); err != nil {
		return err
	}
	if flagIsSet["chainid"] {
		if !eb.IsPopulated() {
			// The chain must already exist if the user specifies
			// -chainid.
			return fmt.Errorf("The specified chainid does not exist.\n" +
				"Use -token and -identity to attempt to create it.")
		}
		// Get NameIDs for chain to check if this chain is valid.
		first := eb.Entries[0]
		if err := first.Get(); err != nil {
			return err
		}
		if !first.IsPopulated() {
			return fmt.Errorf("Failed to populate Entry%+v", eb.Entries[0])
		}
		if !fat0.ValidTokenNameIDs(first.ExtIDs) {
			return fmt.Errorf("Not a valid token chain")
		}
		token = string(first.ExtIDs[1])
		copy(identity.ChainID[:], first.ExtIDs[1])
	} else if !eb.IsPopulated() {
		// Create the chain
		e := factom.Entry{ExtIDs: fat0.NameIDs(token, identity.ChainID)}
		txID, err := e.Create(ecpub)
		if err != nil {
			return err
		}
		if !e.IsPopulated() {
			return fmt.Errorf("Failed to create token chain")
		}
		fmt.Println("Created Token Chain")
		fmt.Println("First Entry Hash: ", e.Hash)
		fmt.Println("TxID: ", txID)
		fmt.Println("You must wait until the Token Chain is created " +
			"before attempting to issue the token. The longest " +
			"this can take is 10 minutes.")
		return nil
	}
	if err := identity.Get(); err != nil {
		return err
	}
	if !identity.IsPopulated() {
		return fmt.Errorf("Identity Chain does not exist")
	}
	if *identity.IDKey != *sk1.RCDHash() {
		return fmt.Errorf("Invalid SK1 key for Identity%+v", identity)
	}

	// Create issuance entry
	if err := issuance.MarshalEntry(); err != nil {
		return err
	}
	issuance.Sign(sk1)
	if err := issuance.Valid(*identity.IDKey); err != nil {
		return err
	}
	txID, err := issuance.Create(ecpub)
	if err != nil {
		return err
	}
	fmt.Println("Token Chain ID: ", chainID)
	fmt.Println("Created Issuance Entry")
	fmt.Println("Issuance Entry Hash: ", issuance.Hash)
	fmt.Println("TxID: ", txID)
	return nil
}
