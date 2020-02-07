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


#ifndef RUNTIME_TEST_H
#define RUNTIME_TEST_H


#define INC(val) ((val) + 1)

#define SUCCESS 0

#define GET_HEIGHT_EXP 213518
#define GET_HEIGHT_ERR INC(SUCCESS)

#define GET_SENDER_ERR INC(GET_HEIGHT_ERR)

#define GET_AMOUNT_EXP 5001
#define GET_AMOUNT_ERR INC(GET_SENDER_ERR)

#define GET_ENTRY_HASH_ERR INC(GET_AMOUNT_ERR)

#define GET_TIMESTAMP_EXP 1575938086
#define GET_TIMESTAMP_ERR INC(GET_ENTRY_HASH_ERR)

#define GET_PRECISION_EXP 4
#define GET_PRECISION_ERR INC(GET_TIMESTAMP_ERR)

#define GET_ADDRESS_ERR INC(GET_PRECISION_ERR)

#define GET_BALANCE_EXP 987654321
#define GET_BALANCE_ERR INC(GET_ADDRESS_ERR)

#define GET_BALANCE_OF_EXP 123456789
#define GET_BALANCE_OF_ERR INC(GET_BALANCE_ERR)

#define SEND_AMOUNT INC(GET_BALANCE_OF_ERR)
#define SEND_ERR_BALANCE INC(GET_BALANCE_OF_ERR)
#define SEND_ERR_BALANCE_OF INC(SEND_ERR_BALANCE)

#define BURN_AMOUNT INC(GET_BALANCE_OF_ERR)
#define BURN_ERR_BALANCE INC(SEND_ERR_BALANCE_OF)
#define BURN_ERR_BALANCE_OF INC(BURN_ERR_BALANCE)

#endif // RUNTIME_TEST_H
