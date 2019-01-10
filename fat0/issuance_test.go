package fat0_test

import (
	"encoding/json"
	"testing"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	. "github.com/Factom-Asset-Tokens/fatd/fat0"
	"github.com/FactomProject/ed25519"
	"github.com/stretchr/testify/assert"
)

var (
	identityChainID = factom.NewBytes32(validIdentityChainID())
)

func TestChainID(t *testing.T) {
	assert.Equal(t, "b54c4310530dc4dd361101644fa55cb10aec561e7874a7b786ea3b66f2c6fdfb",
		ChainID("test", identityChainID).String())
}

var validTokenNameIDsTests = []struct {
	Name    string
	NameIDs []factom.Bytes
	Valid   bool
}{{
	Name:    "valid",
	Valid:   true,
	NameIDs: validTokenNameIDs(),
}, {
	Name:    "invalid length (short)",
	NameIDs: validTokenNameIDs()[0:3],
}, {
	Name:    "invalid length (long)",
	NameIDs: append(validTokenNameIDs(), factom.Bytes{}),
}, {
	Name:    "invalid ExtID",
	NameIDs: invalidTokenNameIDs(0),
}, {
	Name:    "invalid ExtID",
	NameIDs: invalidTokenNameIDs(1),
}, {
	Name:    "invalid ExtID",
	NameIDs: invalidTokenNameIDs(2),
}, {
	Name:    "invalid ExtID",
	NameIDs: invalidTokenNameIDs(3),
}}

func TestValidTokenNameIDs(t *testing.T) {
	for _, test := range validTokenNameIDsTests {
		t.Run(test.Name, func(t *testing.T) {
			assert := assert.New(t)
			valid := ValidTokenNameIDs(test.NameIDs)
			if test.Valid {
				assert.True(valid)
			} else {
				assert.False(valid)
			}
		})
	}
}

func validTokenNameIDs() []factom.Bytes {
	return []factom.Bytes{
		factom.Bytes("token"),
		factom.Bytes("valid"),
		factom.Bytes("issuer"),
		identityChainID[:],
	}
}

func invalidTokenNameIDs(i int) []factom.Bytes {
	n := validTokenNameIDs()
	n[i] = factom.Bytes{}
	return n
}

var issuanceTests = []struct {
	Name      string
	Error     string
	IssuerKey factom.Address
	Issuance
}{{
	Name:      "valid",
	IssuerKey: issuerKey,
	Issuance:  validIssuance(),
}, {
	Name:      "valid (omit symbol)",
	IssuerKey: issuerKey,
	Issuance:  omitFieldIssuance("symbol"),
}, {
	Name:      "valid (omit name)",
	IssuerKey: issuerKey,
	Issuance:  omitFieldIssuance("name"),
}, {
	Name:      "valid (omit metadata)",
	IssuerKey: issuerKey,
	Issuance:  omitFieldIssuance("metadata"),
}, {
	Name:      "invalid JSON (unknown field)",
	Error:     `*fat0.Issuance: unexpected JSON length`,
	IssuerKey: issuerKey,
	Issuance:  setFieldIssuance("invalid", 5),
}, {
	Name:      "invalid JSON (invalid type)",
	Error:     `*fat0.Issuance: json: cannot unmarshal array into Go struct field issuance.type of type string`,
	IssuerKey: issuerKey,
	Issuance:  invalidIssuance("type"),
}, {
	Name:      "invalid JSON (invalid supply)",
	Error:     `*fat0.Issuance: json: cannot unmarshal array into Go struct field issuance.supply of type int64`,
	IssuerKey: issuerKey,
	Issuance:  invalidIssuance("supply"),
}, {
	Name:      "invalid JSON (invalid symbol)",
	Error:     `*fat0.Issuance: json: cannot unmarshal array into Go struct field issuance.symbol of type string`,
	IssuerKey: issuerKey,
	Issuance:  invalidIssuance("symbol"),
}, {
	Name:      "invalid JSON (invalid name)",
	Error:     `*fat0.Issuance: json: cannot unmarshal array into Go struct field issuance.name of type string`,
	IssuerKey: issuerKey,
	Issuance:  invalidIssuance("name"),
}, {
	Name:      "invalid JSON (nil)",
	Error:     `unexpected end of JSON input`,
	IssuerKey: issuerKey,
	Issuance:  issuance(nil),
}, {
	Name:      "invalid data (type)",
	Error:     `*fat0.Issuance: invalid "type": "invalid"`,
	IssuerKey: issuerKey,
	Issuance:  setFieldIssuance("type", "invalid"),
}, {
	Name:      "invalid data (type omitted)",
	Error:     `*fat0.Issuance: invalid "type": ""`,
	IssuerKey: issuerKey,
	Issuance:  omitFieldIssuance("type"),
}, {
	Name:      "invalid data (supply: 0)",
	Error:     `*fat0.Issuance: invalid "supply": must be positive or -1`,
	IssuerKey: issuerKey,
	Issuance:  setFieldIssuance("supply", 0),
}, {
	Name:      "invalid data (supply: -5)",
	Error:     `*fat0.Issuance: invalid "supply": must be positive or -1`,
	IssuerKey: issuerKey,
	Issuance:  setFieldIssuance("supply", -5),
}, {
	Name:      "invalid data (supply: omitted)",
	Error:     `*fat0.Issuance: invalid "supply": must be positive or -1`,
	IssuerKey: issuerKey,
	Issuance:  omitFieldIssuance("supply"),
}, {
	Name:      "invalid ExtIDs (timestamp)",
	Error:     `timestamp salt expired`,
	IssuerKey: issuerKey,
	Issuance: func() Issuance {
		i := validIssuance()
		i.ExtIDs[0] = factom.Bytes("10")
		return i
	}(),
}, {
	Name:      "invalid ExtIDs (length)",
	Error:     `incorrect number of ExtIDs`,
	IssuerKey: issuerKey,
	Issuance: func() Issuance {
		i := validIssuance()
		i.ExtIDs = append(i.ExtIDs, factom.Bytes{})
		return i
	}(),
}, {
	Name:     "invalid RCD hash",
	Error:    `invalid RCD`,
	Issuance: validIssuance(),
}}

