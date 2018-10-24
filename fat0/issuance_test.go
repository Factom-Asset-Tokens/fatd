package fat0_test

import (
	"testing"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat0"
	"github.com/stretchr/testify/assert"
)

func TestChainID(t *testing.T) {
	assert.Equal(t, "b54c4310530dc4dd361101644fa55cb10aec561e7874a7b786ea3b66f2c6fdfb",
		fat0.ChainID("test", validIssuerChainID).String())
}

func TestValidTokenNameIDs(t *testing.T) {
	assert := assert.New(t)
	validNameIDs := []factom.Bytes{
		factom.Bytes("token"),
		factom.Bytes("valid"),
		factom.Bytes("issuer"),
		validIssuerChainID[:],
	}

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

func TestIssuance(t *testing.T) {
	t.Run("ValidExtIDs()", func(t *testing.T) {
		assert := assert.New(t)
		invalidIssuance := fat0.NewIssuance(&factom.Entry{})
		assert.Error(invalidIssuance.ValidExtIDs(), "invalid number of ExtIDs")

		invalidIssuance.ExtIDs = append([]factom.Bytes{}, validIssuance.ExtIDs...)

		invalidIssuance.ExtIDs[0] = invalidIssuance.ExtIDs[0][0 : fat0.RCDSize-1]
		assert.Error(invalidIssuance.ValidExtIDs(), "invalid RCD length")
		invalidIssuance.ExtIDs[0] = validIssuance.ExtIDs[0]

		invalidIssuance.ExtIDs[1] =
			invalidIssuance.ExtIDs[1][0 : fat0.SignatureSize-1]
		assert.Error(invalidIssuance.ValidExtIDs(), "invalid signature length")
		invalidIssuance.ExtIDs[1] = validIssuance.ExtIDs[1]

		invalidIssuance.ExtIDs[0][0] = 0
		assert.Error(invalidIssuance.ValidExtIDs(), "invalid RCD type")
		invalidIssuance.ExtIDs[0][0] = fat0.RCDType

		assert.NoError(validIssuance.ValidExtIDs())
		validIssuance.ExtIDs = append(validIssuance.ExtIDs, []byte{0})
		assert.NoError(validIssuance.ValidExtIDs(), "additional ExtIDs")
		validIssuance.ExtIDs = validIssuance.ExtIDs[0:2]
	})
	t.Run("Unmarshal()", func(t *testing.T) {
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
			invalidIssuanceEntryContentMap[k] = v
		}

		assert.NoError(validIssuance.Unmarshal())
		assert.Equal(validIssuanceEntryContentMap["type"],
			validIssuance.Type, "type")
		assert.Equal(validIssuanceEntryContentMap["symbol"],
			validIssuance.Symbol, "symbol")
		assert.Equal(validIssuanceEntryContentMap["supply"],
			validIssuance.Supply, "supply")
		assert.Equal(validIssuanceEntryContentMap["name"],
			validIssuance.Name, "name")

		// Metadata can be any type.
		validIssuanceEntryContentMap["metadata"] = []int{0}
		content := validIssuance.Content
		validIssuance.Content = marshal(validIssuanceEntryContentMap)
		assert.NoError(validIssuance.Unmarshal())
		assert.NotNil(validIssuance.Metadata, "metadata")
		validIssuance.Content = content
	})
	t.Run("ValidData()", func(t *testing.T) {
		assert := assert.New(t)
		invalidIssuance := &fat0.Issuance{
			Type:   validIssuance.Type,
			Supply: validIssuance.Supply,
			Symbol: validIssuance.Symbol,
			Name:   validIssuance.Name,
		}

		invalidIssuance.Type = "invalid"
		assert.Error(invalidIssuance.ValidData(), "type")
		invalidIssuance.Type = validIssuance.Type

		invalidIssuance.Supply = 0
		assert.Error(invalidIssuance.ValidData())
		invalidIssuance.Supply = validIssuance.Supply

		invalidIssuance.Symbol = ""
		assert.NoError(invalidIssuance.ValidData(), "symbol is optional")
		invalidIssuance.Symbol = validIssuance.Symbol

		invalidIssuance.Name = ""
		assert.NoError(invalidIssuance.ValidData(), "name is optional")
		invalidIssuance.Name = validIssuance.Name

		assert.NoError(validIssuance.ValidData())
	})
	t.Run("ValidSignature()", func(t *testing.T) {
		assert := assert.New(t)
		invalidIssuance := fat0.NewIssuance(&factom.Entry{})
		invalidIssuance.ChainID = validIssuance.ChainID
		invalidIssuance.Content = validIssuance.Content
		invalidIssuance.ExtIDs = append([]factom.Bytes{}, validIssuance.ExtIDs...)
		invalidIssuance.ExtIDs[1] = append(factom.Bytes{}, validIssuance.ExtIDs[1]...)
		invalidIssuance.ExtIDs[1][0]++
		assert.False(invalidIssuance.ValidSignature())

		assert.True(validIssuance.ValidSignature())
	})
	t.Run("Valid()", func(t *testing.T) {
		assert := assert.New(t)
		validIDKey := issuerKey.RCDHash()

		invalidIssuance := fat0.NewIssuance(&factom.Entry{})
		invalidIssuance.ChainID = validIssuance.ChainID
		invalidIssuance.Content = append([]byte{}, validIssuance.Content...)

		assert.Error(invalidIssuance.Valid(validIDKey), "invalid ExtIDs")
		invalidIssuance.ExtIDs = append([]factom.Bytes{}, validIssuance.ExtIDs...)

		invalidIDKey := *factom.NewBytes32([]byte{0})
		assert.Error(validIssuance.Valid(invalidIDKey), "invalid id key")

		invalidIssuance.Content[0]++
		assert.Error(invalidIssuance.Valid(validIDKey), "unmarshal")
		invalidIssuance.Content[0]--

		invalidIssuanceEntryContentMap := make(map[string]interface{})
		mapCopy(invalidIssuanceEntryContentMap, validIssuanceEntryContentMap)
		invalidIssuanceEntryContentMap["supply"] = 0
		content := invalidIssuance.Content
		extIDs := invalidIssuance.ExtIDs
		invalidIssuance.Content = marshal(invalidIssuanceEntryContentMap)
		invalidIssuance.Sign(issuerKey)
		assert.Error(invalidIssuance.Valid(validIDKey), "invalid data")
		invalidIssuance.Content = content
		invalidIssuance.ExtIDs = extIDs

		invalidIssuance.ExtIDs[1][0]++
		assert.Error(invalidIssuance.Valid(validIDKey), "invalid signature")
		invalidIssuance.ExtIDs[1][0]--

		assert.NoError(validIssuance.Valid(validIDKey))
	})
}
