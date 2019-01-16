package fat0_test

import (
	"testing"

	"github.com/Factom-Asset-Tokens/fatd/fat"
	"github.com/Factom-Asset-Tokens/fatd/fat/fat0"
	"github.com/stretchr/testify/assert"
)

func TestType(t *testing.T) {
	assert.Equal(t, fat.TypeFAT0, fat0.Type)
}
