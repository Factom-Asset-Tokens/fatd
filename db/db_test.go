package db

import (
	"fmt"
	"testing"

	"github.com/Factom-Asset-Tokens/fatd/flag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidate(t *testing.T) {
	require := require.New(t)
	flag.DBPath = "./test-fatd.db"
	flag.LogDebug = true
	chains, err := OpenAll()
	require.NoError(err, "OpenAll()")
	require.NotEmptyf(chains, "Test database is empty: %v", flag.DBPath)

	for _, chain := range chains {
		fmt.Printf("%+v\n", chain)
		assert.NoErrorf(t, chain.Validate(), "Chain{%v}.Validate()", chain.ID)
	}
}
