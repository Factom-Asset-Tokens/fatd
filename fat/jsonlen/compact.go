package jsonlen

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
)

func Compact(data []byte) []byte {
	buf := bytes.NewBuffer(make([]byte, 0, len(data)))
	json.Compact(buf, data)
	cmp, _ := ioutil.ReadAll(buf)
	return cmp
}
