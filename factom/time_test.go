package factom_test

import (
	"testing"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/stretchr/testify/assert"
)

var (
	JSONTimeInvalids = []struct {
		Data string
		Err  string
	}{
		{Data: `{}`,
			Err: "json: cannot unmarshal object into Go value of type uint64"},
		{Data: `"hello"`,
			Err: "json: cannot unmarshal string into Go value of type uint64"},
		{Data: `["hello"]`,
			Err: "json: cannot unmarshal array into Go value of type uint64"},
		{Data: `123445x`,
			Err: "invalid character 'x' after top-level value"}}
	JSONTimeValid = `1484858820`
)

func TestTimeUnmarshalJSON(t *testing.T) {
	for _, json := range JSONTimeInvalids {
		testTimeUnmarshalJSON(t, "Invalid", json.Data, json.Err)
	}
	json := JSONTimeValid
	t.Run("Valid", func(t *testing.T) {
		var time factom.Time
		assert.NoErrorf(t, time.UnmarshalJSON([]byte(json)), "json: %v", json)
	})
}

func testTimeUnmarshalJSON(t *testing.T, name string, json string, errStr string) {
	t.Run(name, func(t *testing.T) {
		var time factom.Time
		assert.EqualErrorf(t, time.UnmarshalJSON([]byte(json)),
			errStr, "json: %v", json)
	})
}
