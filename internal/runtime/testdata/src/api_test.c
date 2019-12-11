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

const int SIZE = 32;

int test_get_height() {
        if (ext_get_height() != GET_HEIGHT_EXP) {
                return GET_HEIGHT_ERR;
        }
        return SUCCESS;
}

int test_get_precision() {
        if (ext_get_precision() != GET_PRECISION_EXP) {
                return GET_PRECISION_ERR;
        }
        return SUCCESS;
}

int test_get_timestamp() {
        if (ext_get_timestamp() != GET_TIMESTAMP_EXP) {
                return GET_TIMESTAMP_ERR;
        }
        return SUCCESS;
}

int test_get_amount() {
        if (ext_get_amount() != GET_AMOUNT_EXP) {
                return GET_AMOUNT_ERR;
        }
        return SUCCESS;
}

int verifyBuf(char *buf, char val) {
        for (int i = 0; i < SIZE; i++) {
                if (buf[i] != i+val) {
                        return 1;
                }
        }
        return SUCCESS;
}

int test_get_sender() {
        char sender[SIZE];
        ext_get_sender(sender);
        if (SUCCESS != verifyBuf(sender, GET_SENDER_ERR)) {
                return GET_SENDER_ERR;
        }
        return SUCCESS;
}

int test_get_address() {
        char address[SIZE];
        ext_get_address(address);
        if (SUCCESS != verifyBuf(address, GET_ADDRESS_ERR)) {
                return GET_ADDRESS_ERR;
        }
        return SUCCESS;
}

int test_get_entry_hash() {
        char hash[SIZE];
        ext_get_entry_hash(hash);
        if (SUCCESS != verifyBuf(hash, GET_ENTRY_HASH_ERR)) {
                return GET_ENTRY_HASH_ERR;
        }
        return SUCCESS;
}

int test_get_balance() {
        if (ext_get_balance() != GET_BALANCE_EXP) {
                return GET_BALANCE_ERR;
        }
        return SUCCESS;
}

void populateBuf(char *buf, char val) {
        for (int i = 0; i < SIZE; i++) {
                buf[i] = i+val;
        }
}

int test_get_balance_of() {
        char adr[SIZE];
        populateBuf(adr, GET_BALANCE_OF_ERR);
        if (ext_get_balance_of(adr) != GET_BALANCE_OF_EXP) {
                return GET_BALANCE_OF_ERR;
        }
        return SUCCESS;
}

int test_send() {
        char adr[SIZE];
        populateBuf(adr, SEND_ERR_BALANCE);
        int bal = ext_get_balance();
        int bal_of = ext_get_balance_of(adr);
        ext_send(SEND_AMOUNT, adr);
        if (ext_get_balance() != (bal-SEND_AMOUNT)) {
                return SEND_ERR_BALANCE;
        }
        if (ext_get_balance_of(adr) != (bal_of+SEND_AMOUNT)) {
                return SEND_ERR_BALANCE_OF;
        }
        return SUCCESS;
}

int test_burn() {
        char adr[SIZE];
        ext_get_coinbase(adr);
        int bal = ext_get_balance();
        int burned = ext_get_balance_of(adr);
        ext_burn(BURN_AMOUNT);
        if (ext_get_balance() != (bal-BURN_AMOUNT)) {
                return BURN_ERR_BALANCE;
        }
        if (ext_get_balance_of(adr) != (burned+BURN_AMOUNT)) {
                return BURN_ERR_BALANCE_OF;
        }
        return SUCCESS;
}


EXPORT int run_all() {
        RUN(test_get_timestamp());
        RUN(test_get_height());
        RUN(test_get_precision());
        RUN(test_get_amount());

        RUN(test_get_sender());
        RUN(test_get_address());
        RUN(test_get_entry_hash());

        RUN(test_get_balance());
        RUN(test_get_balance_of());

        RUN(test_send());
        RUN(test_burn());

        return SUCCESS;
}
