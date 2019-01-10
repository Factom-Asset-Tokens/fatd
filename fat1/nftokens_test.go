package fat1

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newNFTokens(ids ...NFTokensSetter) NFTokens {
	nfTkns, err := NewNFTokens(ids...)
	if err != nil {
		panic(err)
	}
	return nfTkns
}

var NFTokensMarshalTests = []struct {
	Name   string
	NFTkns NFTokens
	Error  string
	JSON   string
}{{
	Name:   "valid, contiguous, expanded (0-7)",
	NFTkns: newNFTokens(NewNFTokenIDRange(0, 7)),
	JSON:   `[0,1,2,3,4,5,6,7]`,
}, {
	Name:   "valid, contiguous, condensed (0-12)",
	NFTkns: newNFTokens(NewNFTokenIDRange(0, 12)),
	JSON:   `[{"min":0,"max":12}]`,
}, {
	Name: "valid, disjoint (0-7, 9-20, 22, 100-10000, 10002)",
	NFTkns: newNFTokens(
		NewNFTokenIDRange(0, 7),
		NewNFTokenIDRange(9, 20),
		NFTokenID(22),
		NewNFTokenIDRange(1e2, 1e4),
		NFTokenID(1e4+2),
	),
	JSON: `[0,1,2,3,4,5,6,7,{"min":9,"max":20},22,{"min":100,"max":10000},10002]`,
}, {
	Name:   "invalid, empty",
	NFTkns: newNFTokens(),
	Error:  "fat1.NFTokens: empty",
}}

func TestNFTokensMarshal(t *testing.T) {
	for _, test := range NFTokensMarshalTests {
		t.Run(test.Name, func(t *testing.T) {
			assert := assert.New(t)
			data, err := test.NFTkns.MarshalJSON()
			if len(test.Error) > 0 {
				assert.EqualError(err, test.Error)
			} else {
				assert.Equal(test.JSON, string(data))
			}
		})
	}
}

