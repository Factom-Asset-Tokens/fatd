package fat1

import (
	"strconv"
)

const max = 1121

var (
	disjoint1000TknsJSON, disjoint1000Tkns = func() (string, NFTokens) {
		var json = make([]byte, max*7)
		i := copy(json, "[0")
		var tkns = make(NFTokens, max)
		NFTokenID(0).Set(tkns)
		for id := NFTokenID(2); len(tkns) < max; id += 2 {
			i += copy(json[i:], ","+strconv.FormatUint(uint64(id), 10))
			id.Set(tkns)
		}
		i += copy(json[i:], "]")
		return string(json[:i]), tkns
	}()
)
