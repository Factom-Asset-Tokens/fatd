package factom_test

import (
	"encoding/hex"
	"testing"

	. "github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var marshalBinaryTests = []struct {
	Name string
	Hash *Bytes32
	Entry
}{{
	Name: "valid",
	Entry: func() Entry {
		RpcConfig.FactomdServer = courtesyNode
		e := Entry{Hash: NewBytes32(hexToBytes(
			"935e8442a554383e50b02938420d16ef9fcc07d0a0ac03d191bd4275ddd98dee"))}
		if err := e.Get(); err != nil {
			panic(err)
		}
		if !e.IsPopulated() {
			panic("failed to populate")
		}
		return e
	}(),
}, {
	Name: "valid",
	Entry: Entry{
		Hash: NewBytes32(hexToBytes(
			"72177d733dcd0492066b79c5f3e417aef7f22909674f7dc351ca13b04742bb91")),
		ChainID: func() *Bytes32 { c := ChainID([]Bytes{Bytes("test")}); return &c }(),
		Content: hexToBytes("5061796c6f616448657265"),
	},
}}

func TestEntryMarshalBinary(t *testing.T) {
	for _, test := range marshalBinaryTests {
		t.Run(test.Name, func(t *testing.T) {
			e := test.Entry
			hash := e.ComputeHash()
			assert.Equal(t, *e.Hash, hash)
		})
	}
}

var unmarshalBinaryTests = []struct {
	Name  string
	Data  []byte
	Error string
	Hash  *Bytes32
}{{
	Name: "valid",
	Data: hexToBytes(
		"009005bb7dd69fb9910ee0b0db7b8a01198f03623eab6dadf1eba01f9dbc20757700530009436861696e54797065001253494e474c455f50524f4f465f434841494e000448617368002c4a74446f413157476a784f63584a67496365574e6336396a5551524867506835414e337848646b6a7158303d48796742426b32317a79384c576e5a56785a48526c38706b502f366e34377546317664324a4378654238593d"),
	Hash: NewBytes32(hexToBytes(
		"a5e49c1c14762f067b4132c5aa3abf03efdf2569de5d68a3f7cd539577f54942")),
}, {
	Name: "invalid (too short)",
	Data: hexToBytes(
		"009005bb7dd69fb9910ee0b0db7b8a01198f03623eab6dadf1eba01f9dbc207577"),
	Error: "insufficient length",
}, {
	Name: "invalid (version byte)",
	Data: hexToBytes(
		"019005bb7dd69fb9910ee0b0db7b8a01198f03623eab6dadf1eba01f9dbc20757700530009436861696e54797065001253494e474c455f50524f4f465f434841494e000448617368002c4a74446f413157476a784f63584a67496365574e6336396a5551524867506835414e337848646b6a7158303d48796742426b32317a79384c576e5a56785a48526c38706b502f366e34377546317664324a4378654238593d"),
	Error: "invalid version byte",
}, {
	Name: "invalid (ext ID Total Len)",
	Data: hexToBytes(
		"009005bb7dd69fb9910ee0b0db7b8a01198f03623eab6dadf1eba01f9dbc20757700010009436861696e54797065001253494e474c455f50524f4f465f434841494e000448617368002c4a74446f413157476a784f63584a67496365574e6336396a5551524867506835414e337848646b6a7158303d48796742426b32317a79384c576e5a56785a48526c38706b502f366e34377546317664324a4378654238593d"),
	Error: "invalid ExtIDs length",
}, {
	Name: "invalid (ext ID Total Len)",
	Data: hexToBytes(
		"009005bb7dd69fb9910ee0b0db7b8a01198f03623eab6dadf1eba01f9dbc207577ffff0009436861696e54797065001253494e474c455f50524f4f465f434841494e000448617368002c4a74446f413157476a784f63584a67496365574e6336396a5551524867506835414e337848646b6a7158303d48796742426b32317a79384c576e5a56785a48526c38706b502f366e34377546317664324a4378654238593d"),
	Error: "invalid ExtIDs length",
}, {
	Name: "invalid (ext ID len)",
	Data: hexToBytes(
		"009005bb7dd69fb9910ee0b0db7b8a01198f03623eab6dadf1eba01f9dbc20757700530008436861696e54797065001253494e474c455f50524f4f465f434841494e000448617368002c4a74446f413157476a784f63584a67496365574e6336396a5551524867506835414e337848646b6a7158303d48796742426b32317a79384c576e5a56785a48526c38706b502f366e34377546317664324a4378654238593d"),
	Error: "error parsing ExtIDs",
}, {
	Name: "invalid (ext ID len)",
	Data: hexToBytes(
		"009005bb7dd69fb9910ee0b0db7b8a01198f03623eab6dadf1eba01f9dbc2075770053000a436861696e54797065001253494e474c455f50524f4f465f434841494e000448617368002c4a74446f413157476a784f63584a67496365574e6336396a5551524867506835414e337848646b6a7158303d48796742426b32317a79384c576e5a56785a48526c38706b502f366e34377546317664324a4378654238593d"),
	Error: "error parsing ExtIDs",
}, {
	Name: "invalid (ext ID len)",
	Data: hexToBytes(
		"009005bb7dd69fb9910ee0b0db7b8a01198f03623eab6dadf1eba01f9dbc20757700530009436861696e54797065001253494e474c455f50524f4f465f434841494e000448617368002b4a74446f413157476a784f63584a67496365574e6336396a5551524867506835414e337848646b6a7158303d48796742426b32317a79384c576e5a56785a48526c38706b502f366e34377546317664324a4378654238593d"),
	Error: "error parsing ExtIDs",
}}

func TestEntryUnmarshalBinary(t *testing.T) {
	for _, test := range unmarshalBinaryTests {
		t.Run(test.Name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)
			e := Entry{}
			err := e.UnmarshalBinary(test.Data)
			if len(test.Error) == 0 {
				require.NoError(err)
				require.NotNil(e.ChainID)
				assert.Equal(*test.Hash, e.ComputeHash())
			} else {
				require.EqualError(err, test.Error)
				assert.Nil(e.ChainID)
				assert.Nil(e.Content)
				assert.Nil(e.ExtIDs)
			}
		})
	}
}

func hexToBytes(hexStr string) Bytes {
	raw, err := hex.DecodeString(hexStr)
	if err != nil {
		panic(err)
	}
	return Bytes(raw)
}