func TestIssuance(t *testing.T) {
	for _, test := range issuanceTests {
		t.Run(test.Name, func(t *testing.T) {
			assert := assert.New(t)
			i := test.Issuance
			key := test.IssuerKey
			err := i.Valid(key.RCDHash())
			if len(test.Error) == 0 {
				assert.NoError(err)
			} else {
				assert.EqualError(err, test.Error)
			}
		})
	}
}

func validIssuanceEntryContentMap() map[string]interface{} {
	return map[string]interface{}{
		"type":     "FAT-0",
		"supply":   int64(100000),
		"symbol":   "TEST",
		"name":     "Test Token",
		"metadata": []int{0},
	}
}

func validIssuance() Issuance {
	return issuance(marshal(validIssuanceEntryContentMap()))
}

var issuerKey = func() factom.Address {
	a := factom.Address{}
	publicKey, privateKey, err := ed25519.GenerateKey(randSource)
	if err != nil {
		panic(err)
	}
	copy(a.PublicKey()[:], publicKey[:])
	copy(a.PrivateKey()[:], privateKey[:])
	return a
}()

func issuance(content factom.Bytes) Issuance {
	e := factom.Entry{
		ChainID: factom.NewBytes32(nil),
		Content: content,
	}
	i := NewIssuance(e)
	i.Sign(issuerKey)
	return i
}

func invalidIssuance(field string) Issuance {
	return setFieldIssuance(field, []int{0})
}

func omitFieldIssuance(field string) Issuance {
	m := validIssuanceEntryContentMap()
	delete(m, field)
	return issuance(marshal(m))
}

func setFieldIssuance(field string, value interface{}) Issuance {
	m := validIssuanceEntryContentMap()
	m[field] = value
	return issuance(marshal(m))
}

func marshal(v map[string]interface{}) []byte {
	data, _ := json.Marshal(v)
	return data
}

var issuanceMarshalEntryTests = []struct {
	Name  string
	Error string
	Issuance
}{{
	Name:     "valid",
	Issuance: newIssuance(),
}, {
	Name: "valid (metadata)",
	Issuance: func() Issuance {
		i := newIssuance()
		i.Metadata = json.RawMessage(`{"memo":"new token"}`)
		return i
	}(),
}, {
	Name:  "invalid data",
	Error: `json: error calling MarshalJSON for type *fat0.Issuance: invalid "type": "invalid"`,
	Issuance: func() Issuance {
		i := newIssuance()
		i.Type = "invalid"
		return i
	}(),
}, {
	Name:  "invalid metadata JSON",
	Error: `json: error calling MarshalJSON for type *fat0.Issuance: json: error calling MarshalJSON for type json.RawMessage: invalid character 'a' looking for beginning of object key string`,
	Issuance: func() Issuance {
		i := newIssuance()
		i.Metadata = json.RawMessage("{asdf")
		return i
	}(),
}}

func TestIssuanceMarshalEntry(t *testing.T) {
	for _, test := range issuanceMarshalEntryTests {
		t.Run(test.Name, func(t *testing.T) {
			assert := assert.New(t)
			i := test.Issuance
			err := i.MarshalEntry()
			if len(test.Error) == 0 {
				assert.NoError(err)
			} else {
				assert.EqualError(err, test.Error)
			}
		})
	}
}

func newIssuance() Issuance {
	return Issuance{
		Type:   "FAT-0",
		Supply: 1000000,
		Name:   "test coin",
		Symbol: "TEST",
	}
}
