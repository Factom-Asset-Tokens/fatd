package fat0_test

import (
	"encoding/hex"
	"testing"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat0"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChainID(t *testing.T) {
	id, err := hex.DecodeString(
		"88888807e4f3bbb9a2b229645ab6d2f184224190f83e78761674c2362aca4425")
	require.NoError(t, err)
	issuerChainID := factom.NewBytes32(id)
	assert.Equal(t, "b54c4310530dc4dd361101644fa55cb10aec561e7874a7b786ea3b66f2c6fdfb",
		fat0.ChainID("test", issuerChainID).String())
}
