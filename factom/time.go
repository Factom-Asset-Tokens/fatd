package factom

import (
	"fmt"
	"strconv"
	"time"
)

// Time embeds time.Time and implements the json.Unmarshaler interface for
// correctly parsing the timestamps returned by the factomd JSON RPC API.
type Time struct {
	time.Time
}

// UnmarshalJSON unmarshals a string containing a timestamp.
func (t *Time) UnmarshalJSON(data []byte) error {
	sec, err := strconv.ParseUint(string(data), 10, 64)
	if err != nil {
		return fmt.Errorf("invalid timestamp")
	}
	t.Time = time.Unix(int64(sec), 0)
	return nil
}
