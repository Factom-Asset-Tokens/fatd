package fat0_test

import (
	"testing"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat0"
	"github.com/stretchr/testify/assert"
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
			Address = inputs[1]
		invalidTransaction.Content = marshal(invalidTransactionEntryContentMap)
		assert.Errorf(invalidTransaction.Unmarshal(), "duplicate address")
		invalidTransactionEntryContentMap["inputs"].([]addressAmount)[0].
			Address = address

		assert.NoError(validTransaction.Unmarshal())
		assert.Equal(blockheight, validTransaction.Height, "blockheight")
		assert.Equal(validTransactionEntryContentMap["salt"],
			validTransaction.Salt, "salt")
		if assert.NotNil(validTransaction.Inputs, "inputs") &&
			assert.Len(validTransaction.Inputs, len(inputs), "inputs") {
			for i, a := range inputs {
				assert.Contains(validTransaction.Inputs, a.RCDHash(),
					"inputs")
				assert.Equal(inputAmounts[i],
					validTransaction.Inputs[a.RCDHash()],
					"input amounts")
			}
		}
		if assert.NotNil(validTransaction.Outputs, "outputs") &&
			assert.Len(validTransaction.Outputs, len(outputs), "outputs") {
			for i, a := range outputs {
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
		assert.True(true)
	})
	t.Run("ValidExtIDs()", func(t *testing.T) {
		assert := assert.New(t)
		assert.True(true)
	})
	t.Run("ValidSignature()", func(t *testing.T) {
		assert := assert.New(t)
		assert.True(true)
	})
	t.Run("Valid()", func(t *testing.T) {
		assert := assert.New(t)
		assert.True(true)
	})
}
