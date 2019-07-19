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

package factom_test

import (
	"testing"
	"time"

	. "github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var courtesyNode = "https://courtesy-node.factom.com"

func TestDataStructures(t *testing.T) {
	height := uint32(166587)
	c := NewClient()
	c.Factomd.DebugRequest = true
	db := &DBlock{}
	db.Header.Height = height
	t.Run("DBlock", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		// We should start off unpopulated.
		require.False(db.IsPopulated())

		// A bad URL will cause an error.
		c.FactomdServer = "http://example.com"
		assert.Error(db.Get(c))

		c.FactomdServer = courtesyNode
		require.NoError(db.Get(c))

		require.True(db.IsPopulated(), db)
		assert.NoError(db.Get(c)) // Take the early exit code path.

		// Validate this DBlock.
		assert.Len(db.EBlocks, 7)
		assert.Equal(height, db.Header.Height)
		for _, eb := range db.EBlocks {
			assert.NotNil(eb.ChainID)
			assert.NotNil(eb.KeyMR)
		}

		dbk := DBlock{KeyMR: db.KeyMR, FullHash: db.FullHash}
		require.NoError(dbk.Get(c))
		assert.Equal(*db, dbk)

		params := struct {
			Hash *Bytes32 `json:"hash"`
		}{Hash: db.KeyMR}
		var result struct {
			Data Bytes `json:"data"`
		}
		require.NoError(c.FactomdRequest("raw-data", params, &result))

		data, err := db.MarshalBinary()
		require.NoError(err)
		for i := range result.Data {
			assert.Equal(result.Data[i], data[i], i)
		}

		full, err := dbk.ComputeFullHash()
		require.NoError(err, "ComputeFullHash()")
		assert.Equal(*db.FullHash, full, "ComputeFullHash()")

		bodyMR, err := dbk.ComputeBodyMR()
		require.NoError(err, "ComputeBodyMR()")
		assert.Equal(*db.Header.BodyMR, bodyMR, "ComputeBodyMR()")

		keyMR, err := dbk.ComputeKeyMR()
		require.NoError(err, "ComputeKeyMR()")
		assert.Equal(*db.KeyMR, keyMR, "ComputeKeyMR()")

		eb := &db.EBlocks[len(db.EBlocks)-1]
		assert.Equal(eb, db.EBlock(*eb.ChainID))
		assert.Nil(db.EBlock(Bytes32{}))
	})
	t.Run("EBlock", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		// An EBlock without a KeyMR or ChainID should cause an error.
		blank := EBlock{}
		assert.EqualError(blank.Get(c), "no KeyMR or ChainID specified")

		// We'll use the DBlock from the last test, so it must be
		// populated to proceed.
		require.True(db.IsPopulated())

		// This EBlock has multiple entries we can validate against.
		// We'll use a pointer here so that we can reuse this EBlock in
		// the next test.
		eb := &db.EBlocks[4]

		// We start off unpopulated.
		require.False(eb.IsPopulated())

		// A bad URL will cause an error.
		c.FactomdServer = "example.com"
		assert.Error(eb.Get(c))

		c.FactomdServer = courtesyNode
		require.NoError(eb.Get(c))

		require.True(eb.IsPopulated())
		assert.NoError(eb.Get(c)) // Take the early exit code path.

		// Validate the entries.
		assert.Len(eb.Entries, 5)
		assert.Equal(height, eb.Height)
		require.NotNil(eb.PrevKeyMR)
		for _, e := range eb.Entries {
			assert.True(e.ChainID == eb.ChainID)
			assert.NotNil(e.Hash)
			assert.NotNil(e.Timestamp)
			assert.Equal(height, e.Height)
		}

		assert.False(eb.IsFirst())

		// A bad URL will cause an error.
		c.FactomdServer = "example.com"
		_, err := eb.GetAllPrev(c)
		assert.Error(err)

		c.FactomdServer = courtesyNode
		ebs, err := eb.GetAllPrev(c)
		var first EBlock
		if assert.NoError(err) {
			assert.Len(ebs, 6)
			assert.True(ebs[0].IsFirst())
			first = ebs[0].Prev()
			assert.Equal(first.KeyMR, ebs[0].KeyMR,
				"Prev() should return a copy of itself if it is first")
			assert.Equal(eb.KeyMR, ebs[len(ebs)-1].KeyMR)
		}

		// Fetch the chain head EBlock via the ChainID.
		// First use an invalid ChainID and an invalid URL.
		eb2 := EBlock{ChainID: NewBytes32(nil)}
		c.FactomdServer = "example.com"
		assert.Error(eb2.Get(c))
		assert.Error(eb2.GetFirst(c))

		c.FactomdServer = courtesyNode
		require.Error(eb2.Get(c))
		require.False(eb2.IsPopulated())
		assert.EqualError(eb2.GetFirst(c),
			`jsonrpc2.Error{Code:-32009, Message:"Missing Chain Head"}`)
		ebs, err = eb2.GetAllPrev(c)
		assert.EqualError(err,
			`jsonrpc2.Error{Code:-32009, Message:"Missing Chain Head"}`)
		assert.Nil(ebs)

		// A valid ChainID should allow it to be populated.
		eb2.ChainID = eb.ChainID
		require.NoError(eb2.Get(c))
		require.True(eb2.IsPopulated())
		assert.NoError(eb2.GetFirst(c))
		assert.Equal(first.KeyMR, eb2.KeyMR)

		// Make RPC request for this Entry Block.
		params := struct {
			KeyMR *Bytes32 `json:"hash"`
		}{KeyMR: eb2.KeyMR}
		var result struct {
			Data Bytes `json:"data"`
		}
		require.NoError(c.FactomdRequest("raw-data", params, &result))
		data, err := eb2.MarshalBinary()
		require.NoError(err)
		assert.Equal(result.Data, Bytes(data))

		bodyMR, err := eb2.ComputeBodyMR()
		require.NoError(err)
		assert.Equal(*eb2.BodyMR, bodyMR)

		keyMR, err := eb2.ComputeKeyMR()
		require.NoError(err)
		assert.Equal(*eb2.KeyMR, keyMR)
	})
	t.Run("Entry", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		// An EBlock without a KeyMR or ChainID should cause an error.
		blank := Entry{}
		assert.EqualError(blank.Get(c), "Hash is nil")

		// We'll use the DBlock and EBlock from the last test, so they
		// must be populated to proceed.
		require.True(db.IsPopulated())
		eb := db.EBlocks[4]
		require.True(eb.IsPopulated())

		e := eb.Entries[0]
		// We start off unpopulated.
		require.False(e.IsPopulated())

		// A bad URL will cause an error.
		c.FactomdServer = "example.com"
		assert.Error(e.Get(c))

		c.FactomdServer = courtesyNode
		require.NoError(e.Get(c))

		require.True(e.IsPopulated())
		assert.NoError(e.Get(c)) // Take the early exit code path.

		// Validate the entry.
		assert.Len(e.ExtIDs, 6)
		assert.NotEmpty(e.Content)
		assert.Equal(height, e.Height)
		assert.Equal(time.Unix(1542223080, 0), e.Timestamp)
		hash, err := e.ComputeHash()
		assert.NoError(err)
		assert.Equal(*e.Hash, hash)

		e = eb.Entries[1]
		require.NoError(e.Get(c))
		hash, err = e.ComputeHash()
		assert.NoError(err)
		assert.Equal(*e.Hash, hash)
	})

	assert.Equal(t, Bytes32{}, ZeroBytes32())
}
