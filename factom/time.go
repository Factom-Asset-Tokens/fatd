package factom

import (
	"encoding/json"
	"time"
)

// Time implements the json.Unmarshaler and json.Marshaler interface for
// correctly parsing the timestamps returned by the factomd JSON RPC API.
type Time time.Time

// UnmarshalJSON unmarshals a JSON Number containing the Unix seconds since
// epoch timestamp.
func (t *Time) UnmarshalJSON(data []byte) error {
	var sec uint64
	if err := json.Unmarshal(data, &sec); err != nil {
		return err
	}
	*t = Time(time.Unix(int64(sec), 0))
	return nil
}

// MarshalJSON marshals a JSON Number containing a Unix seconds since epoch
// timestamp.
func (t Time) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Time().Unix())
}

func (t Time) Time() time.Time {
	return (time.Time)(t)
}