var NFTokensUnmarshalTests = []struct {
	Name   string
	NFTkns NFTokens
	Error  string
	JSON   string
}{{
	Name:   "valid, contiguous, expanded (0-7)",
	JSON:   `[0,1,2,3,4,5,6,7]`,
	NFTkns: newNFTokens(NewNFTokenIDRange(0, 7)),
}, {
	Name:   "valid, contiguous, expanded, out of order (0-7)",
	JSON:   `[7,0,1,4,3,2,5,6]`,
	NFTkns: newNFTokens(NewNFTokenIDRange(0, 7)),
}, {
	Name:   "valid, contiguous, expanded, out of order ranges (0-7)",
	JSON:   `[7,0,1,4,3,2,5,6]`,
	NFTkns: newNFTokens(NewNFTokenIDRange(0, 7)),
}, {
	Name:   "valid, contiguous, condensed (0-12)",
	JSON:   `[{"min":0,"max":12}]`,
	NFTkns: newNFTokens(NewNFTokenIDRange(0, 12)),
}, {
	Name: "valid, disjoint (0-7, 9-20, 22, 100-10000, 10002)",
	JSON: `[0,1,2,3,4,5,6,7,{"min":9,"max":20},22,{"min":100,"max":10000},10002]`,
	NFTkns: newNFTokens(
		NewNFTokenIDRange(0, 7),
		NewNFTokenIDRange(9, 20),
		NFTokenID(22),
		NewNFTokenIDRange(1e2, 1e4),
		NFTokenID(1e4+2),
	),
}, {
	Name: "valid, disjoint, out of order (0-7, 9-20, 22, 100-10000, 10002)",
	JSON: `[0,{"min":9,"max":20},22,{"min":100,"max":10000},10002,1,2,3,4,5,6,7]`,
	NFTkns: newNFTokens(
		NewNFTokenIDRange(),
		NewNFTokenIDRange(1, 7),
		NewNFTokenIDRange(9, 20),
		NFTokenID(22),
		NewNFTokenIDRange(1e2, 1e4),
		NFTokenID(1e4+2),
	),
}, {
	Name: "valid, disjoint, out of order, inefficient (0-7, 9-20, 22, 100-10000, 10002)",
	JSON: `[0,{"min":9,"max":20},22,{"min":100,"max":10000},{"min":10002,"max":10002},1,2,3,4,5,6,7]`,
	NFTkns: newNFTokens(
		NewNFTokenIDRange(0, 7),
		NewNFTokenIDRange(9, 20),
		NFTokenID(22),
		NewNFTokenIDRange(1e2, 1e4),
		NFTokenID(1e4+2),
	),
}, {
	Name:  "invalid, empty",
	JSON:  `[]`,
	Error: "*fat1.NFTokens: empty",
}, {
	Name:  "invalid, duplicates",
	JSON:  `[0,0]`,
	Error: "*fat1.NFTokens: duplicate NFTokenID: 0",
}, {
	Name:  "invalid, duplicates, overlapping ranges",
	JSON:  `[{"min":0,"max":7},{"min":6,"max":10}]`,
	Error: "*fat1.NFTokens: duplicate NFTokenID: 6",
}, {
	Name:  "invalid, duplicates, overlapping ranges",
	JSON:  `[{"min":0,"max":7},{"min":6,"max":10}]`,
	Error: "*fat1.NFTokens: duplicate NFTokenID: 6",
}, {
	Name:  "invalid, invalid range",
	JSON:  `[{"min":5,"max":0},{"min":6,"max":10}]`,
	Error: "*fat1.NFTokenIDRange: Min is greater than Max",
}, {
	Name:  "invalid, malformed JSON",
	JSON:  `[{"min":5,"max":10},{"min":6,"max":10]]`,
	Error: "*fat1.NFTokens: invalid character ']' after object key:value pair",
}, {
	Name:  "invalid, NFTokenID JSON type",
	JSON:  `["hello",{"min":6,"max":10}]`,
	Error: "json: cannot unmarshal string into Go value of type fat1.NFTokenID",
}, {
	Name:  "invalid, NFTokenIDRange.Min JSON type",
	JSON:  `[{"min":{},"max":10}]`,
	Error: "*fat1.NFTokenIDRange: json: cannot unmarshal object into Go struct field nfTokenIDRange.min of type fat1.NFTokenID",
}, {
	Name:  "invalid, duplicate JSON keys",
	JSON:  `[{"min":6,"max":10,"max":11}]`,
	Error: "*fat1.NFTokenIDRange: unexpected JSON length",
}}

func TestNFTokensUnmarshal(t *testing.T) {
	for _, test := range NFTokensUnmarshalTests {
		t.Run(test.Name, func(t *testing.T) {
			assert := assert.New(t)
			nfTkns, _ := NewNFTokens()
			err := nfTkns.UnmarshalJSON([]byte(test.JSON))
			if len(test.Error) > 0 {
				assert.EqualError(err, test.Error)
			} else {
				assert.Equal(test.NFTkns, nfTkns)
			}
		})
	}
}

func TestNewNFTokens(t *testing.T) {
	var nfTknID NFTokenID
	_, err := NewNFTokens(nfTknID, nfTknID)
	assert.EqualError(t, err, "duplicate NFTokenID: 0")
}

func TestNFTokenIDRangeMarshal(t *testing.T) {
	idRange := NFTokenIDRange{Min: 5}
	_, err := json.Marshal(idRange)
	assert.EqualError(t, err, "json: error calling MarshalJSON for type fat1.NFTokenIDRange: Min is greater than Max")
}

func TestNFTokensIntersect(t *testing.T) {
	nfTkns1 := newNFTokens(NewNFTokenIDRange(0, 5))
	nfTkns2 := newNFTokens(NFTokenID(5))
	nfTkns3 := newNFTokens(NewNFTokenIDRange(6, 8))
	assert := assert.New(t)
	assert.EqualError(nfTkns1.NoIntersection(nfTkns2), "duplicate NFTokenID: 5")
	assert.EqualError(nfTkns2.NoIntersection(nfTkns1), "duplicate NFTokenID: 5")
	assert.NoError(nfTkns1.NoIntersection(nfTkns3))
}
