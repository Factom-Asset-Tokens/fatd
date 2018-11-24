package fat0_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat0"
	"github.com/FactomProject/ed25519"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func twoAddresses() []factom.Address {
	adrs := make([]factom.Address, 2)
	for i := range adrs {
		adrs[i].PublicKey, adrs[i].PrivateKey, _ = ed25519.GenerateKey(randSource)
	}
	return adrs
}

func newTx(content map[string]interface{}, adrs ...factom.Address) fat0.Transaction {
	e := factom.Entry{ChainID: tokenChainID, Content: marshal(content)}
	t := fat0.NewTransaction(e)
	t.Sign(adrs...)
	return t
}

var (
	coinbase factom.Address

	inputAddresses  = twoAddresses()
	outputAddresses = twoAddresses()

	inputAmounts  = []uint64{100, 10}
	outputAmounts = []uint64{90, 10, 10}

	validTxEntryContentMap = map[string]interface{}{
		"inputs": map[string]uint64{
			inputAddresses[0].String(): inputAmounts[0],
			inputAddresses[1].String(): inputAmounts[1],
		},
		"outputs": map[string]uint64{
			outputAddresses[0].String(): outputAmounts[0],
			outputAddresses[1].String(): outputAmounts[1],
			coinbase.String():           outputAmounts[2],
		},
		"salt":     "xyz",
		"metadata": []int{0},
	}
	validCoinbaseTxEntryContentMap = map[string]interface{}{
		"inputs": map[string]uint64{
			coinbase.String(): inputAmounts[0] + inputAmounts[1],
		},
		"outputs": map[string]uint64{
			outputAddresses[0].String(): outputAmounts[0],
			outputAddresses[1].String(): outputAmounts[1] + outputAmounts[2],
		},
		"salt":     "abc",
		"metadata": []int{0},
	}

	tokenChainID    = fat0.ChainID("test", identityChainID)
	validTx         = newTx(validTxEntryContentMap, inputAddresses...)
	validCoinbaseTx = newTx(validCoinbaseTxEntryContentMap, issuerKey)
)

