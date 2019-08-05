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

package engine

import (
	"fmt"
	"math"
	"sync"

	"github.com/Factom-Asset-Tokens/fatd/db"
	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/flag"
)

var (
	Chains = ChainMap{m: map[factom.Bytes32]Chain{
		factom.Bytes32{31: 0x0a}: Chain{ChainStatus: ChainStatusIgnored},
		factom.Bytes32{31: 0x0c}: Chain{ChainStatus: ChainStatusIgnored},
		factom.Bytes32{31: 0x0f}: Chain{ChainStatus: ChainStatusIgnored},
	}, RWMutex: &sync.RWMutex{}}
)

type ChainMap struct {
	m   map[factom.Bytes32]Chain
	ids []factom.Bytes32
	*sync.RWMutex
}

func (cm ChainMap) set(id *factom.Bytes32, chain Chain) {
	defer cm.Unlock()
	cm.Lock()
	cm.m[*id] = chain
}

func (cm ChainMap) ignore(id *factom.Bytes32) {
	cm.set(id, Chain{ChainStatus: ChainStatusIgnored})
}

func (cm ChainMap) Get(id *factom.Bytes32) Chain {
	defer cm.RUnlock()
	cm.RLock()
	return cm.m[*id]
}

func (cm ChainMap) GetIssued() []factom.Bytes32 {
	defer cm.RUnlock()
	cm.RLock()
	return cm.ids
}

func (cm ChainMap) setSync(height uint32, dbKeyMR *factom.Bytes32) error {
	defer cm.Unlock()
	cm.Lock()
	for _, chain := range cm.m {
		if chain.Chain != nil {
			if err := chain.SetSync(height, dbKeyMR); err != nil {
				chain.Log.Errorf("chain.SetSync(): %v", err)
				return err
			}
		}
		cm.m[*chain.ID] = chain
	}
	return nil
}

func (cm ChainMap) Close() {
	defer cm.Unlock()
	cm.Lock()
	for _, chain := range cm.m {
		if chain.Chain != nil {
			chain.Close()
		}
	}
}

func loadChains() (syncHeight uint32, err error) {
	dbChains, err := db.OpenAll()
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			for _, chain := range dbChains {
				chain.Close()
			}
			Chains.Close()
		}
	}()

	// Set whitelisted chains to Tracked.
	for _, chainID := range flag.Whitelist {
		Chains.m[chainID] = Chain{ChainStatus: ChainStatusTracked}
	}
	// Blacklist overrides whitelist. Set chains to Ignore.
	for _, chainID := range flag.Blacklist {
		Chains.m[chainID] = Chain{ChainStatus: ChainStatusIgnored}
	}

	syncHeight = math.MaxUint32
	for i, dbChain := range dbChains {
		chain := Chains.m[*dbChain.ID]

		// Skip blacklisted chains or if there was a whitelist, any
		// non-tracked chain.
		if chain.IsIgnored() || flag.HasWhitelist() && !chain.IsTracked() {
			dbChain.Close()
			continue
		}

		chain.Chain = dbChain
		chain.ChainStatus = ChainStatusTracked
		syncHeight = min(syncHeight, chain.SyncHeight)

		if chain.NetworkID != flag.NetworkID {
			dbChains = dbChains[i:] // Close remaining chains.
			err = fmt.Errorf("invalid NetworkID: %v for Chain{%v}",
				chain.NetworkID, chain.ID)
			return
		}

		if err = chain.Sync(c); err != nil {
			dbChains = dbChains[i:] // Close remaining chains.
			return
		}
		if !flag.SkipDBValidation {
			if err = chain.Validate(); err != nil {
				dbChains = dbChains[i:] // Close remaining chains.
				return
			}
		}

		Chains.m[*dbChain.ID] = chain
	}
	dbChains = nil // Prevent closing any chains from this list.
	return
}
func min(a, b uint32) uint32 {
	if a <= b {
		return a
	}
	return b
}
