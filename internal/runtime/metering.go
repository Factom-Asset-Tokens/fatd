// MIT License
//
// Copyright 2019 Canonical Ledgers, LLC
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
// IN THE SOFTWARE.

package runtime

import (
	"fmt"
)

var Called map[string]struct{}
var Cost = map[string]uint64{
	"get_height":     1,
	"get_precision":  1,
	"get_amount":     1,
	"get_timestamp":  1,
	"get_entry_hash": 1,
	"get_sender":     1,
	"get_address":    1,
	"get_coinbase":   1,
	"get_balance":    1,
	"get_balance_of": 1,
	"send":           1,
	"burn":           1,
	"revert":         0,
	"self_destruct":  0,
}

func (ctx *Context) Meter(fname string) error {
	cost, ok := Cost[fname]
	if !ok {
		ctx.ConsumeAllGas()
		ctx.Err = fmt.Errorf("missing cost for %q", fname)
		return ctx.Err
	}

	used := ctx.GetPointsUsed() + cost
	ctx.SetPointsUsed(used)

	limit := ctx.GetExecLimit()
	if used > limit {
		ctx.Err = ErrorExecLimitExceeded{}
		return ctx.Err
	}

	if Called != nil {
		Called[fname] = struct{}{}
	}

	return nil
}
