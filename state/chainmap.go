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

package state

import (
	"sync"

	"github.com/Factom-Asset-Tokens/fatd/factom"
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

func (cm *ChainMap) set(id *factom.Bytes32, chain *Chain) {
	defer cm.Unlock()
	cm.Lock()
	if chain.IsIssued() {
		if chain, ok := cm.m[*id]; !ok || !chain.IsIssued() {
			cm.ids = append(cm.ids, *id)
		}
	}
	cm.m[*id] = *chain
}

func (cm ChainMap) Get(id *factom.Bytes32) Chain {
	defer cm.RUnlock()
	cm.RLock()
	chain, ok := cm.m[*id]
	if !ok {
		chain.ID = id
	}
	return chain
}

func (cm ChainMap) GetIssued() []factom.Bytes32 {
	defer cm.RUnlock()
	cm.RLock()
	return cm.ids
}
