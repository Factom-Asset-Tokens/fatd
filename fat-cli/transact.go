package main

import (
	"fmt"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat0"
)

func transact() error {
	inputAddresses := make([]factom.Address, 0, len(transaction.Inputs))
	if flagIsSet["coinbase"] {
		eb := factom.EBlock{ChainID: chainID}
		if err := eb.GetFirst(); err != nil {
			return err
		}
		if !eb.IsPopulated() {
			return fmt.Errorf("Token Chain not found")
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
		copy(identity.ChainID[:], first.ExtIDs[3])
		if err := identity.Get(); err != nil {
			return err
		}
		if !identity.IsPopulated() {
			return fmt.Errorf("Identity Chain does not exist")
		}
		if *identity.IDKey != *sk1.RCDHash() {
			return fmt.Errorf("Invalid SK1 key for Identity%+v", identity)
		}
		inputAddresses = append(inputAddresses, sk1)
	} else {
		for rcd := range transaction.Inputs {
			adr := factom.NewAddress(&rcd)
			if err := adr.Get(); err != nil {
				return err
			}
			inputAddresses = append(inputAddresses, adr)
		}
	}
	if err := transaction.MarshalEntry(); err != nil {
		return err
	}
	transaction.Sign(inputAddresses...)
	if err := transaction.Valid(sk1.RCDHash()); err != nil {
		return err
	}
	txID, err := transaction.Create(ecpub)
	if err != nil {
		return err
	}
	fmt.Println("Token Chain ID: ", chainID)
	fmt.Println("Created Transaction Entry")
	fmt.Println("Transaction Entry Hash: ", transaction.Hash)
	fmt.Println("Factom TxID: ", txID)
	return nil
}
