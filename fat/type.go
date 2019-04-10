package fat

import (
	"fmt"
	"strconv"
)

type Type uint64

const (
	TypeFAT0 Type = iota
	TypeFAT1
)

func (t *Type) Set(s string) error {
	format := s[0:len(`FAT-`)]
	if format != `FAT-` {
		return fmt.Errorf("%T: invalid format", t)
	}
	num := s[len(format):]
	var err error
	if *(*uint64)(t), err = strconv.ParseUint(num, 10, 64); err != nil {
		return fmt.Errorf("%T: %v", t, err)
	}
	return nil
}

func (t *Type) UnmarshalJSON(data []byte) error {
	if data[0] != '"' || data[len(data)-1] != '"' {
		return fmt.Errorf("%T: expected JSON string", t)
	}
	data = data[1 : len(data)-1]
	return t.Set(string(data))
}

func (t Type) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%#v", t.String())), nil
}

func (t Type) String() string {
	return fmt.Sprintf("FAT-%v", uint64(t))
}

func (t Type) IsValid() bool {
	switch t {
	case TypeFAT0:
		fallthrough
	case TypeFAT1:
		return true
	}
	return false
}
