package main

import (
	"fmt"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat"
	fctm "github.com/Factom-Asset-Tokens/fatd/fctm"
)

func transactFAT0() error {
	signingAddresses := make([]factom.RCDPrivateKey, 0, len(FAT0transaction.Inputs))
	if flagMap["coinbase"].IsSet {
		eb := factom.EBlock{ChainID: chainID}
		if err := eb.GetFirst(FactomClient); err != nil {
			return err
		}
		if !eb.IsPopulated() {
			return fmt.Errorf("Token Chain not found")
		}
		// Get NameIDs for chain to check if this chain is valid.
		first := eb.Entries[0]
		if err := first.Get(FactomClient); err != nil {
			return err
		}
		if !first.IsPopulated() {
			return fmt.Errorf("Failed to populate Entry%+v", eb.Entries[0])
		}
		if !fat.ValidTokenNameIDs(first.ExtIDs) {
			return fmt.Errorf("Not a valid token chain")
		}
		copy(identity.ChainID[:], first.ExtIDs[3])
		if err := identity.Get(FactomClient); err != nil {
			return err
		}
		if !identity.IsPopulated() {
			return fmt.Errorf("Identity Chain does not exist")
		}
		if identity.ID1 != sk1.ID1Key() {
			return fmt.Errorf("Invalid SK1 key for Identity%+v", identity)
		}
		signingAddresses = append(signingAddresses, sk1)
	} else {
		for fa := range FAT0transaction.Inputs {
			fs, err := fa.GetFsAddress(FactomClient)
			if err != nil {
				return err
			}
			signingAddresses = append(signingAddresses, fs)
		}
	}
	if err := FAT0transaction.MarshalEntry(); err != nil {
		return err
	}
	FAT0transaction.Sign(signingAddresses...)
	if err := FAT0transaction.Valid(&sk1); err != nil {
		return err
	}
	var txID *factom.Bytes32
	var err error
	if len(ecpub) != 0 {
		txID, err = FAT0transaction.Create(FactomClient, ecpub)
		if err != nil {
			return err
		}
		Hash = e.Hash
	} else {
		FAT0transaction.Timestamp = factom.Time{}
		result := struct {
			*factom.Entry
			TxID *factom.Bytes32 `json:"txid"`
		}{Entry: &FAT0transaction.Entry.Entry}
		err := FactomClient.Factomd.Request(APIAddress, "send-transaction",
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
	signingAddresses := make([]factom.RCDPrivateKey, 0, len(FAT1transaction.Inputs))
	if flagMap["coinbase"].IsSet {
		eb := factom.EBlock{ChainID: chainID}
		if err := eb.GetFirst(FactomClient); err != nil {
			return err
		}
		if !eb.IsPopulated() {
			return fmt.Errorf("Token Chain not found")
		}
		// Get NameIDs for chain to check if this chain is valid.
		first := eb.Entries[0]
		if err := first.Get(FactomClient); err != nil {
			return err
		}
		if !first.IsPopulated() {
			return fmt.Errorf("Failed to populate Entry%+v", eb.Entries[0])
		}
		if !fat.ValidTokenNameIDs(first.ExtIDs) {
			return fmt.Errorf("Not a valid token chain")
		}
		copy(identity.ChainID[:], first.ExtIDs[3])
		if err := identity.Get(FactomClient); err != nil {
			return err
		}
		if !identity.IsPopulated() {
			return fmt.Errorf("Identity Chain does not exist")
		}
		if identity.ID1 != sk1.ID1Key() {
			return fmt.Errorf("Invalid SK1 key for Identity%+v", identity)
		}
		signingAddresses = append(signingAddresses, sk1)
	} else {
		for fa := range FAT1transaction.Inputs {
			fs, err := fa.GetFsAddress(FactomClient)
			if err != nil {
				return err
			}
			signingAddresses = append(signingAddresses, fs)
		}
	}
	if err := FAT1transaction.MarshalEntry(); err != nil {
		return err
	}
	FAT1transaction.Sign(signingAddresses...)
	if err := FAT1transaction.Valid(&sk1); err != nil {
		return err
	}
	var txID *factom.Bytes32
	var err error
	if len(ecpub) != 0 {
		txID, err = FAT1transaction.Create(FactomClient, ecpub)
		if err != nil {
			return err
		}
	} else {
		FAT1transaction.Timestamp = factom.Time{}
		result := struct {
			*factom.Entry
			TxID *factom.Bytes32 `json:"txid"`
		}{Entry: &FAT1transaction.Entry.Entry}
		err := FactomClient.Factomd.Request(APIAddress, "send-transaction",
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
