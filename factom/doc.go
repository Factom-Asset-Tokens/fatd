// Package factom provides data types corresponding to some of the Factom
// blockchain's data structures, as well as methods on those types for querying
// the data from factomd and factom-walletd's APIs.
//
// All of the Factom data structure types in this package have the Get and
// IsPopulated methods.
//
// Methods that accept a *Client, like those that start with Get, make calls to
// the factomd or factom-walletd API queries to populate the data in the
// variable on which it is called. The returned error can be checked to see if
// it is a jsonrpc2.Error type, indicating that the networking calls were
// successful, but that there is some error returned by the RPC method.
//
// IsPopulated methods return whether the data in the variable has been
// populated by a successful call to Get.
//
// The Bytes and Bytes32 types are used by other types when JSON marshaling and
// unmarshaling to and from hex encoded data is required. Bytes32 is used for
// Chain IDs and KeyMRs.
//
// The DBlock, EBlock and Entry types allow for exploring the Factom
// blockchain.
//
// The Address interfaces and types allow for working with the four Factom
// address types.
package factom
