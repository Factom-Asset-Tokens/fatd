package fat1

import (
	"encoding/json"
	"fmt"
)

type NFTokenIDMetadataMap map[NFTokenID]json.RawMessage

type NFTokenMetadata struct {
	Tokens   NFTokens        `json:"ids"`
	Metadata json.RawMessage `json:"metadata"`
}

func (m *NFTokenIDMetadataMap) UnmarshalJSON(data []byte) error {
	var tknMs []struct {
		Tokens   json.RawMessage `json:"ids"`
		Metadata json.RawMessage `json:"metadata"`
	}
	if err := json.Unmarshal(data, &tknMs); err != nil {
		return fmt.Errorf("%T: %v", m, err)
	}
	*m = make(NFTokenIDMetadataMap, len(tknMs))
	var expectedJSONLen int
	for _, tknM := range tknMs {
		if len(tknM.Tokens) == 0 {
			return fmt.Errorf(`%T: missing required field "ids"`, m)
		}
		if len(tknM.Metadata) == 0 {
			return fmt.Errorf(`%T: missing required field "metadata"`, m)
		}
		var tkns NFTokens
		if err := tkns.UnmarshalJSON(tknM.Tokens); err != nil {
			return fmt.Errorf("%T: %v", m, err)
		}
		metadata := compactJSON(tknM.Metadata)
		expectedJSONLen += len(metadata) + len(compactJSON(tknM.Tokens))
		for tknID := range tkns {
			if _, ok := (*m)[tknID]; ok {
				return fmt.Errorf("%T: Duplicate NFTokenID: %v", m, tknID)
			}
			(*m)[tknID] = metadata
		}
	}
	expectedJSONLen += len(`[]`) - len(`,`) +
		len(tknMs)*len(`{"ids":,"metadata":},`)
	if expectedJSONLen != len(compactJSON(data)) {
		return fmt.Errorf("%T: unexpected JSON length %v %v ", m, expectedJSONLen, len(compactJSON(data)))

	}
	return nil
}

func (m NFTokenIDMetadataMap) MarshalJSON() ([]byte, error) {
	metadataNFTokens := make(map[string]NFTokens, len(m))
	for tknID, metadata := range m {
		tkns := metadataNFTokens[string(metadata)]
		if tkns == nil {
			tkns = make(NFTokens)
			metadataNFTokens[string(metadata)] = tkns
		}
		if err := tknID.Set(tkns); err != nil {
			return nil, err
		}
	}

	var i int
	tknMs := make([]NFTokenMetadata, len(metadataNFTokens))
	for metadata, tkns := range metadataNFTokens {
		tknMs[i].Tokens = tkns
		tknMs[i].Metadata = json.RawMessage(metadata)
		i++
	}

	return json.Marshal(tknMs)
}

func (m NFTokenIDMetadataMap) IsSubsetOf(tkns NFTokens) error {
	if len(m) > len(tkns) {
		return fmt.Errorf("too many NFTokenIDs")
	}
	for tknID := range m {
		if _, ok := tkns[tknID]; ok {
			continue
		}
		return fmt.Errorf("NFTokenID(%v) is missing", tknID)
	}
	return nil
}

func (m NFTokenIDMetadataMap) Set(md NFTokenMetadata) {
	for tknID := range md.Tokens {
		m[tknID] = md.Metadata
	}
}
