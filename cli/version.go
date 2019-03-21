package main

import (
	"fmt"

	"github.com/Factom-Asset-Tokens/fatd/srv"
)

func version() error {
	fmt.Printf("fat-cli: %v\n", Revision)
	var properties srv.ResultGetDaemonProperties
	err := FactomClient.Request(APIAddress, "get-daemon-properties", nil, &properties)
	if err != nil {
		return err
	}
	fmt.Printf("fatd:    %v\n", properties.FatdVersion)
	fmt.Printf("API:     %v\n", properties.APIVersion)
	return nil
}
