package fat1

import (
	"encoding/json"
	"testing"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/stretchr/testify/assert"
)

var AddressNFTokensMapMarshalTests = []struct {
	Name      string
	AdrNFTkns AddressNFTokensMap
	Error     string
	ErrorOr   string
	JSON      string
}{{
	Name: "valid",
	AdrNFTkns: AddressNFTokensMap{
		factom.RCDHash{0x00}: newNFTokens(NFTokenID(0), NFTokenID(1)),
	},
	JSON: `{"FA1y5ZGuHSLmf2TqNf6hVMkPiNGyQpQDTFJvDLRkKQaoPo4bmbgu":[0,1]}`,
}, {
	Name: "valid",
	AdrNFTkns: AddressNFTokensMap{
		factom.RCDHash{0x00}: newNFTokens(NewNFTokenIDRange(0, 1)),
		factom.RCDHash{0x01}: newNFTokens(NewNFTokenIDRange(2, 3)),
	},
	JSON: `{"FA1y5ZGuHSLmf2TqNf6hVMkPiNGyQpQDTFJvDLRkKQaoPo4bmbgu":[0,1],"FA1yX6omTQwz3WMuMgfTMexUP4Mks31VWAWAW8FMpPDsvhFY44yX":[2,3]}`,
}, {
	Name: "valid",
	AdrNFTkns: AddressNFTokensMap{
		factom.RCDHash{0x00}: newNFTokens(NewNFTokenIDRange(0, 1)),
		factom.RCDHash{0x01}: newNFTokens(NewNFTokenIDRange(2, 3)),
		factom.RCDHash{0x02}: newNFTokens(),
	},
	JSON: `{"FA1y5ZGuHSLmf2TqNf6hVMkPiNGyQpQDTFJvDLRkKQaoPo4bmbgu":[0,1],"FA1yX6omTQwz3WMuMgfTMexUP4Mks31VWAWAW8FMpPDsvhFY44yX":[2,3]}`,
}, {
	Name: "invalid, address with empty NFTokens",
	AdrNFTkns: AddressNFTokensMap{
		factom.RCDHash{0x00}: newNFTokens(),
	},
	Error: "json: error calling MarshalJSON for type fat1.AddressNFTokensMap: empty",
}, {
	Name:      "invalid, no addresses",
	AdrNFTkns: AddressNFTokensMap{},
	Error:     "json: error calling MarshalJSON for type fat1.AddressNFTokensMap: empty",
}, {
	Name: "invalid, has intersection",
	AdrNFTkns: AddressNFTokensMap{
		factom.RCDHash{0x00}: newNFTokens(NewNFTokenIDRange(0, 1)),
		factom.RCDHash{0x01}: newNFTokens(NewNFTokenIDRange(1, 3)),
	},
	Error:   "json: error calling MarshalJSON for type fat1.AddressNFTokensMap: FA1yX6omTQwz3WMuMgfTMexUP4Mks31VWAWAW8FMpPDsvhFY44yX and FA1y5ZGuHSLmf2TqNf6hVMkPiNGyQpQDTFJvDLRkKQaoPo4bmbgu: duplicate NFTokenID: 1",
	ErrorOr: "json: error calling MarshalJSON for type fat1.AddressNFTokensMap: FA1y5ZGuHSLmf2TqNf6hVMkPiNGyQpQDTFJvDLRkKQaoPo4bmbgu and FA1yX6omTQwz3WMuMgfTMexUP4Mks31VWAWAW8FMpPDsvhFY44yX: duplicate NFTokenID: 1",
}}

func TestAddressNFTokensMapMarshal(t *testing.T) {
	for _, test := range AddressNFTokensMapMarshalTests {
		t.Run(test.Name, func(t *testing.T) {
			assert := assert.New(t)
			data, err := json.Marshal(test.AdrNFTkns)
			if len(test.Error) > 0 {
				assert.True(err.Error() == test.Error ||
					err.Error() == test.ErrorOr, err.Error())
			} else {
				assert.Equal(test.JSON, string(data))
			}
		})
	}
}