func TestTransaction(t *testing.T) {
	t.Run("UnmarshalEntry()", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		// Start with a valid Transaction.
		require.NoError(validTx.UnmarshalEntry())
		assert.Equal(validTxEntryContentMap["salt"], validTx.Salt, "salt")
		assert.NotNil(validTx.Metadata, "metadata")
		if assert.NotNil(validTx.Inputs, "inputs") &&
			assert.Len(validTx.Inputs, len(inputAddresses), "inputs") {
			for i, a := range inputAddresses {
				assert.Contains(validTx.Inputs, a.RCDHash(), "inputs")
				assert.Equal(inputAmounts[i], validTx.Inputs[a.RCDHash()],
					"input amounts")
			}
		}
		if assert.NotNil(validTx.Outputs, "outputs") &&
			assert.Len(validTx.Outputs, len(outputAddresses)+1, "outputs") {
			for i, a := range outputAddresses {
				assert.Contains(validTx.Outputs, a.RCDHash(), "outputs")
				assert.Equal(outputAmounts[i], validTx.Outputs[a.RCDHash()],
					"output amounts")
			}
		}

		// Create a working copy.
		invalidTx := validTx
		invalidTxEntryContentMap := copyContentMap(validTxEntryContentMap)

		// An invalid field should cause an error.
		invalidField := "invalid"
		invalidTxEntryContentMap[invalidField] = invalidField
		invalidTx.Content = marshal(invalidTxEntryContentMap)
		assert.EqualError(invalidTx.UnmarshalEntry(),
			fmt.Sprintf("json: unknown field %#v", invalidField))
		delete(invalidTxEntryContentMap, invalidField)
		invalidTx.Content = marshal(invalidTxEntryContentMap)
		require.NoError(invalidTx.UnmarshalEntry())

		// Try to use an invalid value for each field, except for
		// "metadata".
		invalidValue := []int{0}
		delete(invalidTxEntryContentMap, "metadata")
		for k, v := range invalidTxEntryContentMap {
			invalidTxEntryContentMap[k] = invalidValue
			invalidTx.Content = marshal(invalidTxEntryContentMap)
			assert.Errorf(invalidTx.UnmarshalEntry(),
				"invalid type for field %#v", k)
			invalidTxEntryContentMap[k] = v
			invalidTx.Content = marshal(invalidTxEntryContentMap)
			require.NoError(invalidTx.UnmarshalEntry())
		}

		// Zero amounts cannot be unmarshaled.
		invalidTxEntryContentMap["inputs"].(map[string]uint64)[inputAddresses[0].
			String()] = 0
		invalidTx.Content = marshal(invalidTxEntryContentMap)
		assert.Errorf(invalidTx.UnmarshalEntry(), "zero amount")
		invalidTxEntryContentMap["inputs"].(map[string]uint64)[inputAddresses[0].
			String()] = inputAmounts[0]
		invalidTx.Content = marshal(invalidTxEntryContentMap)
		require.NoError(invalidTx.UnmarshalEntry())

		// Duplicate addresses cannot be unmarshaled.
		inputs := invalidTxEntryContentMap["inputs"]
		invalidTxEntryContentMap["inputs"] = json.RawMessage(fmt.Sprintf(
			"{%#v:%v,%#v:%v,%#v:%v}",
			inputAddresses[0], inputAmounts[0],
			inputAddresses[1], inputAmounts[1]-1,
			inputAddresses[1], 1,
		))
		invalidTx.Content = marshal(invalidTxEntryContentMap)
		assert.Errorf(invalidTx.UnmarshalEntry(), "duplicate address")
		invalidTxEntryContentMap["inputs"] = inputs
		invalidTx.Content = marshal(invalidTxEntryContentMap)
		require.NoError(invalidTx.UnmarshalEntry())
	})
	t.Run("ValidData()", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		// Validate the valid Coinbase Transaction.
		require.NoError(validCoinbaseTx.UnmarshalEntry())
		require.NoError(validCoinbaseTx.ValidData())

		// Start off with a valid transaction and make a working copy.
		require.NoError(validTx.UnmarshalEntry())
		require.NoError(validTx.ValidData())
		invalidTx := validTx

		// Invalid Inputs
		invalidTx.Inputs = nil
		assert.EqualError(invalidTx.ValidData(), "no inputs")
		invalidTx.Inputs = copyAddressAmountMap(validTx.Inputs)
		require.NoError(invalidTx.ValidData())

		// Invalid Outputs
		invalidTx.Outputs = nil
		assert.EqualError(invalidTx.ValidData(), "no outputs")
		invalidTx.Outputs = copyAddressAmountMap(validTx.Outputs)
		require.NoError(invalidTx.ValidData())

		// Unequal sums
		for rcdHash := range invalidTx.Inputs {
			invalidTx.Inputs[rcdHash]++
		}
		assert.EqualError(invalidTx.ValidData(), "sum(inputs) != sum(outputs)")
		for rcdHash := range invalidTx.Inputs {
			invalidTx.Inputs[rcdHash]--
		}
		require.NoError(invalidTx.ValidData())

		// Invalid coinbase inputs
		var coinbase factom.Address
		invalidTx.Inputs[coinbase.RCDHash()] = 5
		invalidTx.Outputs[outputAddresses[0].RCDHash()] += 5
		assert.EqualError(invalidTx.ValidData(), "invalid coinbase transaction")
		delete(invalidTx.Inputs, coinbase.RCDHash())
		invalidTx.Outputs[outputAddresses[0].RCDHash()] -= 5
		require.NoError(invalidTx.ValidData())

		// Address repeated in both inputs and outputs
		invalidTx.Inputs[outputAddresses[0].RCDHash()] += 5
		invalidTx.Outputs[outputAddresses[0].RCDHash()] += 5
		assert.EqualError(invalidTx.ValidData(),
			"an address appears in both the inputs and the outputs")
		// An address with amount zero should be ignored.
		invalidTx.Inputs[outputAddresses[0].RCDHash()] = 0
		invalidTx.Outputs[outputAddresses[0].RCDHash()] -= 5
		assert.NoError(invalidTx.ValidData())

		// Validate a tx where there are more inputs than outputs.
		delete(invalidTx.Outputs, coinbase.RCDHash())
		delete(invalidTx.Outputs, outputAddresses[1].RCDHash())
		invalidTx.Outputs[outputAddresses[0].RCDHash()] +=
			outputAmounts[1] + outputAmounts[2]
		assert.NoError(invalidTx.ValidData())
	})
	t.Run("ValidExtIDs()", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		// Start off with a valid transaction and make a working copy.
		require.NoError(validTx.UnmarshalEntry())
		require.NoError(validTx.ValidExtIDs(), fmt.Sprintf("%+v", validTx))
		invalidTx := validTx
		extIDs := copyExtIDs(validTx.ExtIDs)

		// Bad ExtIDs length.
		invalidTx.ExtIDs = nil
		assert.EqualError(invalidTx.ValidExtIDs(), "invalid number of ExtIDs")
		invalidTx.ExtIDs = extIDs
		require.NoError(invalidTx.ValidExtIDs())

		// Additional ExtIDs are not allowed.
		invalidTx.ExtIDs = append(extIDs, []byte{0})
		assert.EqualError(invalidTx.ValidExtIDs(), "invalid number of ExtIDs")
		invalidTx.ExtIDs = extIDs

		// Bad RCD length.
		rcd := extIDs[0]
		invalidTx.ExtIDs[0] = nil
		assert.EqualError(invalidTx.ValidExtIDs(), "invalid RCD size")
		invalidTx.ExtIDs[0] = rcd
		require.NoError(invalidTx.ValidExtIDs())

		// Invalid RCD Type.
		invalidTx.ExtIDs[0][0]++
		assert.EqualError(invalidTx.ValidExtIDs(), "invalid RCD type")
		invalidTx.ExtIDs[0][0]--
		require.NoError(invalidTx.ValidExtIDs())

		// Invalid Signature length.
		sig := extIDs[1]
		invalidTx.ExtIDs[1] = nil
		assert.EqualError(invalidTx.ValidExtIDs(), "invalid signature size")
		invalidTx.ExtIDs[1] = sig
		require.NoError(invalidTx.ValidExtIDs())
	})
	t.Run("ValidSignatures()", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		// Make a working copy.
		require.True(validCoinbaseTx.ValidSignatures())
		require.True(validTx.ValidSignatures())
		invalidTx := validTx
		invalidTx.ExtIDs = copyExtIDs(validTx.ExtIDs)

		// Mucking with the second byte in the RCD and the signature
		// should make the signature fail.
		// We use the second byte because the first byte of the RCD is
		// the type and is not used in signature validation.
		for i := range invalidTx.ExtIDs {
			invalidTx.ExtIDs[i][1]++
			assert.False(invalidTx.ValidSignatures())
			invalidTx.ExtIDs[i][1]--
			require.True(invalidTx.ValidSignatures())
		}
	})
	t.Run("ValidRCDs()", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		// Make a working copy.
		require.True(validTx.ValidRCDs())
		invalidTx := validTx
		invalidTx.ExtIDs = copyExtIDs(validTx.ExtIDs)
		invalidTx.Inputs = copyAddressAmountMap(validTx.Inputs)

		// Mucking with any byte in the RCD should make the RCD not
		// match the address.
		for i := 0; i < len(invalidTx.Inputs); i++ {
			invalidTx.ExtIDs[i*2][0]++
			assert.False(invalidTx.ValidRCDs(), "missing RCD")
			invalidTx.ExtIDs[i*2][0]--
			require.True(invalidTx.ValidRCDs())
		}

		// Adding an input address should fail.
		invalidTx.Inputs[coinbase.RCDHash()] = 5
		assert.False(invalidTx.ValidRCDs(), "extra address")
	})
	t.Run("Valid()", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		validIDKey := issuerKey.RCDHash()
		invalidIDKey := factom.Bytes32{}

		// Start with a valid tx.
		require.NoError(validTx.Valid(validIDKey))
		require.NoError(validCoinbaseTx.Valid(validIDKey))

		// Make a working copy.
		invalidTx := validTx
		invalidTx.Content = append([]byte{}, validTx.Content...)

		// Coinbase tx should not be valid with the wrong issuer identity key.
		assert.EqualError(validCoinbaseTx.Valid(invalidIDKey), "invalid RCD")

		// Invalid JSON
		invalidTx.Content[0]++
		assert.EqualError(invalidTx.Valid(validIDKey),
			"invalid character '|' looking for beginning of value")
		invalidTx.Content[0]--
		require.NoError(invalidTx.Valid(validIDKey))

		// Invalid Data
		content := invalidTx.Content
		invalidTxEntryContentMap := copyContentMap(validTxEntryContentMap)
		delete(invalidTxEntryContentMap, "inputs")
		invalidTx.Content = marshal(invalidTxEntryContentMap)
		invalidTx.Inputs = nil
		assert.EqualError(invalidTx.Valid(validIDKey), "no inputs")
		invalidTx.Content = content
		require.NoError(invalidTx.Valid(validIDKey))

		// Invalid ExtIDs
		extIDs := invalidTx.ExtIDs
		invalidTx.ExtIDs = nil
		assert.EqualError(invalidTx.Valid(validIDKey), "invalid number of ExtIDs")
		invalidTx.ExtIDs = extIDs
		require.NoError(invalidTx.Valid(validIDKey))

		// Invalid RCD
		invalidTx.ExtIDs[0][1]++
		assert.EqualError(invalidTx.Valid(validIDKey), "invalid RCDs")
		invalidTx.ExtIDs[0][1]--
		require.NoError(invalidTx.Valid(validIDKey))

		// Invalid signature
		invalidTx.ExtIDs[1][1]++
		assert.EqualError(invalidTx.Valid(validIDKey), "invalid signatures")

	})
}

func TestAddressAmountMap(t *testing.T) {
	t.Run("MarshalJSON()", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)
		amount := uint64(5)
		aam := fat0.AddressAmountMap{
			inputAddresses[0].RCDHash(): amount,
			coinbase.RCDHash():          0,
		}
		expectedData := fmt.Sprintf(`{%#v:%v}`, inputAddresses[0].String(), amount)

		data, err := aam.MarshalJSON()
		require.NoError(err)
		assert.JSONEq(string(expectedData), string(data))
	})
}

func copyAddressAmountMap(src fat0.AddressAmountMap) fat0.AddressAmountMap {
	dst := make(fat0.AddressAmountMap)
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
