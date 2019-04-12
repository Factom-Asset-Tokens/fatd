package factom

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHeights(t *testing.T) {
	var h Heights
	assert := assert.New(t)
	c := NewClient()
	err := h.Get(c)
	assert.NoError(err)
	zero := uint64(0)
	assert.NotEqual(zero, h.DirectoryBlock)
	assert.NotEqual(zero, h.Leader)
	assert.NotEqual(zero, h.EntryBlock)
	assert.NotEqual(zero, h.Entry)
}
