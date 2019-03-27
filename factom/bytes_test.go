package factom_test

import (
	"testing"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	JSONBytesInvalidTypes     = []string{`{}`, `5.5`, `["hello"]`}
	JSONBytes32InvalidLengths = []string{
		`"00"`,
		`"000000000000000000000000000000000000000000000000000000000000000000"`}
	JSONBytesInvalidSymbol = `"x000000000000000000000000000000000000000000000000000000000000000"`
	JSONBytes32Valid       = `"da56930e8693fb7c0a13aac4d01cf26184d760f2fd92d2f0a62aa630b1a25fa7"`
)

func TestBytes32UnmarshalJSON(t *testing.T) {
	for _, json := range JSONBytesInvalidTypes {
		testBytes32UnmarshalJSON(t, "InvalidType", json, "json: cannot unmarshal object into Go value of type string")
	}
	for _, json := range JSONBytes32InvalidLengths {
		testBytes32UnmarshalJSON(t, "InvalidLength", json, "json: cannot unmarshal number into Go value of type string")
	}
	testBytes32UnmarshalJSON(t, "InvalidSymbol", JSONBytesInvalidSymbol,
		"*factom.Bytes32: encoding/hex: invalid byte: U+0078 'x'")
	json := JSONBytes32Valid
	t.Run("Valid", func(t *testing.T) {
		var b32 factom.Bytes32
		assert.NoErrorf(t, b32.UnmarshalJSON([]byte(json)), "json: %v", json)
	})
}

func testBytes32UnmarshalJSON(t *testing.T, name string, json string, errStr string) {
	t.Run(name, func(t *testing.T) {
		var b32 factom.Bytes32
		assert.EqualErrorf(t, b32.UnmarshalJSON([]byte(json)),
			errStr, "json: %v", json)
	})
}

func TestBytes32MarshalJSON(t *testing.T) {
	b := []byte{0x01}
	b32 := factom.NewBytes32(b)
	assert := assert.New(t)
	assert.Equal(b32[0], b[0])
	assert.Equal(b32[31], byte(0))
	assert.Equal(b32.String(),
		"0100000000000000000000000000000000000000000000000000000000000000")
	data, err := b32.MarshalJSON()
	require.NoError(t, err)
	assert.Equal(string(data),
		`"0100000000000000000000000000000000000000000000000000000000000000"`)
}

func TestBytesUnmarshalJSON(t *testing.T) {
	for _, json := range JSONBytesInvalidTypes {
		testBytesUnmarshalJSON(t, "InvalidType", json,
			"*factom.Bytes: expected JSON string")
	}
	testBytesUnmarshalJSON(t, "InvalidSymbol", JSONBytesInvalidSymbol,
		"*factom.Bytes: encoding/hex: invalid byte: U+0078 'x'")
	json := JSONBytes32Valid
	t.Run("Valid", func(t *testing.T) {
		var b factom.Bytes
		assert.NoErrorf(t, b.UnmarshalJSON([]byte(json)), "json: %v", json)
	})
}

func testBytesUnmarshalJSON(t *testing.T, name string, json string, errStr string) {
	t.Run(name, func(t *testing.T) {
		var b factom.Bytes
		assert.EqualErrorf(t, b.UnmarshalJSON([]byte(json)),
			errStr, "json: %v", json)
	})
}

func TestBytesMarshalJSON(t *testing.T) {
	b := factom.Bytes{0x01}
	assert := assert.New(t)
	assert.Equal(b.String(), "01")
	data, err := b.MarshalJSON()
	require.NoError(t, err)
	assert.Equal(string(data), `"01"`)
}
