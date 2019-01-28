package main

import (
	"fmt"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/srv"
)

func getTransaction() error {
	params := srv.ParamsGetTransaction{
		ParamsToken: srv.ParamsToken{
			ChainID: chainID,
		},
		Hash: txHash,
	}
	result := srv.ResultGetTransaction{}
	err := factom.Request(APIAddress, "get-transaction", params, &result)
	if err != nil {
		return err
	}
	FAT0transaction = result.Tx
	FAT0transaction.Hash = result.Hash
	FAT0transaction.Timestamp = result.Timestamp
	fmt.Printf("Transaction: \n")
	fmt.Printf("\tHash: %v\n", FAT0transaction.Hash)
	fmt.Printf("\tTimestamp: %v\n", FAT0transaction.Timestamp.Time)
	fmt.Printf("\tInputs: \n")
	for rcdHash, amount := range FAT0transaction.Inputs {
		if FAT0transaction.IsCoinbase() {
			fmt.Printf("\t\tCoinbase: %v\n", amount)
			break
		}
		adr := factom.NewAddress(&rcdHash)
		fmt.Printf("\t\t%v: %v\n", adr, amount)
	}
	fmt.Printf("\tOutputs: \n")
	for rcdHash, amount := range FAT0transaction.Inputs {
		adr := factom.NewAddress(&rcdHash)
		fmt.Printf("\t\t%v: %v\n", adr, amount)
	}
	fmt.Printf("\tMetadata: %v\n", FAT0transaction.Metadata)
	fmt.Printf("\n")
	return nil
}
