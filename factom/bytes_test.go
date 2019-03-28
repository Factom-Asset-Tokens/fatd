package factom

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	JSONBytesInvalidTypes     = []string{`{}`, `5.5`, `["hello"]`}
	JSONBytes32InvalidLengths = []string{
		`"00"`,
		`"000000000000000000000000000000000000000000000000000000000000000000"`}
	JSONBytesInvalidSymbol = `"x000000000000000000000000000000000000000000000000000000000000000"`
	JSONBytes32Valid       = `"da56930e8693fb7c0a13aac4d01cf26184d760f2fd92d2f0a62aa630b1a25fa7"`
)

type unmarshalJSONTest struct {
	Name string
	Data string
	Err  string
	Un   interface {
		json.Unmarshaler
		json.Marshaler
	}
	Exp interface {
		json.Unmarshaler
		json.Marshaler
	}
}

var unmarshalJSONtests = []unmarshalJSONTest{{
	Name: "Bytes32/valid",
	Data: `"DA56930e8693fb7c0a13aac4d01cf26184d760f2fd92d2f0a62aa630b1a25fa7"`,
	Un:   new(Bytes32),
	Exp: &Bytes32{0xDA, 0x56, 0x93, 0x0e, 0x86, 0x93, 0xfb, 0x7c, 0x0a, 0x13,
		0xaa, 0xc4, 0xd0, 0x1c, 0xf2, 0x61, 0x84, 0xd7, 0x60, 0xf2, 0xfd,
		0x92, 0xd2, 0xf0, 0xa6, 0x2a, 0xa6, 0x30, 0xb1, 0xa2, 0x5f, 0xa7},
}, {
	Name: "Bytes/valid",
	Data: `"DA56930e8693fb7c0a13aac4d01cf26184d760f2fd92d2f0a62aa630b1a25fa7"`,
	Un:   new(Bytes),
	Exp: &Bytes{0xDA, 0x56, 0x93, 0x0e, 0x86, 0x93, 0xfb, 0x7c, 0x0a, 0x13,
		0xaa, 0xc4, 0xd0, 0x1c, 0xf2, 0x61, 0x84, 0xd7, 0x60, 0xf2, 0xfd,
		0x92, 0xd2, 0xf0, 0xa6, 0x2a, 0xa6, 0x30, 0xb1, 0xa2, 0x5f, 0xa7},
}, {
	Name: "Bytes32/valid",
	Data: `"0000000000000000000000000000000000000000000000000000000000000000"`,
	Un:   new(Bytes32),
	Exp:  &Bytes32{},
}, {
	Name: "Bytes/valid",
	Data: `"0000000000000000000000000000000000000000000000000000000000000000"`,
	Un:   new(Bytes),
	Exp:  func() *Bytes { b := make(Bytes, 32); return &b }(),
}, {
	Name: "invalid symbol",
	Data: `"DA56930e8693fb7c0a13aac4d01cf26184d760f2fd92d2f0a62aa630b1zxcva7"`,
	Err:  "encoding/hex: invalid byte: U+007A 'z'",
}, {
	Name: "invalid type",
	Data: `{}`,
	Err:  "json: cannot unmarshal object into Go value of type string",
}, {
	Name: "invalid type",
	Data: `5.5`,
	Err:  "json: cannot unmarshal number into Go value of type string",
}, {
	Name: "invalid type",
	Data: `["asdf"]`,
	Err:  "json: cannot unmarshal array into Go value of type string",
}, {
	Name: "too long",
	Data: `"DA56930e8693fb7c0a13aac4d01cf26184d760f2fd92d2f0a62aa630b1a25fa71234"`,
	Err:  "invalid length",
	Un:   new(Bytes32),
}, {
	Name: "too short",
	Data: `"DA56930e8693fb7c0a13aac4d01cf26184d760f2fd92d2f0a62aa630b1a25fa71234"`,
	Err:  "invalid length",
	Un:   new(Bytes32),
}}

func testUnmarshalJSON(t *testing.T, test unmarshalJSONTest) {
	assert := assert.New(t)
	err := test.Un.UnmarshalJSON([]byte(test.Data))
	if len(test.Err) > 0 {
		assert.EqualError(err, test.Err)
		return
	}
	assert.NoError(err)
	assert.Equal(test.Exp, test.Un)
}

func TestBytes(t *testing.T) {
	for _, test := range unmarshalJSONtests {
		if test.Un != nil {
			t.Run("UnmarshalJSON/"+test.Name, func(t *testing.T) {
				testUnmarshalJSON(t, test)
			})
			if test.Exp != nil {
				t.Run("MarshalJSON/"+test.Name, func(t *testing.T) {
					data, err := test.Un.MarshalJSON()
					assert := assert.New(t)
					assert.NoError(err)
					assert.Equal(strings.ToLower(test.Data),
						string(data))
				})
			}
			continue
		}
		test.Un = new(Bytes32)
		t.Run("UnmarshalJSON/Bytes32/"+test.Name, func(t *testing.T) {
			testUnmarshalJSON(t, test)
		})
		test.Un = new(Bytes)
		t.Run("UnmarshalJSON/Bytes/"+test.Name, func(t *testing.T) {
			testUnmarshalJSON(t, test)
		})
	}

	t.Run("Scan", func(t *testing.T) {
		var b Bytes32
		err := b.Scan(5)
		assert := assert.New(t)
		assert.EqualError(err, "invalid type")

		in := make([]byte, 32)
		in[0] = 0xff
		err = b.Scan(in[:10])
		assert.EqualError(err, "invalid length")

		err = b.Scan(in)
		assert.NoError(err)
		assert.EqualValues(in, b[:])
	})

	t.Run("Value", func(t *testing.T) {
		var b Bytes32
		b[0] = 0xff
		val, err := b.Value()
		assert := assert.New(t)
		assert.NoError(err)
		assert.Equal(b[:], val)
	})

	t.Run("ZeroBytes32", func(t *testing.T) {
		assert.Equal(t, Bytes32{}, ZeroBytes32())
	})
}
