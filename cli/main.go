package main

import (
	"fmt"
	"os"
)

var Revision string

func main() {
	if err := _main(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
func _main() error {
	Parse()
	// Attempt to run the completion program.
	if Completion.Complete() {
		// The completion program ran, so just return.
		return nil
	}
	if err := Validate(); err != nil {
		return err
	}

	var cmdFunc func() error
	var ok bool
	if cmdFunc, ok = cmdFuncMap[SubCommand]; !ok {
		cmdFunc = usage
	}
	if err := cmdFunc(); err != nil {
		return err
	}

	return nil
}

func usage() error {
	fmt.Println(`usage: fat-cli CHAIN_FLAGS [GLOBAL_FLAGS] COMMAND COMMAND_FLAGS
        CHAIN_FLAGS: -chainid OR -token AND -identity
        GLOBAL_FLAGS: -s, -w, -apiaddress, ...
        COMMAND: balance OR issue OR transact`)
	return nil
}

var cmdFuncMap = map[string]func() error{
	"issue":          issue,
	"transactFAT0":   transactFAT0,
	"transactFAT1":   transactFAT1,
	"balance":        getBalance,
	"getissuance":    getIssuance,
	"getstats":       getStats,
	"listtokens":     listTokens,
	"gettransaction": getTransaction,
	"usage":          usage,
	"version":        version,
}
