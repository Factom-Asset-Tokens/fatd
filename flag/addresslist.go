package flag

import (
	"strings"

	. "github.com/Factom-Asset-Tokens/fatd/factom"
)

type FAAddressList []FAAddress

func (adrs FAAddressList) String() string {
	if len(adrs) == 0 {
		return ""
	}
	var s string
	for _, adr := range adrs {
		s += adr.String() + ","
	}
	return s[:len(s)-1]
}

// Set appends a comma seperated list of FAAddresses.
func (adrs *FAAddressList) Set(s string) error {
	adrStrs := strings.Split(s, ",")
	newAdrs := make(FAAddressList, len(adrStrs))
	for i, adrStr := range adrStrs {
		if err := newAdrs[i].Set(adrStr); err != nil {
			return err
		}
	}
	*adrs = append(*adrs, newAdrs...)
	return nil
}
