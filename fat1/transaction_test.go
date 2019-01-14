package fat1_test

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat0"
	. "github.com/Factom-Asset-Tokens/fatd/fat1"
	"github.com/FactomProject/ed25519"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var transactionTests = []struct {
	Name      string
	Error     string
	ErrorOr   string
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
		out[outputAddresses[0].String()].Append(out[outputAddresses[1].String()])
		out[outputAddresses[0].String()].Append(out[outputAddresses[2].String()])
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
	Error: "unexpected end of JSON input",
	Tx:    transaction(nil),
}, {
	Name:  "invalid JSON (unknown field)",
	Error: `*fat1.Transaction: unexpected JSON length`,
	Tx:    setFieldTransaction("invalid", 5),
}, {
	Name:  "invalid JSON (invalid inputs type)",
	Error: "*fat1.Transaction.Inputs: *fat1.AddressNFTokensMap: json: cannot unmarshal array into Go value of type map[string]json.RawMessage",
	Tx:    invalidField("inputs"),
}, {
	Name:  "invalid JSON (invalid outputs type)",
	Error: "*fat1.Transaction.Outputs: *fat1.AddressNFTokensMap: json: cannot unmarshal array into Go value of type map[string]json.RawMessage",
	Tx:    invalidField("outputs"),
}, {
	Name:  "invalid JSON (invalid inputs, duplicate address)",
	Error: "*fat1.Transaction.Inputs: *fat1.AddressNFTokensMap: unexpected JSON length",
	Tx:    transaction([]byte(`{"inputs":{"FA2HaNAq1f85f1cxzywDa7etvtYCGZUztERvExzQik3CJrGBM4sx":[0],"FA2HaNAq1f85f1cxzywDa7etvtYCGZUztERvExzQik3CJrGBM4sx":[1],"FA3rCRnpU95ieYCwh7YGH99YUWPjdVEjk73mpjqnVpTDt3rUUhX8":[2]},"metadata":[0],"outputs":{"FA1zT4aFpEvcnPqPCigB3fvGu4Q4mTXY22iiuV69DqE1pNhdF2MC":[1],"FA2PJRLbuVDyAKire9BRnJYkh2NZc2Fjco4FCrPtXued7F26wGBP":[0],"FA2uyZviB3vs28VkqkfnhoXRD8XdKP1zaq7iukq2gBfCq3hxeuE8":[2]}}`)),
}, {
	Name:    "invalid JSON (invalid inputs, duplicate ids)",
	Error:   "*fat1.Transaction.Inputs: *fat1.AddressNFTokensMap: FA3rCRnpU95ieYCwh7YGH99YUWPjdVEjk73mpjqnVpTDt3rUUhX8 and FA2HaNAq1f85f1cxzywDa7etvtYCGZUztERvExzQik3CJrGBM4sx: duplicate NFTokenID: 0",
	ErrorOr: "*fat1.Transaction.Inputs: *fat1.AddressNFTokensMap: FA2HaNAq1f85f1cxzywDa7etvtYCGZUztERvExzQik3CJrGBM4sx and FA3rCRnpU95ieYCwh7YGH99YUWPjdVEjk73mpjqnVpTDt3rUUhX8: duplicate NFTokenID: 0",
	Tx:      transaction([]byte(`{"inputs":{"FA2HaNAq1f85f1cxzywDa7etvtYCGZUztERvExzQik3CJrGBM4sx":[0],"FA3rCRnpU95ieYCwh7YGH99YUWPjdVEjk73mpjqnVpTDt3rUUhX8":[0,1,2]},"metadata":[0],"outputs":{"FA1zT4aFpEvcnPqPCigB3fvGu4Q4mTXY22iiuV69DqE1pNhdF2MC":[1],"FA2PJRLbuVDyAKire9BRnJYkh2NZc2Fjco4FCrPtXued7F26wGBP":[0],"FA2uyZviB3vs28VkqkfnhoXRD8XdKP1zaq7iukq2gBfCq3hxeuE8":[2]}}`)),
}, {
	Name:  "invalid JSON (two objects)",
	Error: "invalid character '{' after top-level value",
	Tx:    transaction([]byte(`{"inputs":{"FA2HaNAq1f85f1cxzywDa7etvtYCGZUztERvExzQik3CJrGBM4sx":100,"FA3rCRnpU95ieYCwh7YGH99YUWPjdVEjk73mpjqnVpTDt3rUUhX8":10},"metadata":[0],"outputs":{"FA1zT4aFpEvcnPqPCigB3fvGu4Q4mTXY22iiuV69DqE1pNhdF2MC":10,"FA2PJRLbuVDyAKire9BRnJYkh2NZc2Fjco4FCrPtXued7F26wGBP":90,"FA2uyZviB3vs28VkqkfnhoXRD8XdKP1zaq7iukq2gBfCq3hxeuE8":10}}{}`)),
}, {
	Name:  "invalid data (no inputs)",
	Error: "*fat1.Transaction.Inputs: *fat1.AddressNFTokensMap: empty",
	Tx:    setFieldTransaction("inputs", json.RawMessage(`{}`)),
}, {
	Name:  "invalid data (no outputs)",
	Error: "*fat1.Transaction.Outputs: *fat1.AddressNFTokensMap: empty",
	Tx:    setFieldTransaction("outputs", json.RawMessage(`{}`)),
}, {
	Name:  "invalid data (omit inputs)",
	Error: "*fat1.Transaction.Inputs: *fat1.AddressNFTokensMap: unexpected end of JSON input",
	Tx:    omitFieldTransaction("inputs"),
}, {
	Name:  "invalid data (omit outputs)",
	Error: "*fat1.Transaction.Outputs: *fat1.AddressNFTokensMap: unexpected end of JSON input",
	Tx:    omitFieldTransaction("outputs"),
}, {
	Name:  "invalid data (Input Output mismatch)",
	Error: "*fat1.Transaction: Inputs and Outputs mismatch: number of NFTokenIDs differ",
	Tx: func() Transaction {
		out := outputs()
		NFTokenID(1000).Set(out[outputAddresses[0].String()])
		return setFieldTransaction("outputs", out)
	}(),
}, {
	Name:  "invalid data (Input Output mismatch)",
	Error: "*fat1.Transaction: Inputs and Outputs mismatch: missing NFTokenID: 1000",
	Tx: func() Transaction {
		in := inputs()
		NFTokenID(1001).Set(in[inputAddresses[0].String()])
		out := outputs()
		NFTokenID(1000).Set(out[outputAddresses[0].String()])
		m := validTxEntryContentMap()
		m["inputs"] = in
		m["outputs"] = out
		return transaction(marshal(m))
	}(),
}, {
	Name:      "invalid data (coinbase)",
	Error:     "*fat1.Transaction: invalid coinbase transaction",
	IssuerKey: issuerKey,
	Tx: func() Transaction {
		m := validCoinbaseTxEntryContentMap()
		in := coinbaseInputs()
		in[inputAddresses[0].String()] = newNFTokens(NFTokenID(1000))
		out := coinbaseOutputs()
		out[outputAddresses[0].String()] = newNFTokens(NFTokenID(1000))
		m["inputs"] = in
		m["outputs"] = out
		return transaction(marshal(m))
	}(),
}, {
	Name:      "invalid data (coinbase, coinbase outputs)",
	Error:     "*fat1.Transaction: Inputs and Outputs intersect: duplicate Address: FA1zT4aFpEvcnPqPCigB3fvGu4Q4mTXY22iiuV69DqE1pNhdF2MC",
	IssuerKey: issuerKey,
	Tx: func() Transaction {
		m := validCoinbaseTxEntryContentMap()
		in := coinbaseInputs()
		out := coinbaseOutputs()
		in[coinbase.String()] = newNFTokens(NFTokenID(1000))
		out[coinbase.String()] = newNFTokens(NFTokenID(1000))
		m["inputs"] = in
		m["outputs"] = out
		delete(m, "tokenmetadata")
		return transaction(marshal(m))
	}(),
}, {
	Name:      "invalid data (coinbase, tokenmetadata)",
	Error:     "*fat1.Transaction.TokenMetadata: too many NFTokenIDs",
	IssuerKey: issuerKey,
	Tx: func() Transaction {
		m := validCoinbaseTxEntryContentMap()
		in := coinbaseInputs()
		delete(in[coinbase.String()], NFTokenID(0))
		out := coinbaseOutputs()
		delete(out[coinbaseOutputAddresses[0].String()], NFTokenID(0))

		m["inputs"] = in
		m["outputs"] = out
		return transaction(marshal(m))
	}(),
}, {
	Name:  "invalid data (inputs outputs overlap)",
	Error: "*fat1.Transaction: Inputs and Outputs intersect: duplicate Address: FA3eYH5qH7mxtWpLp9k4aSw1tJiLkp171tKnbM9BW14MVLdiDviB",
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
	Error: "invalid number of ExtIDs",
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
				assert.Truef((err.Error() == test.Error ||
					err.Error() == test.ErrorOr),
					"%v\n%v", string(test.Tx.Content), err.Error())
				return
			}
			require.NoError(t, err, string(test.Tx.Content))
			if test.Coinbase {
				assert.True(tx.IsCoinbase(), string(test.Tx.Content))
			}
		})
	}
}

