package fat0_test

import (
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat0"
	"github.com/stretchr/testify/assert"
)

var (
	validIssuerChainID *factom.Bytes32
)

func init() {
	id, _ := hex.DecodeString(
		"88888807e4f3bbb9a2b229645ab6d2f184224190f83e78761674c2362aca4425")
	validIssuerChainID = factom.NewBytes32(id)
	validIssuanceEntry.Content = marshal(validIssuanceEntryContentMap)
}

func marshal(v map[string]interface{}) []byte {
	data, _ := json.Marshal(v)
	return data
}

func TestChainID(t *testing.T) {
	assert.Equal(t, "b54c4310530dc4dd361101644fa55cb10aec561e7874a7b786ea3b66f2c6fdfb",
		fat0.ChainID("test", validIssuerChainID).String())
}

func TestValidTokenNameIDs(t *testing.T) {
	validNameIDs := []factom.Bytes{
		factom.Bytes("token"),
		factom.Bytes("valid"),
		factom.Bytes("issuer"),
		validIssuerChainID[:],
	}
	assert := assert.New(t)

	invalidNameIDs := append(validNameIDs, []byte{})
	assert.False(fat0.ValidTokenNameIDs(invalidNameIDs), "invalid length")

	invalidNameIDs = invalidNameIDs[:4]
	invalidName := factom.Bytes("")
	for i := 0; i < 4; i++ {
		invalidNameIDs[i] = invalidName
		assert.Falsef(fat0.ValidTokenNameIDs(invalidNameIDs),
			"invalid name id [%v]", i)
		invalidNameIDs[i] = validNameIDs[i]
	}
	assert.True(fat0.ValidTokenNameIDs(validNameIDs))
}

var validIssuanceEntry factom.Entry

var validIssuanceEntryContentMap = map[string]interface{}{
	"type":   "FAT-0",
	"supply": int64(100000),
	"symbol": "TEST",
	"name":   "Test Token",
}

func TestIssuanceUnmarshal(t *testing.T) {
	validIssuance := fat0.NewIssuance(&validIssuanceEntry)
	assert := assert.New(t)

	var invalidIssuanceEntry factom.Entry
	invalidIssuance := fat0.NewIssuance(&invalidIssuanceEntry)
	assert.Error(invalidIssuance.Unmarshal(), "no content")

	// Initialize content map to be equal to the valid map.
	invalidIssuanceEntryContentMap := make(map[string]interface{})
	mapCopy(invalidIssuanceEntryContentMap, validIssuanceEntryContentMap)

	invalidIssuanceEntryContentMap["extra"] = "extra"
	invalidIssuance.Content = marshal(invalidIssuanceEntryContentMap)
	assert.Error(invalidIssuance.Unmarshal(), "extra unrecognized field")
	delete(invalidIssuanceEntryContentMap, "extra")

	// Try to use an invalid value for each field.
	var invalid = []int{0}
	for k, v := range invalidIssuanceEntryContentMap {
		invalidIssuanceEntryContentMap[k] = invalid
		invalidIssuance.Content = marshal(invalidIssuanceEntryContentMap)
		assert.Errorf(invalidIssuance.Unmarshal(),
			"invalid type for field %#v", k)
		invalidIssuanceEntryContentMap["type"] = v
	}

	assert.NoError(validIssuance.Unmarshal())
	assert.Equal(validIssuance.Type, validIssuanceEntryContentMap["type"], "type")
	assert.Equal(validIssuance.Symbol, validIssuanceEntryContentMap["symbol"], "symbol")
	assert.Equal(validIssuance.Supply, validIssuanceEntryContentMap["supply"], "supply")
	assert.Equal(validIssuance.Name, validIssuanceEntryContentMap["name"], "name")

	// Metadata can be any type.
	validIssuanceEntryContentMap["metadata"] = []int{0}
	validIssuance.Content = marshal(validIssuanceEntryContentMap)
	assert.NoError(validIssuance.Unmarshal())
	assert.NotNil(validIssuance.Metadata, "metadata")
}

func mapCopy(dst, src map[string]interface{}) {
	for k, v := range src {
		dst[k] = v
	}
}
