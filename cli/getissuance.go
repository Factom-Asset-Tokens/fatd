package main

import (
	"fmt"

	"github.com/Factom-Asset-Tokens/fatd/srv"
)

func getIssuance() error {
	params := srv.ParamsToken{ChainID: chainID}
	var issuance srv.ResultGetIssuance
	err := FactomClient.Request(APIAddress, "get-issuance", params, &issuance)
	if err != nil {
		return err
	}
	fmt.Printf("Chain ID: %v\n", issuance.ChainID)
	fmt.Printf("Token ID: %v\n", issuance.TokenID)
	fmt.Printf("Issuer Identity Chain ID: %v\n", issuance.IssuerChainID)
	fmt.Printf("Time of Issuance: %v\n", issuance.Timestamp.Time)
	fmt.Printf("Issuance:\n")
	fmt.Printf("\tType: %v\n", issuance.Issuance.Type)
	fmt.Printf("\tSupply: %v\n", issuance.Issuance.Supply)
	fmt.Printf("\tSymbol: %v\n", issuance.Issuance.Symbol)
	fmt.Printf("\tMetadata: %v\n", issuance.Issuance.Metadata)
	return nil
}
