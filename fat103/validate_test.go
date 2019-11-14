package fat103

import (
	"crypto/sha256"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/Factom-Asset-Tokens/factom"
	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	for _, test := range validateTests {
		test := test
		t.Run(test.Name, func(t *testing.T) { testValidate(t, test) })
	}
}

func testValidate(t *testing.T, test validateTest) {
	assert := assert.New(t)
	err := Validate(test.Entry, rcdHashes(test.Expected))
	if len(test.Error) == 0 {
		assert.NoError(err)
		return
	}
	assert.EqualError(err, test.Error)
}

type validateTest struct {
	Name string
	factom.Entry
	Expected []factom.RCDSigner
	Error    string
}

var validateTests = []validateTest{
	func() validateTest {
		e, adrs := validEntry(2)
		return validateTest{
			Name:     "valid",
			Entry:    e,
			Expected: adrs,
		}
	}(), func() validateTest {
		e, adrs := validEntry(100)
		return validateTest{
			Name:     "valid (large signing set)",
			Entry:    e,
			Expected: adrs,
		}
	}(), func() validateTest {
		e, adrs := validEntry(2)
		e.ExtIDs = nil
		return validateTest{
			Name:     "nil ExtIDs",
			Error:    "invalid number of ExtIDs",
			Entry:    e,
			Expected: adrs,
		}
	}(), func() validateTest {
		e, adrs := validEntry(2)
		e.ExtIDs = append(e.ExtIDs, factom.Bytes{})
		return validateTest{
			Name:     "extra ExtIDs",
			Error:    "invalid number of ExtIDs",
			Entry:    e,
			Expected: adrs,
		}
	}(), func() validateTest {
		e, adrs := validEntry(2)
		e.ExtIDs[0] = []byte("xxxx")
		return validateTest{
			Name:     "invalid timestamp (format)",
			Error:    "ExtIDs[0]: timestamp salt: strconv.ParseInt: parsing \"xxxx\": invalid syntax",
			Entry:    e,
			Expected: adrs,
		}
	}(), func() validateTest {
		e, adrs := validEntry(2)
		e.Timestamp = time.Now().Add(-48 * time.Hour)
		return validateTest{
			Name:     "invalid timestamp (expired)",
			Error:    "ExtIDs[0]: timestamp salt: expired",
			Entry:    e,
			Expected: adrs,
		}
	}(), func() validateTest {
		e, adrs := validEntry(2)
		e.Timestamp = time.Now().Add(48 * time.Hour)
		return validateTest{
			Name:     "invalid timestamp (expired)",
			Error:    "ExtIDs[0]: timestamp salt: expired",
			Entry:    e,
			Expected: adrs,
		}
	}(), func() validateTest {
		e, adrs := validEntry(2)
		e.ExtIDs[1] = append(e.ExtIDs[1], 0x00)
		return validateTest{
			Name:     "invalid RCD size",
			Error:    "ExtIDs[1]: invalid RCD size",
			Entry:    e,
			Expected: adrs,
		}
	}(), func() validateTest {
		e, adrs := validEntry(2)
		e.ExtIDs[1][0]++
		return validateTest{
			Name:     "invalid RCD type",
			Error:    "ExtIDs[1]: unsupported RCD",
			Entry:    e,
			Expected: adrs,
		}
	}(), func() validateTest {
		e, adrs := validEntry(2)
		e.ExtIDs[2] = append(e.ExtIDs[2], 0x00)
		return validateTest{
			Name:     "invalid signature size",
			Error:    "ExtIDs[1]: invalid signature size",
			Entry:    e,
			Expected: adrs,
		}
	}(), func() validateTest {
		e, adrs := validEntry(2)
		e.ExtIDs[2][0]++
		return validateTest{
			Name:     "invalid signatures",
			Error:    "ExtIDs[1]: invalid signature",
			Entry:    e,
			Expected: adrs,
		}
	}(), func() validateTest {
		e, adrs := validEntry(2)
		rcdSig := e.ExtIDs[1:3]
		e.ExtIDs[1] = e.ExtIDs[3]
		e.ExtIDs[2] = e.ExtIDs[4]
		e.ExtIDs[3] = rcdSig[0]
		e.ExtIDs[4] = rcdSig[1]
		return validateTest{
			Name:     "invalid signatures (transpose)",
			Error:    "ExtIDs[1]: invalid signature",
			Entry:    e,
			Expected: adrs,
		}
	}(), func() validateTest {
		e, adrs := validEntry(2)
		ts := time.Now().Add(time.Duration(
			-rand.Int63n(int64(12 * time.Hour))))
		timeSalt := []byte(strconv.FormatInt(ts.Unix(), 10))
		e.ExtIDs[0] = timeSalt
		return validateTest{
			Name:     "invalid signatures (timestamp)",
			Error:    "ExtIDs[1]: invalid signature",
			Entry:    e,
			Expected: adrs,
		}
	}(), func() validateTest {
		e, adrs := validEntry(2)
		e.ChainID = new(factom.Bytes32)
		e.ChainID[0] = 0x01
		e.ChainID[2] = 0x02
		return validateTest{
			Name:     "invalid signatures (chain ID)",
			Error:    "ExtIDs[1]: invalid signature",
			Entry:    e,
			Expected: adrs,
		}
	}(), func() validateTest {
		e, _ := validEntry(2)
		return validateTest{
			Name:     "unexpected RCD",
			Error:    "ExtIDs[1]: unexpected or duplicate RCD Hash",
			Entry:    e,
			Expected: genAddresses(2),
		}
	}(), func() validateTest {
		e, adrs := validEntry(3)
		e.ExtIDs = append(e.ExtIDs[:5], e.ExtIDs[1:3]...)
		return validateTest{
			Name:     "unexpected RCD (duplicate)",
			Error:    "ExtIDs[5]: invalid signature",
			Entry:    e,
			Expected: adrs,
		}
	}(),
}

func validEntry(n int) (factom.Entry, []factom.RCDSigner) {
	var e factom.Entry
	e.Content = factom.Bytes("some data to sign")
	e.ChainID = new(factom.Bytes32)
	*e.ChainID = factom.ComputeChainID([]factom.Bytes{factom.Bytes("test chain ID")})
	// Generate valid signatures with blank Addresses.
	adrs := genAddresses(n)
	e = Sign(e, adrs...)
	return e, adrs
}

func genAddresses(n int) []factom.RCDSigner {
	adrs := make([]factom.RCDSigner, n)
	for i := range adrs {
		adr, err := factom.GenerateFsAddress()
		if err != nil {
			panic(err)
		}
		adrs[i] = adr
	}
	return adrs
}

func rcdHashes(adrs []factom.RCDSigner) map[factom.Bytes32]struct{} {
	rcdHashes := make(map[factom.Bytes32]struct{}, len(adrs))
	for _, adr := range adrs {
		rcdHashes[sha256d(adr.RCD())] = struct{}{}
	}
	return rcdHashes
}

func sha256d(data []byte) [sha256.Size]byte {
	hash := sha256.Sum256(data)
	return sha256.Sum256(hash[:])
}
