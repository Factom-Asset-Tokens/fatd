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
	str := string(data)
	if str == "null" {
		return nil
	}
	sec, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		return fmt.Errorf("strconv.ParseUint(%#v, 10, 64): %v", str, err)
	}
	t.Time = time.Unix(int64(sec), 0)
	return nil
}
