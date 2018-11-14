package factom_test

import (
	"testing"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/stretchr/testify/assert"
)

var (
	JSONTimeInvalids = []string{`{}`, `"hello"`, `["hello"]`, `123445x`}
	JSONTimeValid    = `1484858820`
)

func TestTimeUnmarshalJSON(t *testing.T) {
	for _, json := range JSONTimeInvalids {
		testTimeUnmarshalJSON(t, "Invalid", json, "invalid timestamp")
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
