// MIT License
//
// Copyright 2018 Canonical Ledgers, LLC
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
// IN THE SOFTWARE.

package fat_test

import (
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	. "github.com/Factom-Asset-Tokens/fatd/fat"
	"github.com/stretchr/testify/assert"
)

var validExtIDsTests = []struct {
	Name  string
	Error string
	Entry
}{{
	Name:  "valid",
	Entry: validEntry(),
}, {
	Name: "valid (large signing set)",
	Entry: func() Entry {
		e := validEntry()
		adrs := make([]factom.RCDPrivateKey, 100)
		for i := range adrs {
			adr, _ := factom.GenerateFsAddress()
			adrs[i] = adr
		}
		e.Sign(adrs...)
		return e
	}(),
}, {
	Name:  "nil ExtIDs",
	Error: "invalid number of ExtIDs",
	Entry: func() Entry {
		e := validEntry()
		e.ExtIDs = nil
		return e
	}(),
}, {
	Name:  "extra ExtIDs",
	Error: "invalid number of ExtIDs",
	Entry: func() Entry {
		e := validEntry()
		e.ExtIDs = append(e.ExtIDs, factom.Bytes{})
		return e
	}(),
}, {
	Name:  "invalid timestamp (format)",
	Error: "timestamp salt: strconv.ParseInt: parsing \"xxxx\": invalid syntax",
	Entry: func() Entry {
		e := validEntry()
		e.ExtIDs[0] = []byte("xxxx")
		return e
	}(),
}, {
	Name:  "invalid timestamp (expired)",
	Error: "timestamp salt expired",
	Entry: func() Entry {
		e := validEntry()
		e.Timestamp = time.Now().Add(-48 * time.Hour)
		return e
	}(),
}, {
	Name:  "invalid timestamp (expired)",
	Error: "timestamp salt expired",
	Entry: func() Entry {
		e := validEntry()
		e.Timestamp = time.Now().Add(48 * time.Hour)
		return e
	}(),
}, {
	Name:  "invalid RCD size",
	Error: "ExtIDs[1]: invalid RCD size",
	Entry: func() Entry {
		e := validEntry()
		e.ExtIDs[1] = append(e.ExtIDs[1], 0x00)
		return e
	}(),
}, {
	Name:  "invalid RCD type",
	Error: "ExtIDs[1]: invalid RCD type",
	Entry: func() Entry {
		e := validEntry()
		e.ExtIDs[1][0]++
		return e
	}(),
}, {
	Name: "invalid signature size",
	Entry: func() Entry {
		e := validEntry()
		e.ExtIDs[2] = append(e.ExtIDs[2], 0x00)
		return e
	}(),
	Error: "ExtIDs[1]: invalid signature size",
}, {
	Name: "invalid signatures",
	Entry: func() Entry {
		e := validEntry()
		e.ExtIDs[2][0]++
		return e
	}(),
	Error: "ExtIDs[2]: invalid signature",
}, {
	Name: "invalid signatures (transpose)",
	Entry: func() Entry {
		e := validEntry()
		rcdSig := e.ExtIDs[1:3]
		e.ExtIDs[1] = e.ExtIDs[3]
		e.ExtIDs[2] = e.ExtIDs[4]
		e.ExtIDs[3] = rcdSig[0]
		e.ExtIDs[4] = rcdSig[1]
		return e
	}(),
	Error: "ExtIDs[2]: invalid signature",
}, {
	Name: "invalid signatures (timestamp)",
	Entry: func() Entry {
		e := validEntry()
		ts := time.Now().Add(time.Duration(
			-rand.Int63n(int64(12 * time.Hour))))
		timeSalt := []byte(strconv.FormatInt(ts.Unix(), 10))
		e.ExtIDs[0] = timeSalt
		return e
	}(),
	Error: "ExtIDs[2]: invalid signature",
}, {
	Name: "invalid signatures (chain ID)",
	Entry: func() Entry {
		e := validEntry()
		e.ChainID = factom.NewBytes32(factom.Bytes{0x01, 0x02})
		return e
	}(),
	Error: "ExtIDs[2]: invalid signature",
},
}

func TestEntryValidExtIDs(t *testing.T) {
	for _, test := range validExtIDsTests {
		t.Run(test.Name, func(t *testing.T) {
			assert := assert.New(t)
			err := test.Entry.ValidExtIDs(len(test.Entry.ExtIDs) / 2)
			if len(test.Error) == 0 {
				assert.NoError(err)
			} else {
				assert.EqualError(err, test.Error)
			}
		})
	}
}

var randSource = rand.New(rand.NewSource(100))

func validEntry() Entry {
	var e Entry
	e.Content = factom.Bytes{0x00, 0x01, 0x02}
	e.ChainID = factom.NewBytes32(nil)
	// Generate valid signatures with blank Addresses.
	adrs := twoAddresses()
	e.Sign(adrs[0], adrs[1])
	return e
}

func twoAddresses() []factom.FsAddress {
	adrs := make([]factom.FsAddress, 2)
	for i := range adrs {
		adr, err := factom.GenerateFsAddress()
		if err != nil {
			panic(err)
		}
		adrs[i] = adr
	}
	return adrs
}
