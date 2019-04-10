package state_test

import (
	"testing"

	. "github.com/Factom-Asset-Tokens/fatd/state"
	"github.com/stretchr/testify/assert"
)

const (
	unknownID int = iota
	trackedID
	issuedID
	ignoredID
)

var chainStatusTests = []struct {
	ChainStatus
	expected [4]bool
}{{
	ChainStatus: ChainStatusUnknown,
	expected:    [4]bool{unknownID: true},
}, {
	ChainStatus: ChainStatusTracked,
	expected:    [4]bool{trackedID: true},
}, {
	ChainStatus: ChainStatusIssued,
	expected:    [4]bool{trackedID: true, issuedID: true},
}, {
	ChainStatus: ChainStatusIgnored,
	expected:    [4]bool{ignoredID: true},
}}

func TestChainStatus(t *testing.T) {
	for _, test := range chainStatusTests {
		status := test.ChainStatus
		t.Run(status.String(), func(t *testing.T) {
			assert := assert.New(t)
			expected := test.expected
			assert.Equalf(expected[unknownID], status.IsUnknown(),
				"IsUnknown()")
			assert.Equalf(expected[trackedID], status.IsTracked(),
				"IsTracked()")
			assert.Equalf(expected[issuedID], status.IsIssued(),
				"IsIssued()")
			assert.Equalf(expected[ignoredID], status.IsIgnored(),
				"IsIgnored()")
		})
	}
}
