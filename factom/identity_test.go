package factom

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var validIdentityChainIDStr = "88888807e4f3bbb9a2b229645ab6d2f184224190f83e78761674c2362aca4425"

func validIdentityChainID() Bytes {
	return hexToBytes(validIdentityChainIDStr)
}

func hexToBytes(hexStr string) Bytes {
	raw, err := hex.DecodeString(hexStr)
	if err != nil {
		panic(err)
	}
	return Bytes(raw)
}

func newID1Key(b []byte) *ID1Key {
	key := new(ID1Key)
	copy(key[:], b)
	return key
}

var validIdentityChainIDTests = []struct {
	Name    string
	Valid   bool
	ChainID Bytes
}{{
	Name:    "valid",
	ChainID: validIdentityChainID(),
	Valid:   true,
}, {
	Name:    "nil",
	ChainID: nil,
}, {
	Name:    "invalid length (short)",
	ChainID: validIdentityChainID()[0:15],
}, {
	Name:    "invalid length (long)",
	ChainID: append(validIdentityChainID(), 0x00),
}, {
	Name:    "invalid header",
	ChainID: func() Bytes { c := validIdentityChainID(); c[0]++; return c }(),
}, {
	Name:    "invalid header",
	ChainID: func() Bytes { c := validIdentityChainID(); c[1]++; return c }(),
}, {
	Name:    "invalid header",
	ChainID: func() Bytes { c := validIdentityChainID(); c[2]++; return c }(),
}}

func TestValidIdentityChainID(t *testing.T) {
	for _, test := range validIdentityChainIDTests {
		t.Run(test.Name, func(t *testing.T) {
			assert := assert.New(t)
			valid := ValidIdentityChainID(test.ChainID)
			if test.Valid {
				assert.True(valid)
			} else {
				assert.False(valid)
			}
		})
	}
}

func validIdentityNameIDs() []Bytes {
	return []Bytes{
		Bytes{0x00},
		Bytes("Identity Chain"),
		hexToBytes("f825c5629772afb5bce0464e5ea1af244be853a692d16360b8e03d6164b6adb5"),
		hexToBytes("28baa7d04e6c102991a184533b9f2443c9c314cc0327cc3a2f2adc0f3d7373a1"),
		hexToBytes("6095733cf6f5d0b5411d1eeb9f6699fad1ae27f9d4da64583bef97008d7bf0c9"),
		hexToBytes("966ebc2a0e3877ed846167e95ba3dde8561d90ee9eddd1bb74fbd6d1d25dba0f"),
		hexToBytes("33363533323533"),
	}
}

func invalidIdentityNameIDs(i int) []Bytes {
	n := validIdentityNameIDs()
	n[i] = Bytes{}
	return n
}

var validIdentityNameIDsTests = []struct {
	Name    string
	Valid   bool
	NameIDs []Bytes
}{{
	Name:    "valid",
	NameIDs: validIdentityNameIDs(),
	Valid:   true,
}, {
	Name:    "nil",
	NameIDs: nil,
}, {
	Name:    "invalid length (short)",
	NameIDs: validIdentityNameIDs()[0:6],
}, {
	Name:    "invalid length (long)",
	NameIDs: append(validIdentityNameIDs(), Bytes{}),
}, {
	Name:    "invalid length (long)",
	NameIDs: append(validIdentityNameIDs(), Bytes{}),
}, {
	Name:    "invalid ExtID",
	NameIDs: invalidIdentityNameIDs(0),
}, {
	Name:    "invalid ExtID",
	NameIDs: invalidIdentityNameIDs(1),
}, {
	Name:    "invalid ExtID",
	NameIDs: invalidIdentityNameIDs(2),
}, {
	Name:    "invalid ExtID",
	NameIDs: invalidIdentityNameIDs(3),
}, {
	Name:    "invalid ExtID",
	NameIDs: invalidIdentityNameIDs(4),
}, {
	Name:    "invalid ExtID",
	NameIDs: invalidIdentityNameIDs(5),
}}

func TestValidIdentityNameIDs(t *testing.T) {
	for _, test := range validIdentityNameIDsTests {
		t.Run(test.Name, func(t *testing.T) {
			assert := assert.New(t)
			valid := ValidIdentityNameIDs(test.NameIDs)
			if test.Valid {
				assert.True(valid)
			} else {
				assert.False(valid)
			}
		})
	}
}

func validIdentity() (i Identity) {
	i.ChainID = NewBytes32(validIdentityChainID())
	return
}

var identityTests = []struct {
	Name         string
	FactomServer string
	Valid        bool
	Error        string
	Height       uint64
	ID1Key       *ID1Key
	Identity
}{{
	Name:     "valid",
	Valid:    true,
	Identity: validIdentity(),
	Height:   140744,
	ID1Key: newID1Key(hexToBytes(
		"9656dbf91feb7d464971f31b28bfbf38ab201b8e33ec69ea4681e3bef779858e")),
}, {
	Name:     "nil chain ID",
	Error:    "ChainID is nil",
	Identity: Identity{},
}, {
	Name:         "bad factomd endpoint",
	FactomServer: "http://localhost:1000",
	Identity:     validIdentity(),
	Error:        "Post http://localhost:1000/v2: dial tcp [::1]:1000: connect: connection refused",
}, {
	Name: "malformed chain",
	Identity: NewIdentity(NewBytes32(hexToBytes(
		"8888885c2e0b523d9b8ab6d2975639e431eaba3fc9039ead32ce5065dcde86e4"))),
}, {
	Name: "invalid chain id",
	Identity: NewIdentity(NewBytes32(hexToBytes(
		"0088885c2e0b523d9b8ab6d2975639e431eaba3fc9039ead32ce5065dcde86e4"))),
}, {
	Name: "non-existent chain id",
	Identity: NewIdentity(NewBytes32(hexToBytes(
		"8888880000000000000000000000000000000000000000000000000000000000"))),
	Error: `jsonrpc2.Error{Code:-32009, Message:"Missing Chain Head"}`,
}}

var factomServer = "https://courtesy-node.factom.com"

var c = NewClient()

func TestIdentity(t *testing.T) {
	for _, test := range identityTests {
		t.Run(test.Name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)
			if len(test.FactomServer) == 0 {
				test.FactomServer = factomServer
			}
			c.FactomdServer = test.FactomServer
			i := test.Identity
			err := i.Get(c)
			populated := i.IsPopulated()
			if len(test.Error) > 0 {
				assert.EqualError(err, test.Error)
			} else {
				require.NoError(err)
			}
			if !test.Valid {
				assert.False(populated)
				return
			}
			assert.True(populated)
			assert.Equal(int(test.Height), int(i.Height))
			assert.Equal(*test.ID1Key, i.ID1)
			assert.NoError(i.Get(c))
		})
	}
}
