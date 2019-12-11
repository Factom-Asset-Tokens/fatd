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

// ext_get_sender loads the 32 byte address of the sender into adr.
extern void ext_get_sender(char *adr);

// ext_get_entry_hash loads the 32 byte hash of the current transaction entry
// into hash.
extern void ext_get_entry_hash(char *hash);

// ext_get_address loads the 32 byte address of the contract into adr.
extern void ext_get_address(char *adr);

// ext_get_coinbase loads the 32 byte coinbase address into adr.
extern void ext_get_coinbase(char *adr);

// ext_get_balance returns the current balance of the contract's address.
extern uint64_t ext_get_balance();

// ext_get_balance_of returns the current balance of the 32 byte address at
// adr.
extern uint64_t ext_get_balance_of(char *adr);

// ext_get_height returns the current block height.
extern uint32_t ext_get_height(void);

// ext_get_precision returns the precision of the FAT-0 token chain. The
// precision is a value between 0 and 18 denoting the power of 10 that the
// display unit represents.
//
// Note that all amounts and balances are always returned in the base unit,
// which does not at all consider precision. You probably don't need this
// function in your contract.
extern uint32_t ext_get_precision(void);

// ext_get_amount returns the amount of FAT-0 sent in the current contract
// call.
extern uint64_t ext_get_amount(void);

// ext_get_timestamp returns the Unix timestamp in seconds of the current
// block.
extern uint64_t ext_get_timestamp(void);

// ext_send sends amount tokens to the 32 byte address at adr.
//
// If amount exceeds the contract address's balance, this will panic and
// revert/invalidate the transaction.
extern void ext_send(uint64_t amount, char *adr);

// ext_burn burns amount tokens.
//
// If amount exceeds the contract address's balance, this will panic and
// revert/invalidate the transaction.
extern void ext_burn(uint64_t amount);

// ext_revert stops all further execution of this transaction and reverts all
// state changes from this transaction, including any balance transfer, and
// marks the transaction as invalid.
extern void ext_revert(void);

// ext_self_destruct stops all further execution and destroys the contract
// returning control of the address back to its private key.
//
// After this is called, funds held by the contract may be withdrawn with a
// normal transaction using the private key of the address.
extern void ext_self_destruct(void);

#endif // RUNTIME_H
