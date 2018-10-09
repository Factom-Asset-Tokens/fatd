package flag

import (
	"fmt"
	"strings"

	"github.com/AdamSLevy/factom"
)

type AddressList []string

func (a AddressList) String() string {
	return strings.Join(a, ", ")
}
func (aL *AddressList) Set(s string) error {
	list := strings.Fields(s)
	// If not able to split on space, attempt to split on comma.
	if len(list) == 1 {
		list = strings.Split(s, ",")
	}
	for _, a := range list {
		if !factom.IsValidAddress(a) {
			return fmt.Errorf("Invalid FCT address: %#v", a)
		}
	}
	*aL = append(*aL, list...)
	return nil
}
