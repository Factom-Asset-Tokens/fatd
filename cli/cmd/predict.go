// MIT License
//
// Copyright 2018 Canonical Ledgers, LLC
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
// IN THE SOFTWARE.

package cmd

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Factom-Asset-Tokens/factom"
	"github.com/Factom-Asset-Tokens/fatd/srv"
	"github.com/posener/complete"
)

var logErr = func(_ ...interface{}) {}

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

func PredictAppend(predict complete.PredictFunc, suffix string) complete.PredictFunc {
	return func(args complete.Args) []string {
		predictions := predict(args)
		for i := range predictions {
			predictions[i] += suffix
		}
		return predictions
	}
}

func PredictAddressesFunc(fa bool) complete.PredictFunc {
	return func(args complete.Args) []string {
		// Check if the argument we are completing already exceeds the
		// length of an address.
		if len(args.Last) > 52 {
			return nil
		}

		if err := parseAPIFlags(); err != nil {
			return nil
		}

		fss, ess, err := FactomClient.GetPrivateAddresses(context.Background())
		if err != nil {
			logErr(err)
			return nil
		}

		// Return only the public addresses that we need.
		var adrStrs []string
		if fa {
			adrStrs = make([]string, len(fss))
			for i, fs := range fss {
				adrStrs[i] = fs.FAAddress().String()
			}
		} else {
			adrStrs = make([]string, len(ess))
			for i, es := range ess {
				adrStrs[i] = es.ECAddress().String()
			}
		}
		return adrStrs
	}

}

var PredictFAAddresses = PredictAddressesFunc(true)
var PredictECAddresses = PredictAddressesFunc(false)

var PredictChainIDs complete.PredictFunc = func(args complete.Args) []string {
	if len(args.Last) > 64 {
		return nil
	}
	if err := parseAPIFlags(); err != nil {
		return nil
	}
	var chains []srv.ParamsToken
	if err := FATClient.Request(context.Background(),
		"get-daemon-tokens", nil, &chains); err != nil {
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
