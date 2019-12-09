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

#include <runtime.h>
#include "./runtime_test.h"

int verifyBuf(char *buf, int size, char val);

EXPORT int run_all() {
        int32_t height = ext_get_height();
        if (height != GET_HEIGHT_EXP) {
                return GET_HEIGHT_ERR;
        }

        const int adrSize = 32;
        char adr[adrSize];
        ext_get_sender(adr);
        if (0 != verifyBuf(adr, adrSize, GET_SENDER_ERR)) {
                return GET_SENDER_ERR;
        }


        uint64_t amount = ext_get_amount();
        if (amount != GET_AMOUNT_EXP) {
                return GET_AMOUNT_ERR;
        }

        const int hashSize = 32;
        char hash[hashSize];
        ext_get_entry_hash(adr);
        if (0 != verifyBuf(adr, adrSize, GET_ENTRY_HASH_ERR)) {
                return GET_ENTRY_HASH_ERR;
        }

        return 0;
}

int verifyBuf(char *buf, int size, char val) {
        for (int i = 0; i < size; i++) {
                if (buf[i] != i+val) {
                        return 1;
                }
        }
        return 0;
}
