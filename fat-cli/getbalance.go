package main

import (
	"fmt"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/srv"
)

func getBalance() error {
	params := srv.ParamsGetBalance{
		ParamsToken: srv.ParamsToken{
			ChainID: chainID,
		},
		Address: address.RCDHash(),
	}
	var balance uint64
	err := factom.Request(APIAddress, "get-balance", params, &balance)
	if err != nil {
		return err
	}
	fmt.Println(balance)
	return nil
}
