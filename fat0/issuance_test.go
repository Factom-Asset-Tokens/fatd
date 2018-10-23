package fat0_test

import (
	"encoding/hex"
	"encoding/json"
	"math/rand"
	"testing"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat0"
	"github.com/FactomProject/ed25519"
	"github.com/stretchr/testify/assert"
)

var (
	validIssuerChainID *factom.Bytes32
	validIssuanceEntry factom.Entry
	validIssuance      *fat0.Issuance
	address            factom.Address

	validIssuanceEntryContentMap = map[string]interface{}{
		"type":   "FAT-0",
		"supply": int64(100000),
		"symbol": "TEST",
		"name":   "Test Token",
	}
)

func init() {
	id, _ := hex.DecodeString(
		"88888807e4f3bbb9a2b229645ab6d2f184224190f83e78761674c2362aca4425")
	validIssuerChainID = factom.NewBytes32(id)
	validIssuanceEntry.EBlock = &factom.EBlock{
		ChainID: fat0.ChainID("test", validIssuerChainID)}
	validIssuanceEntry.Content = marshal(validIssuanceEntryContentMap)

	rand := rand.New(rand.NewSource(100))
	address.PublicKey, address.PrivateKey, _ = ed25519.GenerateKey(rand)

	validIssuance = fat0.NewIssuance(&validIssuanceEntry)
	validIssuance.Sign(&address)
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

func TestIssuance(t *testing.T) {
	t.Run("ValidExtIDs()", func(t *testing.T) {
		assert := assert.New(t)
		invalidIssuance := fat0.NewIssuance(&factom.Entry{})
		assert.False(invalidIssuance.ValidExtIDs(), "invalid number of ExtIDs")

		invalidIssuance.ExtIDs = append([]factom.Bytes{}, validIssuance.ExtIDs...)

		invalidIssuance.ExtIDs[0] = invalidIssuance.ExtIDs[0][0 : fat0.RCDSize-1]
		assert.False(invalidIssuance.ValidExtIDs(), "invalid RCD length")
		invalidIssuance.ExtIDs[0] = validIssuance.ExtIDs[0]

		invalidIssuance.ExtIDs[1] =
			invalidIssuance.ExtIDs[1][0 : fat0.SignatureSize-1]
		assert.False(invalidIssuance.ValidExtIDs(), "invalid signature length")
		invalidIssuance.ExtIDs[1] = validIssuance.ExtIDs[1]

		invalidIssuance.ExtIDs[0][0] = 0
		assert.False(invalidIssuance.ValidExtIDs(), "invalid RCD type")
		invalidIssuance.ExtIDs[0][0] = fat0.RCDType

		assert.True(validIssuance.ValidExtIDs())
		validIssuance.ExtIDs = append(validIssuance.ExtIDs, []byte{0})
		assert.True(validIssuance.ValidExtIDs(), "additional ExtIDs")
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
			invalidIssuanceEntryContentMap["type"] = v
		}

		assert.NoError(validIssuance.Unmarshal())
		assert.Equal(validIssuance.Type,
			validIssuanceEntryContentMap["type"], "type")
		assert.Equal(validIssuance.Symbol,
			validIssuanceEntryContentMap["symbol"], "symbol")
		assert.Equal(validIssuance.Supply,
			validIssuanceEntryContentMap["supply"], "supply")
		assert.Equal(validIssuance.Name,
			validIssuanceEntryContentMap["name"], "name")

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
		assert.False(invalidIssuance.ValidData(), "type")
		invalidIssuance.Type = validIssuance.Type

		invalidIssuance.Supply = 0
		assert.False(invalidIssuance.ValidData())
		invalidIssuance.Supply = validIssuance.Supply

		invalidIssuance.Symbol = ""
		assert.True(invalidIssuance.ValidData(), "symbol is optional")
		invalidIssuance.Symbol = validIssuance.Symbol

		invalidIssuance.Name = ""
		assert.True(invalidIssuance.ValidData(), "name is optional")
		invalidIssuance.Name = validIssuance.Name

		assert.True(validIssuance.ValidData())
	})
	t.Run("VerifySignature()", func(t *testing.T) {
		assert := assert.New(t)
		invalidIssuance := fat0.NewIssuance(&factom.Entry{EBlock: &factom.EBlock{}})
		invalidIssuance.ChainID = validIssuance.ChainID
		invalidIssuance.Content = validIssuance.Content
		invalidIssuance.ExtIDs = append([]factom.Bytes{}, validIssuance.ExtIDs...)
		invalidIssuance.ExtIDs[1] = append(factom.Bytes{}, validIssuance.ExtIDs[1]...)
		invalidIssuance.ExtIDs[1][0]++
		assert.False(invalidIssuance.VerifySignature())

		assert.True(validIssuance.VerifySignature())
	})
	t.Run("Valid()", func(t *testing.T) {
		assert := assert.New(t)
		validIDKey := address.RCDHash()

		invalidIssuance := fat0.NewIssuance(&factom.Entry{EBlock: &factom.EBlock{}})
		invalidIssuance.ChainID = validIssuance.ChainID
		invalidIssuance.Content = append([]byte{}, validIssuance.Content...)

		assert.False(invalidIssuance.Valid(&validIDKey), "invalid ExtIDs")
		invalidIssuance.ExtIDs = append([]factom.Bytes{}, validIssuance.ExtIDs...)

		invalidIDKey := factom.NewBytes32([]byte{0})
		assert.False(validIssuance.Valid(invalidIDKey), "invalid id key")

		invalidIssuance.Content[0]++
		assert.False(invalidIssuance.Valid(&validIDKey), "unmarshal")
		invalidIssuance.Content[0]--

		invalidIssuanceEntryContentMap := make(map[string]interface{})
		mapCopy(invalidIssuanceEntryContentMap, validIssuanceEntryContentMap)
		invalidIssuanceEntryContentMap["supply"] = 0
		content := invalidIssuance.Content
		extIDs := invalidIssuance.ExtIDs
		invalidIssuance.Content = marshal(invalidIssuanceEntryContentMap)
		invalidIssuance.Sign(&address)
		assert.False(invalidIssuance.Valid(&validIDKey), "invalid data")
		invalidIssuance.Content = content
		invalidIssuance.ExtIDs = extIDs

		invalidIssuance.ExtIDs[1][0]++
		assert.False(invalidIssuance.Valid(&validIDKey), "invalid signature")
		invalidIssuance.ExtIDs[1][0]--

		assert.True(validIssuance.Valid(&validIDKey))
	})
}

func marshal(v map[string]interface{}) []byte {
	data, _ := json.Marshal(v)
	return data
}

func mapCopy(dst, src map[string]interface{}) {
	for k, v := range src {
		dst[k] = v
	}
}
