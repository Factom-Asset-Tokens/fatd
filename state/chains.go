package state

import (
	"sync"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat0"
)

type ChainStatus int

const (
	ChainStatusUnknown ChainStatus = -1
	ChainStatusIgnored ChainStatus = 0
	ChainStatusTracked ChainStatus = 1
	ChainStatusIssued  ChainStatus = 3 // Will also test true for Tracked()
)

func (status ChainStatus) Unknown() bool {
	return status == ChainStatusUnknown
}
func (status ChainStatus) Ignored() bool {
	return status == ChainStatusIgnored
}
func (status ChainStatus) Tracked() bool {
	return status|ChainStatusTracked == ChainStatusTracked
}
func (status ChainStatus) Issued() bool {
	return status|ChainStatusIssued == ChainStatusIssued
}

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

type Chain struct {
	ChainStatus
	fat0.State
	fat0.Identity
}

func NewChain(status ChainStatus) Chain {
	return Chain{ChainStatus: status}
}

func (cm chainMap) Ignore(c *factom.Bytes32) {
	cm.Set(c, NewChain(ChainStatusIgnored))
}
func (cm chainMap) Track(c *factom.Bytes32, identity fat0.Identity) {
	chain := NewChain(ChainStatusTracked)
	chain.Identity = identity
	cm.Set(c, chain)
}
func (cm chainMap) Issue(c *factom.Bytes32, issuance fat0.Issuance) {
	chain := cm.Get(c)
	chain.ChainStatus = ChainStatusIssued
	chain.State = fat0.NewState(issuance)
	cm.Set(c, chain)
}

func (cm chainMap) Set(c *factom.Bytes32, chain Chain) {
	defer cm.Unlock()
	cm.Lock()
	cm.m[*c] = chain
}

func (cm chainMap) Get(c *factom.Bytes32) Chain {
	defer cm.RUnlock()
	cm.RLock()
	chain, ok := cm.m[*c]
	if !ok {
		return NewChain(ChainStatusUnknown)
	}
	return chain
}
