package fat0

import (
	"bytes"
	"encoding/json"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/FactomProject/ed25519"
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

func (e *Entry) Sign(as ...factom.Address) {
	msg := append(e.ChainID[:], e.Content...)
	e.ExtIDs = nil
	for _, a := range as {
		e.ExtIDs = append(e.ExtIDs, a.RCD(), ed25519.Sign(a.PrivateKey, msg)[:])
	}
}
