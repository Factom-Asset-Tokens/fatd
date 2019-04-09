package main

import (
	"fmt"

	"github.com/Factom-Asset-Tokens/fatd/srv"
)

func getBalance() error {
	params := srv.ParamsGetBalance{
		ParamsToken: srv.ParamsToken{
			ChainID: chainID,
		},
		Address: &address,
	}
	var balance uint64
	err := FactomClient.Factomd.Request(APIAddress, "get-balance", params, &balance)
	if err != nil {
		return err
	}
	fmt.Println(balance)
	return nil
}
