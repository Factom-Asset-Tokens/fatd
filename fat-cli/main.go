package main

import (
	"fmt"
	"os"
)

func main() { os.Exit(_main()) }
func _main() (ret int) {
	Parse()
	// Attempt to run the completion program.
	if Completion.Complete() {
		// The completion program ran, so just return.
		return 0
	}
	if err := Validate(); err != nil {
		fmt.Println(err)
		return 1
	}

	switch SubCommand {
	case "issue":
		if err := issue(); err != nil {
			fmt.Println(err)
			return 1
		}
	case "transact":
		if err := transact(); err != nil {
			fmt.Println(err)
			return 1
		}
	case "balance":
		if err := getBalance(); err != nil {
			fmt.Println(err)
			return 1
		}
	case "getissuance":
		if err := getIssuance(); err != nil {
			fmt.Println(err)
			return 1
		}
	case "getstats":
		if err := getStats(); err != nil {
			fmt.Println(err)
			return 1
		}
	case "listtokens":
		if err := listTokens(); err != nil {
			fmt.Println(err)
			return 1
		}
	case "gettransaction":
		if err := getTransaction(); err != nil {
			fmt.Println(err)
			return 1
		}
	default:
		usage()
	}

	return 0
}

func usage() {
	fmt.Println(`usage: fat-cli CHAIN_FLAGS [GLOBAL_FLAGS] COMMAND COMMAND_FLAGS
        CHAIN_FLAGS: -chainid OR -token AND -identity
        GLOBAL_FLAGS: -s, -w, -apiaddress, ...
        COMMAND: balance OR issue OR transact`)
}
