// Package factom provides data types corresponding to some of the Factom
// blockchain's data structures, as well as methods on those types for querying
// the data from factomd's JSON RPC API. This is not intended to be a general
// purpose package for querying factomd's JSON RPC API. Only the data required
// for use in fatd is retained from calls to the factomd JSON RPC API.
//
// Each Factom data structure type has a method called Get, and a method called
// IsPopulated. Get makes the appropriate factomd JSON RPC API queries to
// populate the data in the variable on which it is called. IsPopulated
// indicates whether the data in the variable has been populated by a
// successful call to Get.
//
// Certain data already stored in the struct is used as the arguments for the
// factomd JSON RPC API queries. Which data needs to be populated prior to
// calling Get varies by the type on which it is being called and is documented
// with the respective functions.
//
// Get returns any networking or marshaling errors, but not JSON RPC errors,
// such as not being able to find the given Key MR. To check if the variable
// has been successfully populated, call IsPopulated().
//
// The provided Factom types are designed to reduce moving around large chucks
// of memory when passing them around by copy. All raw data is held in either
// slices or pointers to arrays in the case of chain IDs, key MRs, hashes, etc.
// Keep that in mind when using the Factom data structure types. You should
// normally pass the Factom data structure types around by copy, but keep in
// mind that the underlying data is shared.
//
// The Bytes and Bytes32 types are used when JSON marshaling and unmarshaling
// to and from hex encoded data is required.
package factom
