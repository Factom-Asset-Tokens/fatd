package main

import (
	"fmt"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/srv"
)

func listTokens() error {
	var tkns []srv.ParamsToken
	err := factom.Request(APIAddress, "get-daemon-tokens", nil, &tkns)
	if err != nil {
		return err
	}
	for _, tkn := range tkns {
		fmt.Printf("Chain ID: %v\n", tkn.ChainID)
		fmt.Printf("Token ID: %v\n", tkn.TokenID)
		fmt.Printf("Issuer Identity Chain ID: %v\n\n", tkn.IssuerChainID)
	}
	return nil
}
