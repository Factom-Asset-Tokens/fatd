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
	ids []*factom.Bytes32
	*sync.RWMutex
}

func (cm *ChainMap) set(id *factom.Bytes32, chain *Chain) {
	defer cm.Unlock()
	cm.Lock()
	if chain.IsIssued() {
		if _, ok := cm.m[*id]; !ok {
			cm.ids = append(cm.ids, id)
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

func (cm ChainMap) GetIssued() []*factom.Bytes32 {
	defer cm.RUnlock()
	cm.RLock()
	return cm.ids
}
