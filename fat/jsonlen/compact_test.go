package jsonlen

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompact(t *testing.T) {
	data := []byte(`{     "hello": "world"` + "\n}")
	compact := []byte(`{"hello":"world"}`)
	assert.Equal(t, compact, Compact(data))
}
