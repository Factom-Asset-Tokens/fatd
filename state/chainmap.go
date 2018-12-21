package state

import (
	"sync"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat0"
)

var (
	chains = chainMap{m: map[factom.Bytes32]Chain{
		factom.Bytes32{31: 0x0a}: Chain{ChainStatus: ChainStatusIgnored},
		factom.Bytes32{31: 0x0c}: Chain{ChainStatus: ChainStatusIgnored},
		factom.Bytes32{31: 0x0f}: Chain{ChainStatus: ChainStatusIgnored},
	}, RWMutex: &sync.RWMutex{}}
)

type chainMap struct {
	m map[factom.Bytes32]Chain
	*sync.RWMutex
}

func (cm chainMap) Set(id *factom.Bytes32, chain Chain) {
	defer cm.Unlock()
	cm.Lock()
	cm.m[*id] = chain
}

func (cm chainMap) Get(id *factom.Bytes32) Chain {
	defer cm.RUnlock()
	cm.RLock()
	return cm.m[*id]
}

func (cm chainMap) Ignore(id *factom.Bytes32) {
	cm.Set(id, Chain{ChainStatus: ChainStatusIgnored})
}
func (cm chainMap) Track(id *factom.Bytes32, nameIDs []factom.Bytes) (chain Chain, err error) {
	token := string(nameIDs[1])
	identityChainID := factom.NewBytes32(nameIDs[3])
	chain = Chain{
		ID:          id,
		metadata:    metadata{Token: token, Issuer: identityChainID},
		ChainStatus: ChainStatusTracked,
		Identity:    fat0.Identity{ChainID: identityChainID},
	}
	if err = chain.setupDB(); err != nil {
		return
	}
	cm.Set(id, chain)
	return
}
func (cm chainMap) Issue(id *factom.Bytes32, issuance fat0.Issuance) (chain Chain, err error) {
	chain = cm.Get(id)
	chain.ChainStatus = ChainStatusIssued
	chain.Issuance = issuance
	cm.Set(id, chain)
	return
}
