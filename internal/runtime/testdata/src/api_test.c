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

#define RUN(test) { int ret = (test); if (ret != SUCCESS) { return ret; } }

int test_get_height() {
        if (ext_get_height() != GET_HEIGHT_EXP) {
                return GET_HEIGHT_ERR;
        }
        return 0;
}

int test_get_precision() {
        if (ext_get_precision() != GET_PRECISION_EXP) {
                return GET_PRECISION_ERR;
        }
        return 0;
}

int test_get_timestamp() {
        if (ext_get_timestamp() != GET_TIMESTAMP_EXP) {
                return GET_TIMESTAMP_ERR;
        }
        return 0;
}

int test_get_amount() {
        if (ext_get_amount() != GET_AMOUNT_EXP) {
                return GET_AMOUNT_ERR;
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

const int SIZE = 32;

int test_get_sender() {
        char sender[SIZE];
        ext_get_sender(sender);
        if (0 != verifyBuf(sender, SIZE, GET_SENDER_ERR)) {
                return GET_SENDER_ERR;
        }
        return 0;
}

int test_get_address() {
        char address[SIZE];
        ext_get_address(address);
        if (0 != verifyBuf(address, SIZE, GET_ADDRESS_ERR)) {
                return GET_ADDRESS_ERR;
        }
        return 0;
}

int test_get_entry_hash() {
        char hash[SIZE];
        ext_get_entry_hash(hash);
        if (0 != verifyBuf(hash, SIZE, GET_ENTRY_HASH_ERR)) {
                return GET_ENTRY_HASH_ERR;
        }
        return 0;
}

EXPORT int run_all() {
        RUN(test_get_timestamp());
        RUN(test_get_height());
        RUN(test_get_precision());
        RUN(test_get_amount());

        RUN(test_get_sender());
        RUN(test_get_address());
        RUN(test_get_entry_hash());

        return SUCCESS;
}
