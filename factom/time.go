package factom

import (
	"fmt"
	"strconv"
	"time"
)

type Time struct {
	time.Time
}

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
