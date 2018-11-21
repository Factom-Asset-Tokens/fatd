package fat0_test

//import (
//	"fmt"
//	"testing"
//
//	"github.com/Factom-Asset-Tokens/fatd/factom"
//	"github.com/Factom-Asset-Tokens/fatd/fat0"
//	"github.com/stretchr/testify/assert"
//	"github.com/stretchr/testify/require"
//)
//
//func TestTransaction(t *testing.T) {
//	t.Run("Unmarshal()", func(t *testing.T) {
//		assert := assert.New(t)
//		require := require.New(t)
//		var invalidTransactionEntry factom.Entry
//		invalidTransaction := fat0.NewTransaction(&invalidTransactionEntry)
//		assert.Error(invalidTransaction.Unmarshal(), "no content")
//
//		// Initialize content map to be equal to the valid map.
//		invalidTransactionEntryContentMap := make(map[string]interface{})
//		mapCopy(invalidTransactionEntryContentMap, validTransactionEntryContentMap)
//
//		invalidTransactionEntryContentMap["extra"] = "extra"
//		invalidTransaction.Content = marshal(invalidTransactionEntryContentMap)
//		assert.EqualError(invalidTransaction.Unmarshal(),
//			`json: unknown field "extra"`)
//		delete(invalidTransactionEntryContentMap, "extra")
//		invalidTransaction.Content = marshal(invalidTransactionEntryContentMap)
//		require.NoError(invalidTransaction.Unmarshal())
//
//		// Try to use an invalid value for each field.
//		var invalid = []int{0}
//		for k, v := range invalidTransactionEntryContentMap {
//			invalidTransactionEntryContentMap[k] = invalid
//			invalidTransaction.Content = marshal(
//				invalidTransactionEntryContentMap)
//			assert.Errorf(invalidTransaction.Unmarshal(),
//				"invalid type for field %#v", k)
//			invalidTransactionEntryContentMap[k] = v
//			invalidTransaction.Content = marshal(invalidTransactionEntryContentMap)
//			require.NoError(invalidTransaction.Unmarshal())
//		}
//
//		amount := invalidTransactionEntryContentMap["inputs"].([]addressAmount)[0].
//			Amount
//		invalidTransactionEntryContentMap["inputs"].([]addressAmount)[0].Amount = 0
//		invalidTransaction.Content = marshal(invalidTransactionEntryContentMap)
//		assert.Errorf(invalidTransaction.Unmarshal(), "zero amount")
//		invalidTransactionEntryContentMap["inputs"].([]addressAmount)[0].
//			Amount = amount
//		invalidTransaction.Content = marshal(invalidTransactionEntryContentMap)
//		require.NoError(invalidTransaction.Unmarshal())
//
//		address := invalidTransactionEntryContentMap["inputs"].([]addressAmount)[0].
//			Address
//		invalidTransactionEntryContentMap["inputs"].([]addressAmount)[0].
//			Address = inputAddresses[1]
//		invalidTransaction.Content = marshal(invalidTransactionEntryContentMap)
//		assert.Errorf(invalidTransaction.Unmarshal(), "duplicate address")
//		invalidTransactionEntryContentMap["inputs"].([]addressAmount)[0].
//			Address = address
//		invalidTransaction.Content = marshal(invalidTransactionEntryContentMap)
//		require.NoError(invalidTransaction.Unmarshal())
//
//		assert.NoError(validTransaction.Unmarshal())
//		assert.Equal(blockheight, validTransaction.Height, "blockheight")
//		assert.Equal(validTransactionEntryContentMap["salt"],
//			validTransaction.Salt, "salt")
//		if assert.NotNil(validTransaction.Inputs, "inputs") &&
//			assert.Len(validTransaction.Inputs, len(inputAddresses), "inputs") {
//			for i, a := range inputAddresses {
//				assert.Contains(validTransaction.Inputs, a.RCDHash(),
//					"inputs")
//				assert.Equal(inputAmounts[i],
//					validTransaction.Inputs[a.RCDHash()],
//					"input amounts")
//			}
//		}
//		if assert.NotNil(validTransaction.Outputs, "outputs") &&
//			assert.Len(validTransaction.Outputs,
//				len(outputAddresses)+1, "outputs") {
//			for i, a := range outputAddresses {
//				assert.Contains(validTransaction.Outputs, a.RCDHash(),
//					"outputs")
//				assert.Equal(outputAmounts[i],
//					validTransaction.Outputs[a.RCDHash()],
//					"output amounts")
//			}
//		}
//
//		// Metadata can be any type.
//		validTransactionEntryContentMap["metadata"] = []int{0}
//		content := validTransaction.Content
//		validTransaction.Content = marshal(validTransactionEntryContentMap)
//		assert.NoError(validTransaction.Unmarshal())
//		assert.NotNil(validTransaction.Metadata, "metadata")
//		validTransaction.Content = content
//	})
//	t.Run("ValidData()", func(t *testing.T) {
//		assert := assert.New(t)
//		require := require.New(t)
//		// Ensure that we start off with a valid transaction
//		invalidTransaction := *validTransaction
//		require.NoError(invalidTransaction.ValidData())
//
//		// Invalid Heights
//		invalidTransaction.Height = blockheight + 1
//		assert.EqualError(invalidTransaction.ValidData(), "invalid height",
//			"tx.Height > tx.Entry.Height")
//		invalidTransaction.Height = blockheight - fat0.MaxHeightDifference - 1
//		assert.EqualError(invalidTransaction.ValidData(), "invalid height",
//			"tx.Height - tx.Entry.Height > MaxHeightDifference")
//		invalidTransaction.Height = blockheight
//		require.NoError(invalidTransaction.ValidData())
//
//		// Invalid Inputs
//		inputs := invalidTransaction.Inputs
//		invalidTransaction.Inputs = nil
//		assert.EqualError(invalidTransaction.ValidData(), "no inputs")
//		invalidTransaction.Inputs = inputs
//		require.NoError(invalidTransaction.ValidData())
//
//		// Invalid Outputs
//		outputs := invalidTransaction.Outputs
//		invalidTransaction.Outputs = nil
//		assert.EqualError(invalidTransaction.ValidData(), "no outputs")
//		invalidTransaction.Outputs = outputs
//		require.NoError(invalidTransaction.ValidData())
//
//		// Unequal sums
//		for rcdHash := range invalidTransaction.Inputs {
//			invalidTransaction.Inputs[rcdHash]++
//		}
//		assert.EqualError(invalidTransaction.ValidData(),
//			"sum(inputs) != sum(outputs)")
//		for rcdHash := range invalidTransaction.Inputs {
//			invalidTransaction.Inputs[rcdHash]--
//		}
//		require.NoError(invalidTransaction.ValidData())
//
//		// Invalid coinbase inputs
//		var coinbase factom.Address
//		invalidTransaction.Inputs[coinbase.RCDHash()] = 5
//		invalidTransaction.Outputs[outputAddresses[0].RCDHash()] += 5
//		assert.EqualError(invalidTransaction.ValidData(),
//			"invalid coinbase transaction")
//		delete(invalidTransaction.Inputs, coinbase.RCDHash())
//		invalidTransaction.Outputs[outputAddresses[0].RCDHash()] -= 5
//		require.NoError(invalidTransaction.ValidData())
//
//		// Address repeated in both inputs and outputs
//		invalidTransaction.Inputs[outputAddresses[0].RCDHash()] += 5
//		invalidTransaction.Outputs[outputAddresses[0].RCDHash()] += 5
//		assert.EqualError(invalidTransaction.ValidData(),
//			fmt.Sprintf("%v appears in both inputs and outputs",
//				outputAddresses[0]))
//		delete(invalidTransaction.Inputs, outputAddresses[0].RCDHash())
//		invalidTransaction.Outputs[outputAddresses[0].RCDHash()] -= 5
//		require.NoError(invalidTransaction.ValidData())
//
//		assert.NoError(validTransaction.ValidData())
//
//		// Valid coinbase transaction
//		require.NoError(validCoinbaseTransaction.Unmarshal())
//		assert.NoError(validCoinbaseTransaction.ValidData())
//	})
//	t.Run("ValidExtIDs()", func(t *testing.T) {
//		assert := assert.New(t)
//		require := require.New(t)
//		invalidTransaction := *validTransaction
//		validExtIDs := validTransaction.ExtIDs
//		require.NoError(invalidTransaction.ValidExtIDs())
//		invalidTransaction.ExtIDs = nil
//		assert.EqualError(invalidTransaction.ValidExtIDs(),
//			"insufficient number of ExtIDs")
//		invalidTransaction.ExtIDs = append([]factom.Bytes{}, validExtIDs...)
//		require.NoError(invalidTransaction.ValidExtIDs())
//
//		invalidTransaction.ExtIDs[0] = validExtIDs[0][0 : fat0.RCDSize-1]
//		assert.EqualError(invalidTransaction.ValidExtIDs(), "invalid RCD size")
//		invalidTransaction.ExtIDs[0] = validExtIDs[0]
//		require.NoError(invalidTransaction.ValidExtIDs())
//
//		invalidTransaction.ExtIDs[0][0]++
//		assert.EqualError(invalidTransaction.ValidExtIDs(), "invalid RCD type")
//		invalidTransaction.ExtIDs[0][0]--
//		require.NoError(invalidTransaction.ValidExtIDs())
//
//		invalidTransaction.ExtIDs[1] = validExtIDs[1][0 : fat0.SignatureSize-1]
//		assert.EqualError(invalidTransaction.ValidExtIDs(), "invalid signature size")
//		invalidTransaction.ExtIDs[1] = validExtIDs[1]
//		require.NoError(invalidTransaction.ValidExtIDs())
//
//		assert.NoError(validTransaction.ValidExtIDs())
//		validTransaction.ExtIDs = append(validTransaction.ExtIDs, []byte{0})
//		assert.NoError(validTransaction.ValidExtIDs(), "additional ExtIDs")
//		validTransaction.ExtIDs = validExtIDs[0 : len(validTransaction.Inputs)*2]
//		require.NoError(validTransaction.ValidExtIDs())
//	})
//	t.Run("ValidSignatures()", func(t *testing.T) {
//		assert := assert.New(t)
//		require := require.New(t)
//		invalidTransaction := *validTransaction
//		validExtIDs := validTransaction.ExtIDs
//		invalidTransaction.ExtIDs = append([]factom.Bytes{}, validExtIDs...)
//		invalidTransaction.ExtIDs[1] = append(factom.Bytes{}, validExtIDs[1]...)
//		invalidTransaction.ExtIDs[1][0]++
//		assert.False(invalidTransaction.ValidSignatures())
//		invalidTransaction.ExtIDs[1][0]--
//		assert.True(invalidTransaction.ValidSignatures())
//		require.True(validTransaction.ValidSignatures())
//	})
//	t.Run("ValidRCDs()", func(t *testing.T) {
//		assert := assert.New(t)
//		validTransaction.ExtIDs[0][0]++
//		assert.False(validTransaction.ValidRCDs())
//		validTransaction.ExtIDs[0][0]--
//		assert.True(validTransaction.ValidRCDs())
//	})
//	t.Run("Valid()", func(t *testing.T) {
//		assert := assert.New(t)
//		require := require.New(t)
//		validIDKey := issuerKey.RCDHash()
//		invalidIDKey := coinbase.RCDHash()
//		assert.NoError(validTransaction.Valid(nil))
//		assert.NoError(validTransaction.Valid(&validIDKey))
//
//		validTransaction.Content[0] = ':'
//		assert.EqualError(validTransaction.Valid(&validIDKey),
//			"invalid character ':' looking for beginning of value")
//		validTransaction.Content[0] = '{'
//		require.NoError(validTransaction.Valid(&validIDKey))
//
//		validTransaction.Entry.Height = blockheight + 5
//		assert.EqualError(validTransaction.Valid(&validIDKey), "invalid height")
//		validTransaction.Entry.Height = blockheight
//		require.NoError(validTransaction.Valid(&validIDKey))
//
//		extIDs := validTransaction.ExtIDs
//		validTransaction.ExtIDs = nil
//		assert.EqualError(validTransaction.Valid(&validIDKey),
//			"insufficient number of ExtIDs")
//		validTransaction.ExtIDs = extIDs
//		require.NoError(validTransaction.Valid(&validIDKey))
//
//		validTransaction.ExtIDs[0][1]++
//		assert.EqualError(validTransaction.Valid(&validIDKey), "invalid RCDs")
//		validTransaction.ExtIDs[0][1]--
//		require.NoError(validTransaction.Valid(&validIDKey))
//
//		validTransaction.ExtIDs[1][1]++
//		assert.EqualError(validTransaction.Valid(&validIDKey), "invalid signatures")
//		validTransaction.ExtIDs[1][1]--
//		require.NoError(validTransaction.Valid(&validIDKey))
//
//		assert.NoError(validCoinbaseTransaction.Valid(&validIDKey))
//		assert.EqualError(validCoinbaseTransaction.Valid(&invalidIDKey),
//			"invalid RCD")
//	})
//}
