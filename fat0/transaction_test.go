package fat0_test

import (
	"encoding/json"
	"testing"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	. "github.com/Factom-Asset-Tokens/fatd/fat0"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var transactionTests = []struct {
	Name      string
	Error     string
	IssuerKey factom.Address
	Coinbase  bool
	Tx        Transaction
}{{
	Name: "valid",
	Tx:   validTx(),
}, {
	Name: "valid (single outputs)",
	Tx: func() Transaction {
		m := validTxEntryContentMap()
		in := inputs()
		out := outputs()
		out[0].Amount += out[1].Amount + out[2].Amount
		out = out[0:1]
		m["inputs"] = in
		m["outputs"] = out
		return transaction(marshal(m))
	}(),
}, {
	Name:      "valid (coinbase)",
	IssuerKey: issuerKey,
	Tx:        coinbaseTx(),
}, {
	Name: "valid (omit metadata)",
	Tx: func() Transaction {
		m := validTxEntryContentMap()
		delete(m, "metadata")
		return transaction(marshal(m))
	}(),
}, {
	Name:  "invalid JSON (nil)",
	Error: "EOF",
	Tx:    transaction(nil),
}, {
	Name:  "invalid JSON (unknown field)",
	Error: `json: unknown field "invalid"`,
	Tx:    transaction(factom.Bytes(`{"invalid":5}`)),
}, {
	Name:  "invalid JSON (invalid inputs type)",
	Error: "json: cannot unmarshal number into Go value of type fat0.AddressAmount",
	Tx:    invalidField("inputs"),
}, {
	Name:  "invalid JSON (invalid outputs type)",
	Error: "json: cannot unmarshal number into Go value of type fat0.AddressAmount",
	Tx:    invalidField("outputs"),
}, {
	Name:  "invalid JSON (invalid inputs, zero amount)",
	Error: "invalid amount (0) for address: FA2HaNAq1f85f1cxzywDa7etvtYCGZUztERvExzQik3CJrGBM4sx",
	Tx: func() Transaction {
		m := validTxEntryContentMap()
		in := inputs()
		in[0].Amount = 0
		m["inputs"] = in
		return transaction(marshal(m))
	}(),
}, {
	Name:  "invalid JSON (invalid inputs, duplicate)",
	Error: "duplicate address: FA2HaNAq1f85f1cxzywDa7etvtYCGZUztERvExzQik3CJrGBM4sx",
	Tx: func() Transaction {
		m := validTxEntryContentMap()
		in := inputs()
		in[1].Address = in[0].Address
		m["inputs"] = in
		return transaction(marshal(m))
	}(),
}, {
	Name:  "invalid data (no inputs)",
	Error: "no inputs",
	Tx: func() Transaction {
		m := validTxEntryContentMap()
		m["inputs"] = []AddressAmount{}
		return transaction(marshal(m))
	}(),
}, {
	Name:  "invalid data (no outputs)",
	Error: "no outputs",
	Tx: func() Transaction {
		m := validTxEntryContentMap()
		m["outputs"] = []AddressAmount{}
		return transaction(marshal(m))
	}(),
}, {
	Name:  "invalid data (sum mismatch)",
	Error: "sum(inputs) != sum(outputs)",
	Tx: func() Transaction {
		m := validTxEntryContentMap()
		out := outputs()
		out[0].Amount++
		m["outputs"] = out
		return transaction(marshal(m))
	}(),
}, {
	Name:      "invalid data (coinbase)",
	Error:     "invalid coinbase transaction",
	IssuerKey: issuerKey,
	Tx: func() Transaction {
		m := validCoinbaseTxEntryContentMap()
		in := append(coinbaseInputs(), AddressAmount{
			Address: inputAddresses[0],
			Amount:  1,
		})
		out := outputs()
		out[0].Amount++
		m["inputs"] = in
		m["outputs"] = out
		return transaction(marshal(m))
	}(),
}, {
	Name:      "invalid data (coinbase, coinbase outputs)",
	Error:     "an address appears in both the inputs and the outputs",
	IssuerKey: issuerKey,
	Tx: func() Transaction {
		m := validCoinbaseTxEntryContentMap()
		out := append(coinbaseOutputs(), AddressAmount{
			Address: coinbase,
			Amount:  1,
		})
		in := coinbaseInputs()
		in[0].Amount++
		m["inputs"] = in
		m["outputs"] = out
		return transaction(marshal(m))
	}(),
}, {
	Name:  "invalid data (inputs outputs overlap)",
	Error: "an address appears in both the inputs and the outputs",
	Tx: func() Transaction {
		m := validTxEntryContentMap()
		in := inputs()
		out := outputs()
		in[0].Address = out[0].Address
		m["inputs"] = in
		m["outputs"] = out
		return transaction(marshal(m))
	}(),
}, {
	Name:  "invalid ExtIDs (timestamp)",
	Error: "timestamp salt expired",
	Tx: func() Transaction {
		t := validTx()
		t.ExtIDs[0] = factom.Bytes("100")
		return t
	}(),
}, {
	Name:  "invalid ExtIDs (length)",
	Error: "incorrect number of ExtIDs",
	Tx: func() Transaction {
		t := validTx()
		t.ExtIDs = append(t.ExtIDs, factom.Bytes{})
		return t
	}(),
}, {
	Name:  "invalid coinbase issuer key",
	Error: "invalid RCD",
	Tx:    coinbaseTx(),
}, {
	Name:  "RCD input mismatch",
	Error: "invalid RCDs",
	Tx: func() Transaction {
		t := validTx()
		t.Sign(twoAddresses()...)
		return t
	}(),
}}

func TestTransaction(t *testing.T) {
	for _, test := range transactionTests {
		t.Run(test.Name, func(t *testing.T) {
			assert := assert.New(t)
			tx := test.Tx
			key := test.IssuerKey
			err := tx.Valid(key.RCDHash())
			if len(test.Error) != 0 {
				assert.EqualError(err, test.Error)
				return
			}
			require.NoError(t, err)
			if test.Coinbase {
				assert.True(tx.IsCoinbase())
			}
		})
	}
}

var (
	coinbase factom.Address

	inputAddresses  = twoAddresses()
	outputAddresses = append(twoAddresses(), coinbase)

	inputAmounts  = []uint64{100, 10}
	outputAmounts = []uint64{90, 10, 10}

	coinbaseInputAddresses  = []factom.Address{coinbase}
	coinbaseOutputAddresses = twoAddresses()

	coinbaseInputAmounts  = []uint64{110}
	coinbaseOutputAmounts = []uint64{90, 20}

	tokenChainID = ChainID("test", identityChainID)
)

// Transactions
func validTx() Transaction {
	return transaction(marshal(validTxEntryContentMap()))
}
func coinbaseTx() Transaction {
	t := transaction(marshal(validCoinbaseTxEntryContentMap()))
	t.Sign(issuerKey)
	return t
}
func transaction(content factom.Bytes) Transaction {
	e := factom.Entry{
		ChainID: tokenChainID,
		Content: content,
	}
	t := NewTransaction(e)
	t.Sign(inputAddresses...)
	return t
}
func invalidField(field string) Transaction {
	m := validTxEntryContentMap()
	m[field] = []int{0}
	return transaction(marshal(m))
}

// Content maps
func validTxEntryContentMap() map[string]interface{} {
	return map[string]interface{}{
		"inputs":   inputs(),
		"outputs":  outputs(),
		"metadata": []int{0},
	}
}
func validCoinbaseTxEntryContentMap() map[string]interface{} {
	return map[string]interface{}{
		"inputs":   coinbaseInputs(),
		"outputs":  coinbaseOutputs(),
		"metadata": []int{0},
	}
}

// inputs/outputs
func inputs() []AddressAmount {
	inputs := []AddressAmount{}
	for i := range inputAddresses {
		inputs = append(inputs, AddressAmount{
			Address: inputAddresses[i],
			Amount:  inputAmounts[i],
		})
	}
	return inputs
}
func outputs() []AddressAmount {
	outputs := []AddressAmount{}
	for i := range outputAddresses {
		outputs = append(outputs, AddressAmount{
			Address: outputAddresses[i],
			Amount:  outputAmounts[i],
		})
	}
	return outputs
}
func coinbaseInputs() []AddressAmount {
	inputs := []AddressAmount{}
	for i := range coinbaseInputAddresses {
		inputs = append(inputs, AddressAmount{
			Address: coinbaseInputAddresses[i],
			Amount:  coinbaseInputAmounts[i],
		})
	}
	return inputs
}
func coinbaseOutputs() []AddressAmount {
	outputs := []AddressAmount{}
	for i := range coinbaseOutputAddresses {
		outputs = append(outputs, AddressAmount{
			Address: coinbaseOutputAddresses[i],
			Amount:  coinbaseOutputAmounts[i],
		})
	}
	return outputs
}

var transactionMarshalEntryTests = []struct {
	Name  string
	Error string
	Tx    Transaction
}{{
	Name: "valid",
	Tx:   newTransaction(),
}, {
	Name: "valid (omit zero balances)",
	Tx: func() Transaction {
		t := newTransaction()
		t.Inputs[coinbase.RCDHash()] = 0
		return t
	}(),
}, {
	Name: "valid (metadata)",
	Tx: func() Transaction {
		t := newTransaction()
		t.Metadata = json.RawMessage(`{"memo":"Rent for Dec 2018"}`)
		return t
	}(),
}, {
	Name:  "invalid data",
	Error: "sum(inputs) != sum(outputs)",
	Tx: func() Transaction {
		t := newTransaction()
		t.Inputs[inputAddresses[0].RCDHash()]++
		return t
	}(),
}, {
	Name:  "invalid metadata JSON",
	Error: "json: error calling MarshalJSON for type json.RawMessage: invalid character 'a' looking for beginning of object key string",
	Tx: func() Transaction {
		t := newTransaction()
		t.Metadata = json.RawMessage("{asdf")
		return t
	}(),
}}

func TestTransactionMarshalEntry(t *testing.T) {
	for _, test := range transactionMarshalEntryTests {
		t.Run(test.Name, func(t *testing.T) {
			assert := assert.New(t)
			tx := test.Tx
			err := tx.MarshalEntry()
			if len(test.Error) == 0 {
				assert.NoError(err)
			} else {
				assert.EqualError(err, test.Error)
			}
		})
	}
}

func newTransaction() Transaction {
	return Transaction{
		Inputs:  inputAddressAmountMap(),
		Outputs: outputAddressAmountMap(),
	}
}
func inputAddressAmountMap() AddressAmountMap {
	return addressAmountMap(inputs())
}
func outputAddressAmountMap() AddressAmountMap {
	return addressAmountMap(outputs())
}
func addressAmountMap(aas []AddressAmount) AddressAmountMap {
	m := make(AddressAmountMap)
	for _, aa := range aas {
		m[aa.Address.RCDHash()] = aa.Amount
	}
	return m

}
