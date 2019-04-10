package fat1_test

import (
	"testing"

	"github.com/Factom-Asset-Tokens/fatd/fat"
	"github.com/Factom-Asset-Tokens/fatd/fat/fat1"
	"github.com/stretchr/testify/assert"
)

func TestType(t *testing.T) {
	assert.Equal(t, fat.TypeFAT1, fat1.Type)
}
