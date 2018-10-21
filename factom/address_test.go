package factom_test

import (
	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestZeroAddress(t *testing.T) {
	a := factom.Address{}
	assert.Equal(t, "FA1zT4aFpEvcnPqPCigB3fvGu4Q4mTXY22iiuV69DqE1pNhdF2MC", a.String())
}
