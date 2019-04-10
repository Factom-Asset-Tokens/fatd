// Package factom provides data types corresponding to some of the Factom
// blockchain's data structures, as well as methods on those types for querying
// the data from factomd's JSON RPC API. This is not intended to be a general
// purpose package for querying factomd's JSON RPC API. Only the data required
// for use in fatd is queried and retained.
//
// All of the Factom data structure types in this package have the Get and
// IsPopulated methods.
//
// Get makes the appropriate factomd JSON RPC API queries to populate the data
// in the variable on which it is called.
//
// IsPopulated returns whether the data in the variable has been populated by a
// successful call to Get.
//
// Certain data already stored in the struct are used as the arguments for the
// factomd JSON RPC API queries. Which data needs to be populated prior to
// calling Get varies by the data type and is documented with the respective
// functions.
//
// Get returns any critical errors, like networking or marshaling errors, but
// not JSON RPC errors, such as factomd not being able to find the given KeyMR.
// To check if the variable has been successfully populated, call
// IsPopulated().
//
// The provided Factom types are designed to reduce moving around large chucks
// of memory when passing them around by copy. All raw data is held in either
// slices or pointers to arrays in the case of fixed length data that needs to
// be compared, like chain IDs, key MRs, hashes, etc. You should normally pass
// the Factom data structure types around by copy, but keep in mind that the
// underlying byte data is shared.
//
// The Bytes and Bytes32 types are used when JSON marshaling and unmarshaling
// to and from hex encoded data is required.
package factom
