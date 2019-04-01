package jsonlen

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var int64Tests = []struct {
	Name string
	D    int64
	Len  int
}{{
	Name: "zero",
	Len:  1,
}, {
	Name: "two digits",
	D:    10,
	Len:  2,
}, {
	Name: "two digits",
	D:    99,
	Len:  2,
}, {
	Name: "three digits",
	D:    100,
	Len:  3,
}, {
	Name: "three digits",
	D:    999,
	Len:  3,
}, {
	Name: "four digits",
	D:    1000,
	Len:  4,
}, {
	Name: "four digits",
	D:    9999,
	Len:  4,
}, {
	Name: "ten digits",
	D:    1000000000,
	Len:  10,
}, {
	Name: "ten digits",
	D:    9999999999,
	Len:  10,
}}

func TestInt64(t *testing.T) {
	for _, test := range int64Tests {
		t.Run(test.Name, func(t *testing.T) {
			assert := assert.New(t)
			assert.Equal(test.Len, Int64(test.D))
			if test.D != 0 {
				assert.Equal(test.Len+1, Int64(-test.D), "negative")
			}
		})
	}

}
