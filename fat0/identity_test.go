package fat0_test

import (
	"testing"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat0"
	"github.com/stretchr/testify/assert"
)

func TestValidIdentityChainID(t *testing.T) {
	assert := assert.New(t)

	validIssuerChainID := validIssuerChainID[:]
	assert.False(fat0.ValidIdentityChainID(validIssuerChainID[1:]), "invalid length")

	invalidIssuerChainID := factom.NewBytes32(validIssuerChainID)[:]
	for i := 0; i < 3; i++ {
		invalidIssuerChainID[i] = 0
		assert.Falsef(fat0.ValidIdentityChainID(invalidIssuerChainID),
			"invalid byte [%v]", i)
		invalidIssuerChainID[i] = validIssuerChainID[i]
	}

	assert.True(fat0.ValidIdentityChainID(validIssuerChainID))
}
