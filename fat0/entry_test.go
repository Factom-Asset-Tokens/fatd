package fat0

import (
	"testing"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/stretchr/testify/assert"
)

func TestEntryValidSignatures(t *testing.T) {
	var e Entry
	assert := assert.New(t)

	// None of these should be tree because the number is invalid or there
	// are insufficient ExtIDs.
	assert.False(e.validSignatures(-1), "num = -1")
	assert.False(e.validSignatures(0), "num = zero")
	assert.False(e.validSignatures(1), "2*num > len(e.ExtIds)")

	// Generate valid signatures with blank Addresses.
	as := [2]factom.Address{}
	e.ChainID = factom.NewBytes32(nil)
	e.Sign(as[:]...)

	// Now these Signatures should be valid.
	assert.True(e.validSignatures(1))
	assert.True(e.validSignatures(2))

	// Mess up a signature.
	e.ExtIDs[1][0]++
	assert.False(e.validSignatures(2), "invalid signature")
}
