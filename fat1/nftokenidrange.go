package fat1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// NFTokenIDRange represents a contiguous range of NFTokenIDs.
type NFTokenIDRange struct {
	Min NFTokenID `json:"min"`
	Max NFTokenID `json:"max"`
}

func NewNFTokenIDRange(minMax ...NFTokenID) NFTokenIDRange {
	var min, max NFTokenID
	if len(minMax) >= 2 {
		min, max = minMax[0], minMax[1]
		if min > max {
			min, max = max, min
		}
	} else if len(minMax) == 1 {
		min, max = minMax[0], minMax[0]
	}
	return NFTokenIDRange{Min: min, Max: max}
}

func (idRange NFTokenIDRange) IsEfficient() bool {
	var expandedLen int
	for id := idRange.Min; id <= idRange.Max; id++ {
		expandedLen += id.jsonLen() + len(`,`)
	}
	return idRange.jsonLen() <= expandedLen
}

func (idRange NFTokenIDRange) Len() int {
	return int(idRange.Max - idRange.Min + 1)
}

func (idRange NFTokenIDRange) Set(tkns NFTokens) error {
	if len(tkns)+idRange.Len() > maxCapacity {
		return fmt.Errorf("%T(len:%v): %T(%v): %v",
			tkns, len(tkns), idRange, idRange, ErrorCapacity)
	}
	for id := idRange.Min; id <= idRange.Max; id++ {
		if err := id.Set(tkns); err != nil {
			return err
		}
	}
	return nil
}

func (idRange NFTokenIDRange) Valid() error {
	if idRange.Len() > maxCapacity {
		return ErrorCapacity
	}
	if idRange.Min > idRange.Max {
		return fmt.Errorf("Min is greater than Max")
	}
	return nil
}

type nfTokenIDRange NFTokenIDRange

func (idRange NFTokenIDRange) MarshalJSON() ([]byte, error) {
	if err := idRange.Valid(); err != nil {
		return nil, err
	}
	return json.Marshal(nfTokenIDRange(idRange))
}

func (idRange *NFTokenIDRange) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, (*nfTokenIDRange)(idRange)); err != nil {
		return fmt.Errorf("%T: %v", idRange, err)
	}
	if err := idRange.Valid(); err != nil {
		return fmt.Errorf("%T: %v", idRange, err)
	}
	if len(compactJSON(data)) != idRange.jsonLen() {
		return fmt.Errorf("%T: unexpected JSON length", idRange)
	}
	return nil
}
func (idRange NFTokenIDRange) jsonLen() int {
	return len(`{"min":`) +
		idRange.Min.jsonLen() +
		len(`,"max":`) +
		idRange.Max.jsonLen() +
		len(`}`)
}

func compactJSON(data []byte) []byte {
	buf := bytes.NewBuffer(make([]byte, 0, len(data)))
	json.Compact(buf, data)
	cmp, _ := ioutil.ReadAll(buf)
	return cmp
}
