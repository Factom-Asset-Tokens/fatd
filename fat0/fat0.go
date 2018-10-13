package fat0

import (
	. "bitbucket.org/canonical-ledgers/fatd/factom"
	"unicode/utf8"
)

func ValidExtID(b []Bytes) bool {
	if len(b) != 4 ||
		string(b[0]) != "token" || string(b[2]) != "issuer" ||
		len(b[3]) != 32 || !utf8.Valid(b[1]) {
		return false
	}
	return true
}

type Issuance struct {
	*Entry `json:"-"`
	Type   string `json:"type"`
	Name   string `json:"name"`
	Symbol string `json:"symbol"`
	Supply uint64 `json:"supply"`
}

type Transaction struct {
	*Entry
}
