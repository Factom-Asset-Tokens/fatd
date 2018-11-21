package fat0

import (
	"fmt"
	"strconv"
	"time"
)

// Time embeds time.Time and implements the json.Unmarshaler interface for
// correctly parsing the timestamp format used by fat0.
type Time struct {
	time.Time
}

// UnmarshalJSON unmarshals a string containing a Unix milliseconds since epoch
// timestamp.
func (t *Time) UnmarshalJSON(data []byte) error {
	ms, err := strconv.ParseUint(string(data), 10, 64)
	if err != nil {
		return fmt.Errorf("invalid timestamp")
	}
	t.Time = time.Unix(toSec(ms), toNanoSec(ms))
	return nil
}

func toSec(ms uint64) int64 {
	return int64(ms) / 1e3
}

func toNanoSec(ms uint64) int64 {
	return (int64(ms) % 1e3) * 1e6
}
