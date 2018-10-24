package fat0_test

import (
	"fmt"
	"testing"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat0"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransaction(t *testing.T) {
	t.Run("Unmarshal()", func(t *testing.T) {
		assert := assert.New(t)
		var invalidTransactionEntry factom.Entry
		invalidTransaction := fat0.NewTransaction(&invalidTransactionEntry)
		assert.Error(invalidTransaction.Unmarshal(), "no content")

		// Initialize content map to be equal to the valid map.
		invalidTransactionEntryContentMap := make(map[string]interface{})
		mapCopy(invalidTransactionEntryContentMap, validTransactionEntryContentMap)

		invalidTransactionEntryContentMap["extra"] = "extra"
		invalidTransaction.Content = marshal(invalidTransactionEntryContentMap)
		assert.Error(invalidTransaction.Unmarshal(), "extra unrecognized field")
		delete(invalidTransactionEntryContentMap, "extra")

		// Try to use an invalid value for each field.
		var invalid = []int{0}
		for k, v := range invalidTransactionEntryContentMap {
			invalidTransactionEntryContentMap[k] = invalid
			invalidTransaction.Content = marshal(invalidTransactionEntryContentMap)
			assert.Errorf(invalidTransaction.Unmarshal(),
				"invalid type for field %#v", k)
			invalidTransactionEntryContentMap[k] = v
		}

		amount := invalidTransactionEntryContentMap["inputs"].([]addressAmount)[0].
			Amount
		invalidTransactionEntryContentMap["inputs"].([]addressAmount)[0].Amount = 0
		invalidTransaction.Content = marshal(invalidTransactionEntryContentMap)
		assert.Errorf(invalidTransaction.Unmarshal(), "zero amount")
		invalidTransactionEntryContentMap["inputs"].([]addressAmount)[0].
			Amount = amount

		address := invalidTransactionEntryContentMap["inputs"].([]addressAmount)[0].
			Address
		invalidTransactionEntryContentMap["inputs"].([]addressAmount)[0].
			Address = inputAddresses[1]
		invalidTransaction.Content = marshal(invalidTransactionEntryContentMap)
		assert.Errorf(invalidTransaction.Unmarshal(), "duplicate address")
		invalidTransactionEntryContentMap["inputs"].([]addressAmount)[0].
			Address = address

		assert.NoError(validTransaction.Unmarshal())
		assert.Equal(blockheight, validTransaction.Height, "blockheight")
		assert.Equal(validTransactionEntryContentMap["salt"],
			validTransaction.Salt, "salt")
		if assert.NotNil(validTransaction.Inputs, "inputs") &&
			assert.Len(validTransaction.Inputs, len(inputAddresses), "inputs") {
			for i, a := range inputAddresses {
				assert.Contains(validTransaction.Inputs, a.RCDHash(),
					"inputs")
				assert.Equal(inputAmounts[i],
					validTransaction.Inputs[a.RCDHash()],
					"input amounts")
			}
		}
		if assert.NotNil(validTransaction.Outputs, "outputs") &&
			assert.Len(validTransaction.Outputs, len(outputAddresses), "outputs") {
			for i, a := range outputAddresses {
				assert.Contains(validTransaction.Outputs, a.RCDHash(),
					"outputs")
				assert.Equal(outputAmounts[i],
					validTransaction.Outputs[a.RCDHash()],
					"output amounts")
			}
		}

		// Metadata can be any type.
		validTransactionEntryContentMap["metadata"] = []int{0}
		content := validTransaction.Content
		validTransaction.Content = marshal(validTransactionEntryContentMap)
		assert.NoError(validTransaction.Unmarshal())
		assert.NotNil(validTransaction.Metadata, "metadata")
		validTransaction.Content = content
	})
	t.Run("ValidData()", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)
		invalidTransaction := *validTransaction

		// Invalid Heights
		invalidTransaction.Height = blockheight + 1
		assert.EqualError(invalidTransaction.ValidData(), "invalid height",
			"tx.Height > tx.Entry.Height")
		invalidTransaction.Height = blockheight - fat0.MaxHeightDifference - 1
		assert.EqualError(invalidTransaction.ValidData(), "invalid height",
			"tx.Height - tx.Entry.Height > MaxHeightDifference")
		invalidTransaction.Height = blockheight
		require.NoError(invalidTransaction.ValidData())

		// Invalid Inputs
		inputs := invalidTransaction.Inputs
		invalidTransaction.Inputs = nil
		assert.EqualError(invalidTransaction.ValidData(), "no inputs")
		invalidTransaction.Inputs = inputs
		require.NoError(invalidTransaction.ValidData())

		// Invalid Outputs
		outputs := invalidTransaction.Outputs
		invalidTransaction.Outputs = nil
		assert.EqualError(invalidTransaction.ValidData(), "no outputs")
		invalidTransaction.Outputs = outputs
		require.NoError(invalidTransaction.ValidData())

		// Unequal sums
		for rcdHash := range invalidTransaction.Inputs {
			invalidTransaction.Inputs[rcdHash]++
		}
		assert.EqualError(invalidTransaction.ValidData(),
			"sum(inputs) != sum(outputs)")
		for rcdHash := range invalidTransaction.Inputs {
			invalidTransaction.Inputs[rcdHash]--
		}
		require.NoError(invalidTransaction.ValidData())

		// Invalid coinbase inputs
		var coinbase factom.Address
		invalidTransaction.Inputs[coinbase.RCDHash()] = 5
		invalidTransaction.Outputs[outputAddresses[0].RCDHash()] += 5
		assert.EqualError(invalidTransaction.ValidData(),
			"invalid coinbase transaction")
		delete(invalidTransaction.Inputs, coinbase.RCDHash())
		invalidTransaction.Outputs[outputAddresses[0].RCDHash()] -= 5
		require.NoError(invalidTransaction.ValidData())

		// Address repeated in both inputs and outputs
		invalidTransaction.Inputs[outputAddresses[0].RCDHash()] += 5
		invalidTransaction.Outputs[outputAddresses[0].RCDHash()] += 5
		assert.EqualError(invalidTransaction.ValidData(),
			fmt.Sprintf("%v appears in both inputs and outputs",
				outputAddresses[0]))
		delete(invalidTransaction.Inputs, outputAddresses[0].RCDHash())
		invalidTransaction.Outputs[outputAddresses[0].RCDHash()] -= 5
		require.NoError(invalidTransaction.ValidData())

		assert.NoError(validTransaction.ValidData())

		// Valid coinbase transaction
		inputs = validTransaction.Inputs
		validTransaction.Inputs = fat0.AddressAmountMap{coinbase.RCDHash(): 110}
		assert.NoError(validTransaction.ValidData())
		validTransaction.Inputs = inputs
	})
	t.Run("ValidExtIDs()", func(t *testing.T) {
		assert := assert.New(t)
		assert.NoError(validTransaction.ValidExtIDs())
	})
	t.Run("ValidSignatures()", func(t *testing.T) {
		assert := assert.New(t)
		assert.True(validTransaction.ValidSignatures())
	})
	t.Run("ValidRCDs()", func(t *testing.T) {
		assert := assert.New(t)
		assert.True(validTransaction.ValidSignatures())
	})
	t.Run("Valid()", func(t *testing.T) {
		assert := assert.New(t)
		assert.True(true)
	})
}
