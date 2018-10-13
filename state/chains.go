package state

import (
	"sync"

	"bitbucket.org/canonical-ledgers/fatd/factom"
)

type ChainStatus int

const (
	ChainStatusUnknown ChainStatus = -1
	ChainStatusIgnored ChainStatus = 0
	ChainStatusTracked ChainStatus = 1
	ChainStatusIssued  ChainStatus = 2
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
	chains = chainMap{m: map[factom.Bytes32]ChainStatus{
		factom.Bytes32{31: 0x0a}: ChainStatusIgnored,
		factom.Bytes32{31: 0x0c}: ChainStatusIgnored,
		factom.Bytes32{31: 0x0f}: ChainStatusIgnored,
	}}
)

type chainMap struct {
	m map[factom.Bytes32]ChainStatus
	sync.RWMutex
}

func (cm chainMap) Ignore(c *factom.Bytes32) {
	cm.Set(c, ChainStatusIgnored)
}
func (cm chainMap) Track(c *factom.Bytes32) {
	cm.Set(c, ChainStatusTracked)
}
func (cm chainMap) Issue(c *factom.Bytes32) {
	cm.Set(c, ChainStatusIssued|ChainStatusTracked)
}

func (cm chainMap) Set(c *factom.Bytes32, status ChainStatus) {
	defer cm.Unlock()
	cm.Lock()
	cm.m[*c] = status
}

func (cm chainMap) Get(c *factom.Bytes32) ChainStatus {
	defer cm.RUnlock()
	cm.RLock()
	status, ok := cm.m[*c]
	if !ok {
		return ChainStatusUnknown
	}
	return status
}
