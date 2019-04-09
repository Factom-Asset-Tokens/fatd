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
	JSON      string
}{{
	Name: "valid",
	AdrNFTkns: AddressNFTokensMap{
		factom.FAAddress{0x00}: newNFTokens(NFTokenID(0), NFTokenID(1)),
	},
	JSON: `{"FA1y5ZGuHSLmf2TqNf6hVMkPiNGyQpQDTFJvDLRkKQaoPo4bmbgu":[0,1]}`,
}, {
	Name: "valid",
	AdrNFTkns: AddressNFTokensMap{
		factom.FAAddress{0x00}: newNFTokens(NewNFTokenIDRange(0, 1)),
		factom.FAAddress{0x01}: newNFTokens(NewNFTokenIDRange(2, 3)),
	},
	JSON: `{"FA1y5ZGuHSLmf2TqNf6hVMkPiNGyQpQDTFJvDLRkKQaoPo4bmbgu":[0,1],"FA1yX6omTQwz3WMuMgfTMexUP4Mks31VWAWAW8FMpPDsvhFY44yX":[2,3]}`,
}, {
	Name: "valid",
	AdrNFTkns: AddressNFTokensMap{
		factom.FAAddress{0x00}: newNFTokens(NewNFTokenIDRange(0, 1)),
		factom.FAAddress{0x01}: newNFTokens(NewNFTokenIDRange(2, 3)),
		factom.FAAddress{0x02}: newNFTokens(),
	},
	JSON: `{"FA1y5ZGuHSLmf2TqNf6hVMkPiNGyQpQDTFJvDLRkKQaoPo4bmbgu":[0,1],"FA1yX6omTQwz3WMuMgfTMexUP4Mks31VWAWAW8FMpPDsvhFY44yX":[2,3]}`,
}, {
	Name: "invalid, address with empty NFTokens",
	AdrNFTkns: AddressNFTokensMap{
		factom.FAAddress{0x00}: newNFTokens(),
	},
	Error: "json: error calling MarshalJSON for type fat1.AddressNFTokensMap: empty",
}, {
	Name:      "invalid, no addresses",
	AdrNFTkns: AddressNFTokensMap{},
	Error:     "json: error calling MarshalJSON for type fat1.AddressNFTokensMap: empty",
}, {
	Name: "invalid, has intersection",
	AdrNFTkns: AddressNFTokensMap{
		factom.FAAddress{0x00}: newNFTokens(NewNFTokenIDRange(0, 1)),
		factom.FAAddress{0x01}: newNFTokens(NewNFTokenIDRange(1, 3)),
	},
	Error: "json: error calling MarshalJSON for type fat1.AddressNFTokensMap: duplicate NFTokenID: ",
}}

func TestAddressNFTokensMapMarshal(t *testing.T) {
	for _, test := range AddressNFTokensMapMarshalTests {
		t.Run(test.Name, func(t *testing.T) {
			assert := assert.New(t)
			data, err := json.Marshal(test.AdrNFTkns)
			if len(test.Error) > 0 {
				assert.Contains(err.Error(), test.Error)
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
	JSON      string
}{{
	Name: "valid",
	AdrNFTkns: AddressNFTokensMap{
		factom.FAAddress{0x00}: newNFTokens(NFTokenID(0), NFTokenID(1)),
	},
	JSON: `{"FA1y5ZGuHSLmf2TqNf6hVMkPiNGyQpQDTFJvDLRkKQaoPo4bmbgu":[0,1]}`,
}, {
	Name: "valid",
	AdrNFTkns: AddressNFTokensMap{
		factom.FAAddress{0x00}: newNFTokens(NewNFTokenIDRange(0, 1)),
		factom.FAAddress{0x01}: newNFTokens(NewNFTokenIDRange(2, 3)),
	},
	JSON: `{"FA1y5ZGuHSLmf2TqNf6hVMkPiNGyQpQDTFJvDLRkKQaoPo4bmbgu":[0,1],"FA1yX6omTQwz3WMuMgfTMexUP4Mks31VWAWAW8FMpPDsvhFY44yX":[2,3]}`,
}, {
	Name:  "invalid, address with empty NFTokens",
	JSON:  `{"FA1y5ZGuHSLmf2TqNf6hVMkPiNGyQpQDTFJvDLRkKQaoPo4bmbgu":[],"FA1yX6omTQwz3WMuMgfTMexUP4Mks31VWAWAW8FMpPDsvhFY44yX":[2,3]}`,
	Error: "*fat1.AddressNFTokensMap: *fat1.NFTokens: empty: ",
}, {
	Name:  "invalid, no addresses",
	JSON:  `{}`,
	Error: "*fat1.AddressNFTokensMap: empty",
}, {
	Name:  "invalid, invalid NFTokens, duplicate NFTokenID",
	JSON:  `{"FA1y5ZGuHSLmf2TqNf6hVMkPiNGyQpQDTFJvDLRkKQaoPo4bmbgu":[0,0],"FA1yX6omTQwz3WMuMgfTMexUP4Mks31VWAWAW8FMpPDsvhFY44yX":[2,3]}`,
	Error: "*fat1.AddressNFTokensMap: *fat1.NFTokens: duplicate NFTokenID: 0: ",
}, {
	Name:  "invalid, has intersection",
	JSON:  `{"FA1y5ZGuHSLmf2TqNf6hVMkPiNGyQpQDTFJvDLRkKQaoPo4bmbgu":[0,1],"FA1yX6omTQwz3WMuMgfTMexUP4Mks31VWAWAW8FMpPDsvhFY44yX":[1,3]}`,
	Error: "*fat1.AddressNFTokensMap: duplicate NFTokenID: 1",
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
	Name:  "invalid, capacity exceeded",
	JSON:  `{"FA1y5ZGuHSLmf2TqNf6hVMkPiNGyQpQDTFJvDLRkKQaoPo4bmbgu":[{"min":1,"max":400000}],"FA1yX6omTQwz3WMuMgfTMexUP4Mks31VWAWAW8FMpPDsvhFY44yX":[2,3]}`,
	Error: `NFTokenID max capacity (400000) exceeded`,
}, {
	Name:  "invalid, capacity exceeded",
	JSON:  `{"FA1y5ZGuHSLmf2TqNf6hVMkPiNGyQpQDTFJvDLRkKQaoPo4bmbgu":[{"min":2,"max":400000}],"FA1yX6omTQwz3WMuMgfTMexUP4Mks31VWAWAW8FMpPDsvhFY44yX":[2,3]}`,
	Error: `NFTokenID max capacity (400000) exceeded`,
}, {
	Name:  "invalid, capacity exceeded",
	JSON:  `{"FA1y5ZGuHSLmf2TqNf6hVMkPiNGyQpQDTFJvDLRkKQaoPo4bmbgu":[{"min":0,"max":400000}],"FA1yX6omTQwz3WMuMgfTMexUP4Mks31VWAWAW8FMpPDsvhFY44yX":[2,3]}`,
	Error: `NFTokenID max capacity (400000) exceeded`,
}}

func TestAddressNFTokensMapUnmarshal(t *testing.T) {
	for _, test := range AddressNFTokensMapUnmarshalTests {
		t.Run(test.Name, func(t *testing.T) {
			assert := assert.New(t)
			adrNFTkns := AddressNFTokensMap{}
			err := adrNFTkns.UnmarshalJSON([]byte(test.JSON))
			if len(test.Error) > 0 {
				assert.Contains(err.Error(), test.Error)
			} else {
				assert.Equal(test.AdrNFTkns, adrNFTkns)
			}
		})
	}
}
