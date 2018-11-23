package fat0_test

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat0"
	"github.com/FactomProject/ed25519"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	identityChainID = factom.NewBytes32(validIdentityChainID)
)

func TestChainID(t *testing.T) {
	assert.Equal(t, "b54c4310530dc4dd361101644fa55cb10aec561e7874a7b786ea3b66f2c6fdfb",
		fat0.ChainID("test", identityChainID).String())
}

var (
	validNameIDs = []factom.Bytes{
		factom.Bytes("token"),
		factom.Bytes("valid"),
		factom.Bytes("issuer"),
		identityChainID[:],
	}
)

func TestValidTokenNameIDs(t *testing.T) {
	assert := assert.New(t)

	invalidNameIDs := append([]factom.Bytes{}, validNameIDs...)
	assert.False(fat0.ValidTokenNameIDs(invalidNameIDs[:3]), "invalid length")

	invalidName := factom.Bytes{}
	for i := 0; i < 4; i++ {
		invalidNameIDs[i] = invalidName
		assert.Falsef(fat0.ValidTokenNameIDs(invalidNameIDs),
			"invalid name id [%v]", i)
		invalidNameIDs[i] = validNameIDs[i]
	}
	assert.True(fat0.ValidTokenNameIDs(validNameIDs))
}

var (
	randSource = rand.New(rand.NewSource(100))
	issuerKey  = func() factom.Address {
		a := factom.Address{}
		a.PublicKey, a.PrivateKey, _ = ed25519.GenerateKey(randSource)
		return a
	}()
	validIssuanceEntryContentMap = map[string]interface{}{
		"type":     "FAT-0",
		"supply":   int64(100000),
		"symbol":   "TEST",
		"name":     "Test Token",
		"metadata": []int{0},
	}
	validIssuance = func() fat0.Issuance {
		e := factom.Entry{
			ChainID: factom.NewBytes32(nil),
			Content: marshal(validIssuanceEntryContentMap),
		}
		i := fat0.NewIssuance(e)
		i.Sign(issuerKey)
		return i
	}()
)

