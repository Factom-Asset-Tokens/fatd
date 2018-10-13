package fat0

import (
	"encoding/json"
	"unicode/utf8"

	"bitbucket.org/canonical-ledgers/fatd/factom"
)

type Entry struct {
	*factom.Entry `json:"-"`
	Metadata      json.RawMessage `json:"metadata"`
}

func (e *Entry) Unmarshal(v interface{}) error {
	return json.Unmarshal(e.Content, v)
}

func (e *Entry) ValidExtID() bool {
	if e := e.ExtIDs; len(e) != 4 ||
		string(e[0]) != "token" || string(e[2]) != "issuer" ||
		len(e[3]) != 32 || !utf8.Valid(e[1]) {
		return false
	}
	return true
}

type Issuance struct {
	Entry
	Type   string `json:"type"`
	Name   string `json:"name"`
	Symbol string `json:"symbol"`
	Supply uint64 `json:"supply"`
}

func (i *Issuance) UnmarshalEntry() error {
	return i.Entry.Unmarshal(i)
}

type Transaction struct {
	*Entry
}

func (t *Transaction) UnmarshalEntry() error {
	return t.Entry.Unmarshal(t)
}
