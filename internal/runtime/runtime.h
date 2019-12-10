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

#ifndef RUNTIME_H
#define RUNTIME_H

#define EXPORT __attribute__((used))
//#define IMPORT __attribute__((weak))

#include <stdint.h>

extern void ext_get_sender(char *);
extern void ext_get_entry_hash(char *);
extern void ext_get_address(char *);

extern uint64_t ext_get_balance(char *);

extern uint32_t ext_get_height();
extern uint32_t ext_get_precision();
extern uint64_t ext_get_amount();
extern uint64_t ext_get_timestamp();

#endif // RUNTIME_H
