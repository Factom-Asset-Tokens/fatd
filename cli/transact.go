package main

import (
	"fmt"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat"
	fctm "github.com/Factom-Asset-Tokens/fatd/fctm"
)

func transactFAT0() error {
	signingAddresses := make([]factom.Address, 0, len(FAT0transaction.Inputs))
	if flagMap["coinbase"].IsSet {
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
		if !fat.ValidTokenNameIDs(first.ExtIDs) {
			return fmt.Errorf("Not a valid token chain")
		}
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
		signingAddresses = append(signingAddresses, sk1)
	} else {
		for i := range allAddresses {
			adr := allAddresses[i]
			if _, ok := FAT0transaction.Inputs[*adr.RCDHash()]; !ok {
				continue
			}
			if err := adr.Get(); err != nil {
				return err
			}
			signingAddresses = append(signingAddresses, adr)
		}
	}
	if err := FAT0transaction.MarshalEntry(); err != nil {
		return err
	}
	FAT0transaction.Sign(signingAddresses...)
	if err := FAT0transaction.Valid(sk1.RCDHash()); err != nil {
		return err
	}
	var txID *fctm.Bytes32
	var Hash *fctm.Bytes32
	zeroEC := fctm.ECAddress{}
	zeroEs := fctm.EsAddress{}
	if ecadr != zeroEC || esadr != zeroEs {
		e := fctmEntry(FAT0transaction.Entry.Entry)
		zero := fctm.EsAddress{}
		var err error
		if esadr != zero {
			txID, err = e.ComposeCreate(FactomClient, esadr)
			if err != nil {
				return err
			}
		} else {
			txID, err = e.Create(FactomClient, ecadr)
			if err != nil {
				return err
			}
		}
		Hash = e.Hash
	} else {
		FAT0transaction.Timestamp = nil
		result := struct {
			*factom.Entry
			TxID *factom.Bytes32 `json:"txid"`
		}{Entry: &FAT0transaction.Entry.Entry}
		err := factom.Request(APIAddress, "send-transaction",
			FAT0transaction.Entry.Entry, &result)
		if err != nil {
			return err
		}
		txID = (*fctm.Bytes32)(result.TxID)
		Hash = (*fctm.Bytes32)(FAT0transaction.Hash)
	}

	fmt.Println("Created Transaction Entry")
	fmt.Println("Token Chain ID: ", chainID)
	fmt.Println("Transaction Entry Hash: ", Hash)
	fmt.Println("Factom TxID: ", txID)
	return nil
}

func transactFAT1() error {
	signingAddresses := make([]factom.Address, 0, len(FAT1transaction.Inputs))
	if flagMap["coinbase"].IsSet {
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
		if !fat.ValidTokenNameIDs(first.ExtIDs) {
			return fmt.Errorf("Not a valid token chain")
		}
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
		signingAddresses = append(signingAddresses, sk1)
	} else {
		for i := range allAddresses {
			adr := allAddresses[i]
			if _, ok := FAT1transaction.Inputs[*adr.RCDHash()]; !ok {
				continue
			}
			if err := adr.Get(); err != nil {
				return err
			}
			signingAddresses = append(signingAddresses, adr)
		}
	}
	if err := FAT1transaction.MarshalEntry(); err != nil {
		return err
	}
	FAT1transaction.Sign(signingAddresses...)
	if err := FAT1transaction.Valid(sk1.RCDHash()); err != nil {
		return err
	}
	var txID *fctm.Bytes32
	zeroEC := fctm.ECAddress{}
	zeroEs := fctm.EsAddress{}
	if ecadr != zeroEC || esadr != zeroEs {
		e := fctmEntry(FAT1transaction.Entry.Entry)
		zero := fctm.EsAddress{}
		var err error
		if esadr != zero {
			txID, err = e.ComposeCreate(FactomClient, esadr)
			if err != nil {
				return err
			}

		} else {
			txID, err = e.Create(FactomClient, ecadr)
			if err != nil {
				return err
			}
		}
	} else {
		FAT1transaction.Timestamp = nil
		result := struct {
			*factom.Entry
			TxID *factom.Bytes32 `json:"txid"`
		}{Entry: &FAT1transaction.Entry.Entry}
		err := factom.Request(APIAddress, "send-transaction",
			FAT1transaction.Entry.Entry, &result)
		if err != nil {
			return err
		}
		txID = (*fctm.Bytes32)(result.TxID)
	}

	fmt.Println("Created Transaction Entry")
	fmt.Println("Token Chain ID: ", chainID)
	fmt.Println("Transaction Entry Hash: ", FAT1transaction.Hash)
	fmt.Println("Factom TxID: ", txID)
	return nil
}
