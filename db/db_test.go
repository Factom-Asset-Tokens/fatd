package db

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/flag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDB(t *testing.T) {
	var err error
	flag.DBPath, err = ioutil.TempDir(os.TempDir(), "fatd.db-test")
	defer func() {
		if err := os.RemoveAll(flag.DBPath); err != nil {
			fmt.Println("failed to remove temp dir:", err)
		}
	}()
	require.NoError(t, err)
	t.Run("Open", func(t *testing.T) {
		require := require.New(t)
		assert := assert.New(t)
		chains, err := OpenAll()
		assert.Empty(chains)
		require.NoError(err)
		eblocks := genChain()
		cp, err := Open(eblocks[0].ChainID)
		require.NoError(err)
		defer cp.Close()
		var dbKeyMR factom.Bytes32
		eID := int64(1)
		for _, eb := range eblocks {
			require.NoError(InsertEBlock(cp.Conn, eb, &dbKeyMR))
			dbKeyMR[0]++
			for _, e := range eb.Entries {
				id, err := InsertEntry(cp.Conn, e, eb.Sequence)
				require.NoError(err)
				assert.Equal(id, eID)
				eID++
			}
		}
		// Ensure only EBlocks with sequential KeyMRs and Sequence
		// numbers can be inserted.
		eb := eblocks[5]
		eb.Sequence = 100
		assert.EqualError(InsertEBlock(cp.Conn, eb, &dbKeyMR),
			"invalid EBlock{}.PrevKeyMR")
		eb = eblocks[5]
		eb.PrevKeyMR = new(factom.Bytes32)
		assert.EqualError(InsertEBlock(cp.Conn, eb, &dbKeyMR),
			"invalid EBlock{}.PrevKeyMR")

		assert.NoError(ValidateChain(cp.Conn, eb.ChainID))
	})
}

var entryCount int

func genNewEntry(chainID *factom.Bytes32) factom.Entry {
	extID := []byte(fmt.Sprintf("%v", entryCount))
	entryCount++
	data := []byte("hello world")
	e := factom.Entry{ChainID: chainID, ExtIDs: []factom.Bytes{extID}, Content: data}
	hash, err := e.ComputeHash()
	if err != nil {
		panic(err)
	}
	e.Hash = &hash
	return e
}

func genChain() []factom.EBlock {
	eblocks := make([]factom.EBlock, 6)
	eb := &eblocks[0]
	height := uint32(10000)
	timestamp := time.Date(2019, 5, 5, 5, 0, 0, 0, time.Local)
	eb.Timestamp = timestamp
	eb.Height = 10000
	eb.PrevKeyMR = new(factom.Bytes32)
	eb.PrevFullHash = new(factom.Bytes32)
	eb.Entries = []factom.Entry{genNewEntry(nil)}
	eb.Entries[0].Timestamp = timestamp.Add(time.Duration(rand.Intn(10)) * time.Minute)
	chainID := eb.Entries[0].ChainID
	eb.ChainID = chainID

	bodyMR, err := eb.ComputeBodyMR()
	if err != nil {
		panic(err)
	}
	eb.BodyMR = &bodyMR

	keyMR, err := eb.ComputeKeyMR()
	if err != nil {
		panic(err)
	}
	eb.KeyMR = &keyMR

	fullHash, err := eb.ComputeFullHash()
	if err != nil {
		panic(err)
	}
	prevKeyMR := &keyMR
	prevFullHash := &fullHash
	for i := range eblocks[1:] {
		eb := &eblocks[i+1]
		eb.Sequence = uint32(i + 1)
		eb.ChainID = chainID
		numBlocks := uint32(rand.Intn(10) + 1)
		height += numBlocks
		timestamp = timestamp.Add(time.Duration(numBlocks) * 10 * time.Minute)
		eb.Timestamp = timestamp
		eb.Height = height
		eb.PrevKeyMR = prevKeyMR
		eb.PrevFullHash = prevFullHash
		eb.Entries = make([]factom.Entry, 2)
		lastTimestamp := eb.Timestamp
		randMinRange := 10
		for i := range eb.Entries {
			e := genNewEntry(chainID)
			// Ensure that the Timestamp is always greater than or
			// equal to the last Entry timestamp.
			rMin := rand.Intn(randMinRange)
			randMinRange -= rMin
			e.Timestamp = lastTimestamp.Add(time.Duration(rMin) * time.Minute)
			lastTimestamp = e.Timestamp

			eb.Entries[i] = e
		}
		bodyMR, err := eb.ComputeBodyMR()
		if err != nil {
			panic(err)
		}
		eb.BodyMR = &bodyMR

		keyMR, err := eb.ComputeKeyMR()
		if err != nil {
			panic(err)
		}
		eb.KeyMR = &keyMR
		prevKeyMR = &keyMR

		fullHash, err := eb.ComputeFullHash()
		if err != nil {
			panic(err)
		}
		prevFullHash = &fullHash
	}
	return eblocks
}

func init() {
	rand.Seed(100)
}
