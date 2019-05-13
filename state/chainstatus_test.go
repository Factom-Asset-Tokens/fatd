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
