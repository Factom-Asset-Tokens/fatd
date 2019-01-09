package fat1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sort"
)

// NFTokenIDRange represents a contiguous range of NFTokenIDs.
type NFTokenIDRange struct {
	Min NFTokenID `json:"min"`
	Max NFTokenID `json:"max"`
}

func NewNFTokenIDRange(minMax ...NFTokenID) NFTokenIDRange {
	if len(minMax) >= 2 {
		sort.Slice(minMax, func(i, j int) bool {
			return minMax[i] < minMax[j]
		})
		return NFTokenIDRange{Min: minMax[0], Max: minMax[1]}
	}
	if len(minMax) == 1 {
		return NFTokenIDRange{Min: minMax[0], Max: minMax[0]}
	}
	return NFTokenIDRange{}
}

func (idRange NFTokenIDRange) JSONLen() int {
	return len(`{"min":`) +
		idRange.Min.JSONLen() +
		len(`,"max":`) +
		idRange.Max.JSONLen() +
		len(`}`)
}

func (idRange NFTokenIDRange) IsEfficient() bool {
	var expandedLen int
	for id := idRange.Min; id <= idRange.Max; id++ {
		expandedLen += id.JSONLen() + len(`,`)
	}
	return idRange.JSONLen() <= expandedLen
}

func (idRange NFTokenIDRange) Slice() []NFTokenID {
	nfTknIDs := make([]NFTokenID, idRange.Len())
	for i, id := 0, idRange.Min; id <= idRange.Max; i, id = i+1, id+1 {
		nfTknIDs[i] = id
	}
	return nfTknIDs
}

func (idRange NFTokenIDRange) Len() int {
	return int(idRange.Max - idRange.Min + 1)
}

func (idRange NFTokenIDRange) Set(nfTkns NFTokens) error {
	for id := idRange.Min; id <= idRange.Max; id++ {
		if err := id.Set(nfTkns); err != nil {
			return err
		}
	}
	return nil
}

func (idRange NFTokenIDRange) IsValid() error {
	if idRange.Min > idRange.Max {
		return fmt.Errorf("Min is greater than Max")
	}
	return nil
}

type nfTokenIDRange NFTokenIDRange

func (idRange NFTokenIDRange) MarshalJSON() ([]byte, error) {
	if err := idRange.IsValid(); err != nil {
		return nil, fmt.Errorf("%T: %v", idRange, err)
	}
	return json.Marshal(nfTokenIDRange(idRange))
}

func (idRange *NFTokenIDRange) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, (*nfTokenIDRange)(idRange)); err != nil {
		return fmt.Errorf("%T: %v", idRange, err)
	}
	jsonLen := compactJSONLen(data)
	if jsonLen != idRange.JSONLen() {
		return fmt.Errorf("%T: unexpected JSON length", idRange)
	}
	if err := idRange.IsValid(); err != nil {
		return fmt.Errorf("%T: %v", idRange, err)
	}
	return nil
}

func compactJSONLen(data []byte) int {
	buf := bytes.NewBuffer(make([]byte, 0, len(data)))
	json.Compact(buf, data)
	cmp, _ := ioutil.ReadAll(buf)
	return len(cmp)
}
