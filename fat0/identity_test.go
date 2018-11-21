package fat0_test

import (
	"encoding/hex"
	"testing"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat0"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	validIdentityChainID = hexToBytes(
		"88888807e4f3bbb9a2b229645ab6d2f184224190f83e78761674c2362aca4425")
)

func hexToBytes(hexStr string) factom.Bytes {
	raw, err := hex.DecodeString(hexStr)
	if err != nil {
		panic(err)
	}
	return factom.Bytes(raw)
}

func TestValidIdentityChainID(t *testing.T) {
	assert := assert.New(t)

	assert.True(fat0.ValidIdentityChainID(validIdentityChainID))

	assert.False(fat0.ValidIdentityChainID(validIdentityChainID[0:10]), "invalid length")

	// Make a copy that we can use for invalid tests.
	invalidIdentityChainID := append(factom.Bytes{}, validIdentityChainID...)
	for i := 0; i < 3; i++ {
		invalidIdentityChainID[i] = 0
		assert.Falsef(fat0.ValidIdentityChainID(invalidIdentityChainID),
			"invalid byte [%v]", i)
		invalidIdentityChainID[i] = validIdentityChainID[i]
	}
}

var (
	validIdentityNameIDs = []factom.Bytes{
		factom.Bytes{0x00},
		factom.Bytes("Identity Chain"),
		hexToBytes("f825c5629772afb5bce0464e5ea1af244be853a692d16360b8e03d6164b6adb5"),
		hexToBytes("28baa7d04e6c102991a184533b9f2443c9c314cc0327cc3a2f2adc0f3d7373a1"),
		hexToBytes("6095733cf6f5d0b5411d1eeb9f6699fad1ae27f9d4da64583bef97008d7bf0c9"),
		hexToBytes("966ebc2a0e3877ed846167e95ba3dde8561d90ee9eddd1bb74fbd6d1d25dba0f"),
		hexToBytes("33363533323533"),
	}
)

func TestValidIdentityNameIDs(t *testing.T) {
	assert := assert.New(t)
	assert.True(fat0.ValidIdentityNameIDs(validIdentityNameIDs))

	// Make a copy that we can use for invalid tests.
	invalidIdentityNameIDs := append([]factom.Bytes{}, validIdentityNameIDs...)
	for i := range validIdentityNameIDs {
		invalidIdentityNameIDs[i] = append(factom.Bytes{}, validIdentityNameIDs[i]...)
	}
	assert.False(fat0.ValidIdentityNameIDs(append(invalidIdentityNameIDs, factom.Bytes{})),
		"invalid length")
	assert.False(fat0.ValidIdentityNameIDs(invalidIdentityNameIDs[0:6]), "invalid length")

	for i := 0; i < 6; i++ {
		tmp := invalidIdentityNameIDs[i]
		invalidIdentityNameIDs[i] = factom.Bytes{}
		assert.False(fat0.ValidIdentityNameIDs(invalidIdentityNameIDs), "invalid byte")
		invalidIdentityNameIDs[i] = tmp
	}
}

func TestIdentity(t *testing.T) {
	i := fat0.Identity{}
	assert := assert.New(t)
	require := require.New(t)
	assert.False(i.IsPopulated())

	assert.EqualError(i.Get(), "ChainID is nil")

	// Invalid Identity ChainID
	i.ChainID = &factom.Bytes32{}
	require.NoError(i.Get())
	assert.False(i.IsPopulated())

	// Valid but non-existant Identity ChainID
	i.ChainID = &factom.Bytes32{0x88, 0x88, 0x88}
	assert.EqualError(i.Get(), "Post http:///v2: http: no Host in request URL")

	factom.RpcConfig.FactomdServer = "courtesy-node.factom.com"
	require.NoError(i.Get())
	assert.False(i.IsPopulated())

	// Valid but existing Identity ChainID, but a malformed Identity Chain.
	i.ChainID = factom.NewBytes32(hexToBytes(
		"8888885c2e0b523d9b8ab6d2975639e431eaba3fc9039ead32ce5065dcde86e4"))
	require.NoError(i.Get())
	assert.False(i.IsPopulated())

	// Valid existing Identity ChainID
	i.ChainID = factom.NewBytes32(validIdentityChainID)
	require.NoError(i.Get())
	require.True(i.IsPopulated())
	// Take early exit path for an already populated Identity.
	require.NoError(i.Get())
}
