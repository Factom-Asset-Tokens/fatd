package factom_test

import (
	"fmt"
	"testing"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var humanReadableZeroAddress = "FA1zT4aFpEvcnPqPCigB3fvGu4Q4mTXY22iiuV69DqE1pNhdF2MC"

var humanReadableAddress = "FA2MwhbJFxPckPahsmntwF1ogKjXGz8FSqo2cLWtshdU47GQVZDC"

func TestZeroAddress(t *testing.T) {
	a := factom.Address{}
	require := require.New(t)
	require.Equal(humanReadableZeroAddress, a.String())
	rcdHash := a.RCDHash()
	a2 := factom.NewAddress(rcdHash)
	require.Equal(humanReadableZeroAddress, a2.String())
}

var (
	JSONAddressInvalidTypes   = []string{`{}`, `5.5`, `["hello"]`}
	JSONAddressInvalidLengths = []string{
		`"FA0"`, `"FA1zT4aFpEvcnPqPCigB3fvGu4Q4mTXY22iiuV69DqE1pNhdF2MC1zT4"`}
	JSONAddressInvalidPrefix   = `"Fs1zT4aFpEvcnPqPCigB3fvGu4Q4mTXY22iiuV69DqE1pNhdF2MC"`
	JSONAddressInvalidSymbol   = `"FA2zT4aFpEvcnPqPCigB3fvGu4Q4mTXY22l0uV69DqE1pNhdF2MC"`
	JSONAddressInvalidChecksum = `"FA2zT4aFpEvcnPqPCigB3fvGu4Q4mTXY22iiuV69DqE1pNhdF2MC"`
)

func TestAddressUnmarshalJSON(t *testing.T) {
	for _, json := range JSONAddressInvalidTypes {
		testAddressUnmarshalJSON(t, "InvalidType", json,
			"*factom.Address: expected JSON string")
	}
	for _, json := range JSONAddressInvalidLengths {
		testAddressUnmarshalJSON(t, "InvalidLength", json,
			"*factom.Address: invalid length")
	}
	testAddressUnmarshalJSON(t, "InvalidPrefix", JSONAddressInvalidPrefix,
		"*factom.Address: invalid prefix")
	testAddressUnmarshalJSON(t, "InvalidSymbol", JSONAddressInvalidSymbol,
		"invalid format: version and/or checksum bytes missing")
	testAddressUnmarshalJSON(t, "InvalidChecksum", JSONAddressInvalidChecksum,
		"checksum error")
	json := fmt.Sprintf("%#v", humanReadableAddress)
	t.Run("Valid", func(t *testing.T) {
		var address factom.Address
		assert := assert.New(t)
		assert.NoErrorf(address.UnmarshalJSON([]byte(json)), "json: %v", json)
		assert.Equal(humanReadableAddress, address.String())
	})
}

func testAddressUnmarshalJSON(t *testing.T, name string, json string, errStr string) {
	t.Run(name, func(t *testing.T) {
		var address factom.Address
		assert.EqualErrorf(t, address.UnmarshalJSON([]byte(json)),
			errStr, "json: %v", json)
	})
}

func TestMarshalJSON(t *testing.T) {
	json := fmt.Sprintf("%#v", humanReadableZeroAddress)
	require := require.New(t)
	var address factom.Address
	require.NoErrorf(address.UnmarshalJSON([]byte(json)), "json: %v", json)
	data, err := address.MarshalJSON()
	require.NoError(err)
	require.Equal(string(data), json)
}