var (
	coinbase factom.Address

	inputAddresses  = twoAddresses()
	outputAddresses = append(twoAddresses(), coinbase)

	inputNFTokens = []NFTokens{newNFTokens(NewNFTokenIDRange(0, 10)),
		newNFTokens(NFTokenID(11))}
	outputNFTokens = []NFTokens{newNFTokens(NewNFTokenIDRange(0, 5)),
		newNFTokens(NewNFTokenIDRange(6, 10)),
		newNFTokens(NFTokenID(11))}

	coinbaseInputAddresses  = []factom.Address{coinbase}
	coinbaseOutputAddresses = twoAddresses()

	coinbaseInputNFTokens  = []NFTokens{newNFTokens(NewNFTokenIDRange(0, 11))}
	coinbaseOutputNFTokens = []NFTokens{newNFTokens(NewNFTokenIDRange(0, 5)),
		newNFTokens(NewNFTokenIDRange(6, 11))}

	identityChainID = factom.NewBytes32(validIdentityChainID())
	tokenChainID    = fat0.ChainID("test", identityChainID)
)

func newNFTokens(ids ...NFTokensSetter) NFTokens {
	nfTkns, err := NewNFTokens(ids...)
	if err != nil {
		panic(err)
	}
	return nfTkns
}

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
		"inputs":        coinbaseInputs(),
		"outputs":       coinbaseOutputs(),
		"metadata":      []int{0},
		"tokenmetadata": tokenMetadata(),
	}
}

