package fat0

import (
	"bytes"
	"encoding/json"

	"github.com/Factom-Asset-Tokens/fatd/factom"
)

type Entry struct {
	Metadata json.RawMessage `json:"metadata,omitempty"`

	*factom.Entry `json:"-"`
}

func (e *Entry) Unmarshal(v interface{}) error {
	d := json.NewDecoder(bytes.NewReader(e.Content))
	d.DisallowUnknownFields()
	return d.Decode(v)
}
