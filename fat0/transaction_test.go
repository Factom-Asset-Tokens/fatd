package fat0_test

import (
	"encoding/json"
	"fmt"
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
		out := outputs()
		out[outputAddresses[0].String()] += out[outputAddresses[1].String()] +
			out[outputAddresses[2].String()]
		delete(out, outputAddresses[1].String())
		delete(out, outputAddresses[2].String())
		return setFieldTransaction("outputs", out)
	}(),
}, {
	Name:      "valid (coinbase)",
	IssuerKey: issuerKey,
	Tx:        coinbaseTx(),
}, {
	Name: "valid (omit metadata)",
	Tx:   omitFieldTransaction("metadata"),
}, {
	Name:  "invalid JSON (nil)",
	Error: "not a single valid JSON",
	Tx:    transaction(nil),
}, {
	Name:  "invalid JSON (unknown field)",
	Error: `json: unknown field "invalid"`,
	Tx:    transaction(factom.Bytes(`{"invalid":5}`)),
}, {
	Name:  "invalid JSON (invalid inputs type)",
	Error: "json: cannot unmarshal array into Go value of type map[string]uint64",
	Tx:    invalidField("inputs"),
}, {
	Name:  "invalid JSON (invalid outputs type)",
	Error: "json: cannot unmarshal array into Go value of type map[string]uint64",
	Tx:    invalidField("outputs"),
}, {
	Name:  "invalid JSON (invalid inputs, zero amount)",
	Error: "*fat0.AddressAmountMap: invalid amount (0) for address: FA2HaNAq1f85f1cxzywDa7etvtYCGZUztERvExzQik3CJrGBM4sx",
	Tx: func() Transaction {
		in := inputs()
		in[inputAddresses[0].String()] = 0
		return setFieldTransaction("inputs", in)
	}(),
}, {
	Name:  "invalid JSON (invalid inputs, duplicate)",
	Error: "contentJSONLen (389) != expectedJSONLen (330)",
	Tx:    transaction([]byte(`{"inputs":{"FA2HaNAq1f85f1cxzywDa7etvtYCGZUztERvExzQik3CJrGBM4sx":100,"FA2HaNAq1f85f1cxzywDa7etvtYCGZUztERvExzQik3CJrGBM4sx":100,"FA3rCRnpU95ieYCwh7YGH99YUWPjdVEjk73mpjqnVpTDt3rUUhX8":10},"metadata":[0],"outputs":{"FA1zT4aFpEvcnPqPCigB3fvGu4Q4mTXY22iiuV69DqE1pNhdF2MC":10,"FA2PJRLbuVDyAKire9BRnJYkh2NZc2Fjco4FCrPtXued7F26wGBP":90,"FA2uyZviB3vs28VkqkfnhoXRD8XdKP1zaq7iukq2gBfCq3hxeuE8":10}}`)),
}, {
	Name:  "invalid JSON (two objects)",
	Error: "not a single valid JSON",
	Tx:    transaction([]byte(`{"inputs":{"FA2HaNAq1f85f1cxzywDa7etvtYCGZUztERvExzQik3CJrGBM4sx":100,"FA3rCRnpU95ieYCwh7YGH99YUWPjdVEjk73mpjqnVpTDt3rUUhX8":10},"metadata":[0],"outputs":{"FA1zT4aFpEvcnPqPCigB3fvGu4Q4mTXY22iiuV69DqE1pNhdF2MC":10,"FA2PJRLbuVDyAKire9BRnJYkh2NZc2Fjco4FCrPtXued7F26wGBP":90,"FA2uyZviB3vs28VkqkfnhoXRD8XdKP1zaq7iukq2gBfCq3hxeuE8":10}}{}`)),
}, {
	Name:  "invalid data (no inputs)",
	Error: "no inputs",
	Tx:    setFieldTransaction("inputs", AddressAmountMap{}),
}, {
	Name:  "invalid data (no outputs)",
	Error: "no outputs",
	Tx:    setFieldTransaction("outputs", AddressAmountMap{}),
}, {
	Name:  "invalid data (omit inputs)",
	Error: "contentJSONLen (202) != expectedJSONLen (214)",
	Tx:    omitFieldTransaction("inputs"),
}, {
	Name:  "invalid data (omit outputs)",
	Error: "contentJSONLen (144) != expectedJSONLen (157)",
	Tx:    omitFieldTransaction("outputs"),
}, {
	Name:  "invalid data (sum mismatch)",
	Error: "sum(inputs) != sum(outputs)",
	Tx: func() Transaction {
		out := outputs()
		out[outputAddresses[0].String()]++
		return setFieldTransaction("outputs", out)
	}(),
}, {
	Name:      "invalid data (coinbase)",
	Error:     "invalid coinbase transaction",
	IssuerKey: issuerKey,
	Tx: func() Transaction {
		m := validCoinbaseTxEntryContentMap()
		in := coinbaseInputs()
		in[inputAddresses[0].String()] = 1
		out := coinbaseOutputs()
		out[outputAddresses[0].String()]++
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
		in := coinbaseInputs()
		out := coinbaseOutputs()
		in[coinbase.String()]++
		out[coinbase.String()]++
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
		in[outputAddresses[0].String()] = in[inputAddresses[0].String()]
		delete(in, inputAddresses[0].String())
		m["inputs"] = in
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
func omitFieldTransaction(field string) Transaction {
	m := validTxEntryContentMap()
	delete(m, field)
	return transaction(marshal(m))
}
func setFieldTransaction(field string, value interface{}) Transaction {
	m := validTxEntryContentMap()
	m[field] = value
	return transaction(marshal(m))
}
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
		ChainID: &tokenChainID,
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
func inputs() map[string]uint64 {
	inputs := map[string]uint64{}
	for i := range inputAddresses {
		inputs[inputAddresses[i].String()] = inputAmounts[i]
	}
	return inputs
}
func outputs() map[string]uint64 {
	outputs := map[string]uint64{}
	for i := range outputAddresses {
		outputs[outputAddresses[i].String()] = outputAmounts[i]
	}
	return outputs
}
func coinbaseInputs() map[string]uint64 {
	inputs := map[string]uint64{}
	for i := range coinbaseInputAddresses {
		inputs[coinbaseInputAddresses[i].String()] = coinbaseInputAmounts[i]
	}
	return inputs
}
func coinbaseOutputs() map[string]uint64 {
	outputs := map[string]uint64{}
	for i := range coinbaseOutputAddresses {
		outputs[coinbaseOutputAddresses[i].String()] = coinbaseOutputAmounts[i]
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
		t.Inputs[*coinbase.RCDHash()] = 0
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
		t.Inputs[*inputAddresses[0].RCDHash()]++
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
func addressAmountMap(aas map[string]uint64) AddressAmountMap {
	m := make(AddressAmountMap)
	for addressStr, amount := range aas {
		a := factom.Address{}
		if err := json.Unmarshal(
			[]byte(fmt.Sprintf("%#v", addressStr)), &a); err != nil {
			panic(err)
		}
		m[*a.RCDHash()] = amount
	}
	return m
}