// inputs/outputs
func inputs() map[string]NFTokens {
	inputs := map[string]NFTokens{}
	for i := range inputAddresses {
		tkns := newNFTokens()
		tkns.Append(inputNFTokens[i])
		inputs[inputAddresses[i].String()] = tkns
	}
	return inputs
}
func outputs() map[string]NFTokens {
	outputs := map[string]NFTokens{}
	for i := range outputAddresses {
		tkns := newNFTokens()
		tkns.Append(outputNFTokens[i])
		outputs[outputAddresses[i].String()] = tkns
	}
	return outputs
}
func coinbaseInputs() map[string]NFTokens {
	inputs := map[string]NFTokens{}
	for i := range coinbaseInputAddresses {
		tkns := newNFTokens()
		tkns.Append(coinbaseInputNFTokens[i])
		inputs[coinbaseInputAddresses[i].String()] = tkns
	}
	return inputs
}
func coinbaseOutputs() map[string]NFTokens {
	outputs := map[string]NFTokens{}
	for i := range coinbaseOutputAddresses {
		tkns := newNFTokens()
		tkns.Append(coinbaseOutputNFTokens[i])
		outputs[coinbaseOutputAddresses[i].String()] = tkns
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
		t.Inputs[*coinbase.RCDHash()], _ = NewNFTokens()
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
	Error: "json: error calling MarshalJSON for type *fat1.Transaction: Inputs and Outputs mismatch: number of NFTokenIDs differ",
	Tx: func() Transaction {
		t := newTransaction()
		t.Inputs[*inputAddresses[0].RCDHash()].Set(NFTokenID(12345))
		return t
	}(),
}, {
	Name:  "invalid metadata JSON",
	Error: "json: error calling MarshalJSON for type *fat1.Transaction: json: error calling MarshalJSON for type json.RawMessage: invalid character 'a' looking for beginning of object key string",
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
				assert.NoError(err, string(test.Tx.Content))
			} else {
				assert.EqualError(err, test.Error, string(test.Tx.Content))
			}
		})
	}
}

func newTransaction() Transaction {
	return Transaction{
		Inputs:  inputAddressNFTokensMap(),
		Outputs: outputAddressNFTokensMap(),
	}
}
func inputAddressNFTokensMap() AddressNFTokensMap {
	return addressNFTokensMap(inputs())
}
func outputAddressNFTokensMap() AddressNFTokensMap {
	return addressNFTokensMap(outputs())
}
func addressNFTokensMap(aas map[string]NFTokens) AddressNFTokensMap {
	m := make(AddressNFTokensMap)
	for addressStr, amount := range aas {
		a := factom.Address{}
		if err := a.FromString(addressStr); err != nil {
			panic(err)
		}
		m[*a.RCDHash()] = amount
	}
	return m
}

func twoAddresses() []factom.Address {
	adrs := make([]factom.Address, 2)
	for i := range adrs {
		publicKey, privateKey, err := ed25519.GenerateKey(randSource)
		if err != nil {
			panic(err)
		}
		copy(adrs[i].PublicKey()[:], publicKey[:])
		copy(adrs[i].PrivateKey()[:], privateKey[:])

	}
	return adrs
}

func validIdentityChainID() factom.Bytes {
	return hexToBytes("88888807e4f3bbb9a2b229645ab6d2f184224190f83e78761674c2362aca4425")
}

func hexToBytes(hexStr string) factom.Bytes {
	raw, err := hex.DecodeString(hexStr)
	if err != nil {
		panic(err)
	}
	return factom.Bytes(raw)
}

func tokenMetadata() NFTokenIDMetadataMap {
	m := make(NFTokenIDMetadataMap, len(coinbaseInputNFTokens[0]))
	for i, tkns := range inputNFTokens {
		m.Set(NFTokenMetadata{Tokens: tkns,
			Metadata: json.RawMessage(fmt.Sprintf("%v", i))})
	}
	return m
}
