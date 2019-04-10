package main

import (
	"fmt"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/srv"
)

func getStats() error {
	params := srv.ParamsToken{ChainID: chainID}
	var stats srv.ResultGetStats
	err := factom.Request(APIAddress, "get-stats", params, &stats)
	if err != nil {
		return err
	}
	fmt.Printf("Supply: %v\n", stats.Supply)
	fmt.Printf("Circulating Supply: %v\n", stats.CirculatingSupply)
	fmt.Printf("Burned: %v\n", stats.Burned)
	fmt.Printf("Number of Transactions: %v\n", stats.Transactions)
	fmt.Printf("Time of Issuance: %v\n", stats.IssuanceTimestamp.Time)
	fmt.Printf("Time of Latest Transaction: %v\n", stats.LastTransactionTimestamp.Time)
	return nil
}
