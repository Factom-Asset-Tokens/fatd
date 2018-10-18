package fat0

import (
	"fmt"
	"strconv"
	"time"
)

type Time struct {
	time.Time
}

func (t *Time) UnmarshalJSON(data []byte) error {
	ms, err := strconv.ParseUint(string(data), 10, 64)
	if err != nil {
		return fmt.Errorf("strconv.ParseUint(%#v, 10, 64): %v", string(data), err)
	}
	t.Time = time.Unix(int64(ms)/1e3, (int64(ms)%1e3)*1e6)
	return nil
}
