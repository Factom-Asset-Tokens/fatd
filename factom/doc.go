// MIT License
//
// Copyright 2018 Canonical Ledgers, LLC
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

// Package factom provides data types corresponding to some of the Factom
// blockchain's data structures, as well as methods on those types for querying
// the data from factomd and factom-walletd's APIs.
//
// All of the Factom data structure types in this package have the Get and
// IsPopulated methods.
//
// Methods that accept a *Client, like those that start with Get, may make
// calls to the factomd or factom-walletd API queries to populate the data in
// the variable on which it is called. The returned error can be checked to see
// if it is a jsonrpc2.Error type, indicating that the networking calls were
// successful, but that there is some error returned by the RPC method.
//
// IsPopulated methods return whether the data in the variable has been
// populated by a successful call to Get.
//
// The DBlock, EBlock and Entry types allow for exploring the Factom
// blockchain.
//
// The Bytes and Bytes32 types are used by other types when JSON marshaling and
// unmarshaling to and from hex encoded data is required. Bytes32 is used for
// Chain IDs and KeyMRs.
//
// The Address interfaces and types allow for working with the four Factom
// address types.
//
// The IDKey interfaces and types allow for working with the id/sk key pairs
// for server identities.
//
// Currently this package supports creating new chains and entries using both
// the factom-walletd "compose" methods, and by locally generating the commit
// and reveal data, if the private entry credit key is available locally. See
// Entry.Create and Entry.ComposeCreate.
//
// This package does not yet support Factoid transactions, nor does it support
// the binary data structures for DBlocks or EBlocks. Additionally, working
// with Identity Chains is not yet supported beyond querying the ID1Key.
package factom
