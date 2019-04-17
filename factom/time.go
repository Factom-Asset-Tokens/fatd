package factom

import (
	"fmt"
	"strconv"
	"time"
)

// Time implements the json.Unmarshaler interface for correctly parsing the
// timestamps returned by the factomd JSON RPC API.
type Time time.Time

// UnmarshalJSON unmarshals a string containing a Unix seconds since epoch
// timestamp.
func (t *Time) UnmarshalJSON(data []byte) error {
	sec, err := strconv.ParseUint(string(data), 10, 64)
	if err != nil {
		return fmt.Errorf("invalid timestamp")
	}
	*t = Time(time.Unix(int64(sec), 0))
	return nil
}

func (t Time) Time() time.Time {
	return (time.Time)(t)
}