func TestIssuance(t *testing.T) {
	t.Run("ValidExtIDs()", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		// Blank ExtIDs
		invalidIssuance := fat0.NewIssuance(factom.Entry{})
		assert.EqualError(invalidIssuance.ValidExtIDs(),
			"insufficient number of ExtIDs")

		// Create a working copy of the valid ExtIDs.
		invalidIssuance.ExtIDs = copyExtIDs(validIssuance.ExtIDs)
		require.NoError(invalidIssuance.ValidExtIDs())

		// Bad RCD length.
		tmp := invalidIssuance.ExtIDs[0]
		invalidIssuance.ExtIDs[0] = invalidIssuance.ExtIDs[0][0 : factom.RCDSize-1]
		assert.EqualError(invalidIssuance.ValidExtIDs(), "invalid RCD size")
		invalidIssuance.ExtIDs[0] = tmp
		require.NoError(invalidIssuance.ValidExtIDs())

		// Bad RCD type.
		invalidIssuance.ExtIDs[0][0] = 0
		assert.EqualError(invalidIssuance.ValidExtIDs(), "invalid RCD type")
		invalidIssuance.ExtIDs[0][0] = factom.RCDType
		require.NoError(invalidIssuance.ValidExtIDs())

		// Bad Signature length.
		invalidIssuance.ExtIDs[1] =
			invalidIssuance.ExtIDs[1][0 : factom.SignatureSize-1]
		assert.EqualError(invalidIssuance.ValidExtIDs(), "invalid signature size")
		invalidIssuance.ExtIDs[1] = validIssuance.ExtIDs[1]
		require.NoError(invalidIssuance.ValidExtIDs())

		// Additional ExtIDs are allowed.
		validIssuance.ExtIDs = append(validIssuance.ExtIDs, []byte{0})
		assert.NoError(validIssuance.ValidExtIDs(), "additional ExtIDs")
		validIssuance.ExtIDs = validIssuance.ExtIDs[0:2]
	})
	t.Run("UnmarshalEntry()", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		// We must start with a validIssuance.
		require.NoError(validIssuance.UnmarshalEntry())
		assert.Equal(validIssuanceEntryContentMap["type"],
			validIssuance.Type, "type")
		assert.Equal(validIssuanceEntryContentMap["symbol"],
			validIssuance.Symbol, "symbol")
		assert.Equal(validIssuanceEntryContentMap["supply"],
			validIssuance.Supply, "supply")
		assert.Equal(validIssuanceEntryContentMap["name"],
			validIssuance.Name, "name")
		// Metadata can be any type.
		assert.NotNil(validIssuance.Metadata, "metadata")

		// Nil content should fail.
		invalidIssuance := validIssuance
		invalidIssuance.Content = nil
		assert.Error(invalidIssuance.UnmarshalEntry(), "no content")

		// Initialize content map to be equal to the valid map.
		invalidIssuanceEntryContentMap := copyContentMap(validIssuanceEntryContentMap)

		// An invalid field should cause an error.
		invalidField := "invalid"
		invalidIssuanceEntryContentMap[invalidField] = invalidField
		invalidIssuance.Content = marshal(invalidIssuanceEntryContentMap)
		assert.EqualError(invalidIssuance.UnmarshalEntry(),
			fmt.Sprintf("json: unknown field %#v", invalidField))
		delete(invalidIssuanceEntryContentMap, invalidField)
		invalidIssuance.Content = marshal(invalidIssuanceEntryContentMap)
		require.NoError(invalidIssuance.UnmarshalEntry())

		// Try to use an invalid value for each field, except for
		// "metadata".
		invalidValue := []int{0}
		delete(invalidIssuanceEntryContentMap, "metadata")
		for k, v := range invalidIssuanceEntryContentMap {
			invalidIssuanceEntryContentMap[k] = invalidValue
			invalidIssuance.Content = marshal(invalidIssuanceEntryContentMap)
			assert.Errorf(invalidIssuance.UnmarshalEntry(),
				"invalid type for field %#v", k)
			invalidIssuanceEntryContentMap[k] = v
			invalidIssuance.Content = marshal(invalidIssuanceEntryContentMap)
			require.NoError(invalidIssuance.UnmarshalEntry())
		}

	})
	t.Run("ValidData()", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		// Start with validData.
		require.NoError(validIssuance.UnmarshalEntry())
		require.NoError(validIssuance.ValidData())
		invalidIssuance := validIssuance

		// Invalid Type
		invalidIssuance.Type = "invalid"
		assert.EqualError(invalidIssuance.ValidData(),
			fmt.Sprintf(`invalid "type": %#v`, invalidIssuance.Type))
		invalidIssuance = validIssuance
		require.NoError(invalidIssuance.ValidData())

		// Invalid Supply
		invalidIssuance.Supply = 0
		assert.Error(invalidIssuance.ValidData())
		invalidIssuance = validIssuance
		require.NoError(invalidIssuance.ValidData())

		// Optional Symbol
		invalidIssuance.Symbol = ""
		assert.NoError(invalidIssuance.ValidData(), "symbol is optional")
		invalidIssuance = validIssuance
		require.NoError(invalidIssuance.ValidData())

		// Optional Name
		invalidIssuance.Name = ""
		assert.NoError(invalidIssuance.ValidData(), "name is optional")
		invalidIssuance.Name = validIssuance.Name
	})
	t.Run("ValidSignature()", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		// Make a working copy.
		require.True(validIssuance.ValidSignature())
		invalidIssuance := validIssuance
		invalidIssuance.ExtIDs = copyExtIDs(validIssuance.ExtIDs)

		// Mucking with the second byte in the RCD and the signature
		// should make the signature fail.
		// We use the second byte because the first byte of the RCD is
		// the type and is not used in signature validation.
		for i := range invalidIssuance.ExtIDs {
			invalidIssuance.ExtIDs[i][1]++
			assert.False(invalidIssuance.ValidSignature())
			invalidIssuance.ExtIDs[i][1]--
			require.True(invalidIssuance.ValidSignature())
		}
	})
	t.Run("Valid()", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		// Start with a valid Issuance.
		validIDKey := issuerKey.RCDHash()
		require.NoError(validIssuance.Valid(validIDKey))

		// Make a working copy.
		invalidIssuance := validIssuance
		invalidIssuance.Content = append([]byte{}, validIssuance.Content...)

		// Invalid ExtIDs
		invalidIssuance.ExtIDs = nil
		assert.EqualError(invalidIssuance.Valid(validIDKey),
			"insufficient number of ExtIDs")
		invalidIssuance.ExtIDs = copyExtIDs(validIssuance.ExtIDs)
		require.NoError(invalidIssuance.Valid(validIDKey))

		// Invalid Issuer Identity Key
		assert.EqualError(validIssuance.Valid(factom.Bytes32{}), "invalid RCD")

		// Invalid JSON
		invalidIssuance.Content[0]++
		assert.EqualError(invalidIssuance.Valid(validIDKey),
			"invalid character '|' looking for beginning of value")
		invalidIssuance.Content[0]--
		require.NoError(invalidIssuance.Valid(validIDKey))

		// Invalid Data
		invalidIssuanceEntryContentMap := copyContentMap(validIssuanceEntryContentMap)
		invalidIssuanceEntryContentMap["supply"] = 0
		content := invalidIssuance.Content
		extIDs := invalidIssuance.ExtIDs
		invalidIssuance.Content = marshal(invalidIssuanceEntryContentMap)
		invalidIssuance.Sign(issuerKey)
		assert.EqualError(invalidIssuance.Valid(validIDKey),
			`invalid "supply": must be positive or -1`)
		invalidIssuance.Content = content
		invalidIssuance.ExtIDs = extIDs
		require.NoError(invalidIssuance.Valid(validIDKey))

		// Invalid Signature
		invalidIssuance.ExtIDs[1][0]++
		assert.EqualError(invalidIssuance.Valid(validIDKey), "invalid signature")
	})
}

func marshal(v map[string]interface{}) []byte {
	data, _ := json.Marshal(v)
	return data
}

func copyExtIDs(src []factom.Bytes) []factom.Bytes {
	// Create a working copy of the valid ExtIDs.
	dst := append([]factom.Bytes{}, src...)
	for i, extID := range src {
		dst[i] = append(factom.Bytes{}, extID...)
	}
	return dst
}

func copyContentMap(src map[string]interface{}) map[string]interface{} {
	dst := make(map[string]interface{})
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
