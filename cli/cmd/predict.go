package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Factom-Asset-Tokens/fatd/srv"
	"github.com/posener/complete"
	flag "github.com/spf13/pflag"
)

func parseAPIFlags() error {
	args := strings.Fields(os.Getenv("COMP_LINE"))[1:]
	flags := flag.NewFlagSet("", flag.ContinueOnError)
	flags.ParseErrorsWhitelist.UnknownFlags = true
	flags.StringVarP(&FATClient.FatdServer, "fatd", "d",
		"http://localhost:8078", "")
	flags.StringVarP(&FactomClient.FactomdServer, "factomd", "s",
		"http://localhost:8088", "")
	flags.StringVarP(&FactomClient.WalletdServer, "walletd", "w",
		"http://localhost:8089", "")
	if err := flags.Parse(args); err != nil {
		return err
	}
	FATClient.Timeout = time.Second
	FactomClient.Factomd.Timeout = time.Second
	FactomClient.Walletd.Timeout = time.Second
	return nil
}

var PredictFAAddresses complete.PredictFunc = func(_ complete.Args) []string {
	if err := parseAPIFlags(); err != nil {
		return nil
	}
	adrs, err := FactomClient.GetFAAddresses()
	if err != nil {
		return nil
	}
	adrStrs := make([]string, len(adrs))
	for i, adr := range adrs {
		adrStrs[i] = adr.String()
	}
	return adrStrs
}

var PredictChainIDs complete.PredictFunc = func(_ complete.Args) []string {
	if err := parseAPIFlags(); err != nil {
		return nil
	}
	var chains []srv.ParamsToken
	if err := FATClient.Request("get-daemon-tokens", nil, &chains); err != nil {
		return nil
	}
	chainStrs := make([]string, len(chains))
	for i, chain := range chains {
		chainStrs[i] = chain.ChainID.String()
	}
	return chainStrs
}