var AddressNFTokensMapUnmarshalTests = []struct {
	Name      string
	AdrNFTkns AddressNFTokensMap
	Error     string
	ErrorOr   string
	JSON      string
}{{
	Name: "valid",
	AdrNFTkns: AddressNFTokensMap{
		factom.RCDHash{0x00}: newNFTokens(NFTokenID(0), NFTokenID(1)),
	},
	JSON: `{"FA1y5ZGuHSLmf2TqNf6hVMkPiNGyQpQDTFJvDLRkKQaoPo4bmbgu":[0,1]}`,
}, {
	Name: "valid",
	AdrNFTkns: AddressNFTokensMap{
		factom.RCDHash{0x00}: newNFTokens(NewNFTokenIDRange(0, 1)),
		factom.RCDHash{0x01}: newNFTokens(NewNFTokenIDRange(2, 3)),
	},
	JSON: `{"FA1y5ZGuHSLmf2TqNf6hVMkPiNGyQpQDTFJvDLRkKQaoPo4bmbgu":[0,1],"FA1yX6omTQwz3WMuMgfTMexUP4Mks31VWAWAW8FMpPDsvhFY44yX":[2,3]}`,
}, {
	Name:  "invalid, address with empty NFTokens",
	JSON:  `{"FA1y5ZGuHSLmf2TqNf6hVMkPiNGyQpQDTFJvDLRkKQaoPo4bmbgu":[],"FA1yX6omTQwz3WMuMgfTMexUP4Mks31VWAWAW8FMpPDsvhFY44yX":[2,3]}`,
	Error: "*fat1.AddressNFTokensMap: FA1y5ZGuHSLmf2TqNf6hVMkPiNGyQpQDTFJvDLRkKQaoPo4bmbgu: *fat1.NFTokens: empty",
}, {
	Name:  "invalid, no addresses",
	JSON:  `{}`,
	Error: "*fat1.AddressNFTokensMap: empty",
}, {
	Name:  "invalid, invalid NFTokens, duplicate NFTokenID",
	JSON:  `{"FA1y5ZGuHSLmf2TqNf6hVMkPiNGyQpQDTFJvDLRkKQaoPo4bmbgu":[0,0],"FA1yX6omTQwz3WMuMgfTMexUP4Mks31VWAWAW8FMpPDsvhFY44yX":[2,3]}`,
	Error: "*fat1.AddressNFTokensMap: FA1y5ZGuHSLmf2TqNf6hVMkPiNGyQpQDTFJvDLRkKQaoPo4bmbgu: *fat1.NFTokens: duplicate NFTokenID: 0",
}, {
	Name:    "invalid, has intersection",
	JSON:    `{"FA1y5ZGuHSLmf2TqNf6hVMkPiNGyQpQDTFJvDLRkKQaoPo4bmbgu":[0,1],"FA1yX6omTQwz3WMuMgfTMexUP4Mks31VWAWAW8FMpPDsvhFY44yX":[1,3]}`,
	Error:   "*fat1.AddressNFTokensMap: FA1y5ZGuHSLmf2TqNf6hVMkPiNGyQpQDTFJvDLRkKQaoPo4bmbgu and FA1yX6omTQwz3WMuMgfTMexUP4Mks31VWAWAW8FMpPDsvhFY44yX: duplicate NFTokenID: 1",
	ErrorOr: "*fat1.AddressNFTokensMap: FA1yX6omTQwz3WMuMgfTMexUP4Mks31VWAWAW8FMpPDsvhFY44yX and FA1y5ZGuHSLmf2TqNf6hVMkPiNGyQpQDTFJvDLRkKQaoPo4bmbgu: duplicate NFTokenID: 1",
}, {
	Name:  "invalid, invalid address",
	JSON:  `{"FA2y5ZGuHSLmf2TqNf6hVMkPiNGyQpQDTFJvDLRkKQaoPo4bmbgu":[0,1],"FA1yX6omTQwz3WMuMgfTMexUP4Mks31VWAWAW8FMpPDsvhFY44yX":[2,3]}`,
	Error: `*fat1.AddressNFTokensMap: "FA2y5ZGuHSLmf2TqNf6hVMkPiNGyQpQDTFJvDLRkKQaoPo4bmbgu": checksum error`,
}, {
	Name:  "invalid, duplicate address",
	JSON:  `{"FA1y5ZGuHSLmf2TqNf6hVMkPiNGyQpQDTFJvDLRkKQaoPo4bmbgu":[0,1],"FA1y5ZGuHSLmf2TqNf6hVMkPiNGyQpQDTFJvDLRkKQaoPo4bmbgu":[2,3]}`,
	Error: `*fat1.AddressNFTokensMap: unexpected JSON length`,
}, {
	Name:  "invalid, invalid JSON type",
	JSON:  `[0,1]`,
	Error: `*fat1.AddressNFTokensMap: json: cannot unmarshal array into Go value of type map[string]json.RawMessage`,
}, {
	Name:    "invalid, capacity exceeded",
	JSON:    `{"FA1y5ZGuHSLmf2TqNf6hVMkPiNGyQpQDTFJvDLRkKQaoPo4bmbgu":[{"min":1,"max":400000}],"FA1yX6omTQwz3WMuMgfTMexUP4Mks31VWAWAW8FMpPDsvhFY44yX":[2,3]}`,
	Error:   `*fat1.AddressNFTokensMap(len:400000): fat1.NFTokens(len:2): NFTokenID max capacity (400000) exceeded`,
	ErrorOr: `*fat1.AddressNFTokensMap(len:2): fat1.NFTokens(len:400000): NFTokenID max capacity (400000) exceeded`,
}, {
	Name:    "invalid, capacity exceeded",
	JSON:    `{"FA1y5ZGuHSLmf2TqNf6hVMkPiNGyQpQDTFJvDLRkKQaoPo4bmbgu":[{"min":2,"max":400000}],"FA1yX6omTQwz3WMuMgfTMexUP4Mks31VWAWAW8FMpPDsvhFY44yX":[2,3]}`,
	Error:   `*fat1.AddressNFTokensMap(len:399999): fat1.NFTokens(len:2): NFTokenID max capacity (400000) exceeded`,
	ErrorOr: `*fat1.AddressNFTokensMap(len:2): fat1.NFTokens(len:399999): NFTokenID max capacity (400000) exceeded`,
}, {
	Name:  "invalid, capacity exceeded",
	JSON:  `{"FA1y5ZGuHSLmf2TqNf6hVMkPiNGyQpQDTFJvDLRkKQaoPo4bmbgu":[{"min":0,"max":400000}],"FA1yX6omTQwz3WMuMgfTMexUP4Mks31VWAWAW8FMpPDsvhFY44yX":[2,3]}`,
	Error: `*fat1.AddressNFTokensMap: FA1y5ZGuHSLmf2TqNf6hVMkPiNGyQpQDTFJvDLRkKQaoPo4bmbgu: *fat1.NFTokens: *fat1.NFTokenIDRange: NFTokenID max capacity (400000) exceeded`,
}}

func TestAddressNFTokensMapUnmarshal(t *testing.T) {
	for _, test := range AddressNFTokensMapUnmarshalTests {
		t.Run(test.Name, func(t *testing.T) {
			assert := assert.New(t)
			adrNFTkns := AddressNFTokensMap{}
			err := adrNFTkns.UnmarshalJSON([]byte(test.JSON))
			if len(test.Error+test.ErrorOr) > 0 {
				if assert.Error(err) {
					assert.True(err.Error() == test.Error ||
						err.Error() == test.ErrorOr, err.Error())
				}
			} else {
				assert.Equal(test.AdrNFTkns, adrNFTkns)
			}
		})
	}
}
