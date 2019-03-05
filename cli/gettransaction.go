package main

import (
	"fmt"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat"
	"github.com/Factom-Asset-Tokens/fatd/fat/fat0"
	"github.com/Factom-Asset-Tokens/fatd/fat/fat1"
	"github.com/Factom-Asset-Tokens/fatd/srv"
)

func getTransaction() error {
	params := srv.ParamsGetTransaction{
		ParamsToken: srv.ParamsToken{
			ChainID: chainID,
		},
		Hash: txHash,
	}
	stats := struct {
		Type fat.Type `json:"type"`
	}{}
	if err := factom.Request(APIAddress, "get-stats",
		params.ParamsToken, &stats); err != nil {
		return err
	}
	result := srv.ResultGetTransaction{}
	switch stats.Type {
	case fat0.Type:
		result.Tx = &FAT0transaction
	case fat1.Type:
		result.Tx = &FAT1transaction
	default:
		panic(fmt.Sprintf("unknown FAT type: %v", stats.Type))
	}
	if err := factom.Request(APIAddress, "get-transaction", params, &result); err != nil {
		return err
	}
	fmt.Printf("Transaction: \n")
	fmt.Printf("\tHash: %v\n", result.Hash)
	fmt.Printf("\tTimestamp: %v\n", result.Timestamp)
	fmt.Printf("\tInputs: \n")
	switch result.Tx.(type) {
	case *fat0.Transaction:
		for rcdHash, amount := range FAT0transaction.Inputs {
			if FAT0transaction.IsCoinbase() {
				fmt.Printf("\t\tCoinbase: %v\n", amount)
				break
			}
			adr := factom.NewAddress(&rcdHash)
			fmt.Printf("\t\t%v: %v\n", adr, amount)
		}
		fmt.Printf("\tOutputs: \n")
		for rcdHash, amount := range FAT0transaction.Outputs {
			adr := factom.NewAddress(&rcdHash)
			fmt.Printf("\t\t%v: %v\n", adr, amount)
		}
		if len(FAT0transaction.Metadata) > 0 {
			fmt.Printf("\tMetadata: %v\n", FAT0transaction.Metadata)
		}
	case *fat1.Transaction:
		for rcdHash, amount := range FAT1transaction.Inputs {
			if FAT1transaction.IsCoinbase() {
				fmt.Printf("\t\tCoinbase: %v\n", amount)
				break
			}
			adr := factom.NewAddress(&rcdHash)
			fmt.Printf("\t\t%v: %v\n", adr, amount)
		}
		fmt.Printf("\tOutputs: \n")
		for rcdHash, amount := range FAT1transaction.Outputs {
			adr := factom.NewAddress(&rcdHash)
			fmt.Printf("\t\t%v: %v\n", adr, amount)
		}
		if len(FAT1transaction.Metadata) > 0 {
			fmt.Printf("\tMetadata: %v\n", FAT1transaction.Metadata)
		}
	}
	fmt.Printf("\n")
	return nil
}
