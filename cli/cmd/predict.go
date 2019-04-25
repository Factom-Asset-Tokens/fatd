package cmd

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/srv"
	"github.com/posener/complete"
)

var logErr = func(_ ...interface{}) {}

// parseAPIFlags parses
func parseAPIFlags() error {
	args := strings.Fields(os.Getenv("COMP_LINE"))[1:]
	if err := apiFlags.Parse(args); err != nil {
		return err
	}
	if DebugCompletion {
		log.SetOutput(os.Stderr)
		log.SetFlags(0)
		logErr = log.Println
	}
	FATClient.Timeout = time.Second / 3

	// Override --debug flag. --debugfactomd --debugfatd and --debugwalletd
	// may still be used explicitly but these are hidden and not part of
	// normal use.
	Debug = false
	initClients()
	return nil
}

var PredictFAAddresses complete.PredictFunc = func(args complete.Args) []string {
	if len(args.Last) > 52 {
		return nil
	}
	if err := parseAPIFlags(); err != nil {
		return nil
	}
	adrs, err := FactomClient.GetFAAddresses()
	if err != nil {
		logErr(err)
		return nil
	}
	completed := make(map[factom.FAAddress]struct{}, len(args.Completed)-1)
	for _, arg := range args.Completed[1:] {
		var adr factom.FAAddress
		if adr.Set(arg) != nil {
			continue
		}
		completed[adr] = struct{}{}
	}
	adrStrs := make([]string, len(adrs)-len(completed))
	var i int
	for _, adr := range adrs {
		if _, ok := completed[adr]; ok {
			continue
		}
		adrStrs[i] = adr.String()
		i++
	}
	return adrStrs
}

func PredictAppend(predict complete.PredictFunc, suffix string) complete.PredictFunc {
	return func(args complete.Args) []string {
		predictions := predict(args)
		for i := range predictions {
			predictions[i] += suffix
		}
		return predictions
	}
}

var PredictECAddresses complete.PredictFunc = func(args complete.Args) []string {
	if len(args.Last) > 52 {
		return nil
	}
	if err := parseAPIFlags(); err != nil {
		return nil
	}
	adrs, err := FactomClient.GetECAddresses()
	if err != nil {
		logErr(err)
		return nil
	}
	completed := make(map[factom.ECAddress]struct{}, len(args.Completed)-1)
	for _, arg := range args.Completed[1:] {
		var adr factom.ECAddress
		if adr.Set(arg) != nil {
			continue
		}
		completed[adr] = struct{}{}
	}
	adrStrs := make([]string, len(adrs)-len(completed))
	var i int
	for _, adr := range adrs {
		if _, ok := completed[adr]; ok {
			continue
		}
		adrStrs[i] = adr.String()
		i++
	}
	return adrStrs
}

var PredictChainIDs complete.PredictFunc = func(args complete.Args) []string {
	if len(args.Last) > 64 {
		return nil
	}
	if err := parseAPIFlags(); err != nil {
		return nil
	}
	var chains []srv.ParamsToken
	if err := FATClient.Request("get-daemon-tokens", nil, &chains); err != nil {
		logErr(err)
		return nil
	}
	completed := make(map[factom.Bytes32]struct{}, len(args.Completed)-1)
	for _, arg := range args.Completed[1:] {
		var chainID factom.Bytes32
		if chainID.Set(arg) != nil {
			continue
		}
		completed[chainID] = struct{}{}
	}
	chainStrs := make([]string, len(chains)-len(completed))
	var i int
	for _, chain := range chains {
		if _, ok := completed[*chain.ChainID]; ok {
			continue
		}
		chainStrs[i] = chain.ChainID.String()
		i++
	}
	return chainStrs
}
