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

package factom

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"database/sql/driver"
	"fmt"

	"golang.org/x/crypto/ed25519"
)

// Notes: This file contains all types, interfaces, and methods related to
// Factom Addresses as specified by
// https://github.com/FactomProject/FactomDocs/blob/master/factomDataStructureDetails.md
//
// There are four Factom address types, forming two pairs: public and private
// Factoid addresses, and public and private Entry Credit addresses. All
// addresses are a 32 byte payload encoded using base58check with various
// prefixes.

// Address is the interface implemented by the four address types: FAAddress,
// FsAddress, ECAddress, and EsAddress.
type Address interface {
	// PrefixBytes returns the prefix bytes for the Address.
	PrefixBytes() []byte
	// PrefixString returns the encoded prefix string for the Address.
	PrefixString() string

	// String encodes the address to a base58check string with the
	// appropriate prefix.
	String() string
	// Payload returns the address as a byte array.
	Payload() [sha256.Size]byte

	// PublicAddress returns the corresponding public address in an Address
	// interface. Public addresses return themselves. Private addresses
	// compute the public address.
	PublicAddress() Address
	// GetPrivateAddress returns the corresponding private address in a
	// PrivateAddress interface. Public addresses query factom-walletd for
	// the private address. Private addresses return themselves.
	GetPrivateAddress(*Client) (PrivateAddress, error)

	// GetBalance returns the current balance for the address.
	GetBalance(*Client) (uint64, error)

	// Remove queries factom-walletd to remove the public and private
	// addresses from its database.
	// WARNING: DESTRUCTIVE ACTION! LOSS OF KEYS AND FUNDS MAY RESULT!
	Remove(*Client) error
}

// PrivateAddress is the interface implemented by the two private address
// types: FsAddress, and EsAddress.
type PrivateAddress interface {
	Address

	// PrivateKey returns the ed25519.PrivateKey which can be used for
	// signing data.
	PrivateKey() ed25519.PrivateKey
	// PublicKey returns the ed25519.PublicKey which can be used for
	// verifying signatures.
	PublicKey() ed25519.PublicKey
}

// FAAddress is a Public Factoid Address.
type FAAddress [sha256.Size]byte

// FsAddress is the secret key to a FAAddress.
type FsAddress [sha256.Size]byte

// ECAddress is a Public Entry Credit Address.
type ECAddress [sha256.Size]byte

// EsAddress is the secret key to a ECAddress.
type EsAddress [sha256.Size]byte

// Ensure that the Address and PrivateAddress interfaces are implemented.
var _ Address = FAAddress{}
var _ PrivateAddress = FsAddress{}
var _ Address = ECAddress{}
var _ PrivateAddress = EsAddress{}

// Payload returns adr as a byte array.
func (adr FAAddress) Payload() [sha256.Size]byte {
	return adr
}

// Payload returns adr as a byte array.
func (adr FsAddress) Payload() [sha256.Size]byte {
	return adr
}

// Payload returns adr as a byte array.
func (adr ECAddress) Payload() [sha256.Size]byte {
	return adr
}

// Payload returns adr as a byte array.
func (adr EsAddress) Payload() [sha256.Size]byte {
	return adr
}

// payload returns adr as payload. This is syntactic sugar useful in other
// methods that leverage payload.
func (adr FAAddress) payload() payload {
	return payload(adr)
}
func (adr FsAddress) payload() payload {
	return payload(adr)
}
func (adr ECAddress) payload() payload {
	return payload(adr)
}
func (adr EsAddress) payload() payload {
	return payload(adr)
}

// payloadPtr returns adr as *payload. This is syntactic sugar useful in other
// methods that leverage *payload.
func (adr *FAAddress) payloadPtr() *payload {
	return (*payload)(adr)
}
func (adr *FsAddress) payloadPtr() *payload {
	return (*payload)(adr)
}
func (adr *ECAddress) payloadPtr() *payload {
	return (*payload)(adr)
}
func (adr *EsAddress) payloadPtr() *payload {
	return (*payload)(adr)
}

var (
	faPrefixBytes = [...]byte{0x5f, 0xb1}
	fsPrefixBytes = [...]byte{0x64, 0x78}
	ecPrefixBytes = [...]byte{0x59, 0x2a}
	esPrefixBytes = [...]byte{0x5d, 0xb6}
)

// PrefixBytes returns the two byte prefix for the address type as a byte
// array. Note that the prefix for a given address type is always the same and
// does not depend on the address value. Returns []byte{0x5f, 0xb1}.
func (FAAddress) PrefixBytes() []byte {
	prefix := faPrefixBytes
	return prefix[:]
}

// PrefixBytes returns the two byte prefix for the address type as a byte
// array. Note that the prefix for a given address type is always the same and
// does not depend on the address value. Returns []byte{0x64, 0x78}.
func (FsAddress) PrefixBytes() []byte {
	prefix := fsPrefixBytes
	return prefix[:]
}

// PrefixBytes returns the two byte prefix for the address type as a byte
// array. Note that the prefix for a given address type is always the same and
// does not depend on the address value. Returns []byte{0x59, 0x2a}.
func (ECAddress) PrefixBytes() []byte {
	prefix := ecPrefixBytes
	return prefix[:]
}

// PrefixBytes returns the two byte prefix for the address type as a byte
// array. Note that the prefix for a given address type is always the same and
// does not depend on the address value. Returns []byte{0x5d, 0xb6}.
func (EsAddress) PrefixBytes() []byte {
	prefix := esPrefixBytes
	return prefix[:]
}

const (
	faPrefixStr = "FA"
	fsPrefixStr = "Fs"
	ecPrefixStr = "EC"
	esPrefixStr = "Es"
)

// PrefixString returns the two prefix bytes for the address type as an encoded
// string. Note that the prefix for a given address type is always the same and
// does not depend on the address value. Returns "FA".
func (FAAddress) PrefixString() string {
	return faPrefixStr
}

// PrefixString returns the two prefix bytes for the address type as an encoded
// string. Note that the prefix for a given address type is always the same and
// does not depend on the address value. Returns "Fs".
func (FsAddress) PrefixString() string {
	return fsPrefixStr
}

// PrefixString returns the two prefix bytes for the address type as an encoded
// string. Note that the prefix for a given address type is always the same and
// does not depend on the address value. Returns "EC".
func (ECAddress) PrefixString() string {
	return ecPrefixStr
}

// PrefixString returns the two prefix bytes for the address type as an encoded
// string. Note that the prefix for a given address type is always the same and
// does not depend on the address value. Returns "Es".
func (EsAddress) PrefixString() string {
	return esPrefixStr
}

// String encodes adr into its human readable form: a base58check string with
// adr.PrefixBytes().
func (adr FAAddress) String() string {
	return adr.payload().StringPrefix(adr.PrefixBytes())
}

// String encodes adr into its human readable form: a base58check string with
// adr.PrefixBytes().
func (adr FsAddress) String() string {
	return adr.payload().StringPrefix(adr.PrefixBytes())
}

// String encodes adr into its human readable form: a base58check string with
// adr.PrefixBytes().
func (adr ECAddress) String() string {
	return adr.payload().StringPrefix(adr.PrefixBytes())
}

// String encodes adr into its human readable form: a base58check string with
// adr.PrefixBytes().
func (adr EsAddress) String() string {
	return adr.payload().StringPrefix(adr.PrefixBytes())
}

// MarshalJSON encodes adr as a JSON string using adr.String().
func (adr FAAddress) MarshalJSON() ([]byte, error) {
	return adr.payload().MarshalJSONPrefix(adr.PrefixBytes())
}

// MarshalJSON encodes adr as a JSON string using adr.String().
func (adr FsAddress) MarshalJSON() ([]byte, error) {
	return adr.payload().MarshalJSONPrefix(adr.PrefixBytes())
}

// MarshalJSON encodes adr as a JSON string using adr.String().
func (adr ECAddress) MarshalJSON() ([]byte, error) {
	return adr.payload().MarshalJSONPrefix(adr.PrefixBytes())
}

// MarshalJSON encodes adr as a JSON string using adr.String().
func (adr EsAddress) MarshalJSON() ([]byte, error) {
	return adr.payload().MarshalJSONPrefix(adr.PrefixBytes())
}

const adrStrLen = 52

// NewAddress parses adrStr and returns the correct address type as an Address
// interface. This is useful when the address type isn't known prior to parsing
// adrStr. If the address type is known ahead of time, it is generally better
// to just use the appropriate concrete type.
func NewAddress(adrStr string) (Address, error) {
	if len(adrStr) != adrStrLen {
		return nil, fmt.Errorf("invalid length")
	}
	switch adrStr[:2] {
	case FAAddress{}.PrefixString():
		return NewFAAddress(adrStr)
	case FsAddress{}.PrefixString():
		return NewFsAddress(adrStr)
	case ECAddress{}.PrefixString():
		return NewECAddress(adrStr)
	case EsAddress{}.PrefixString():
		return NewEsAddress(adrStr)
	default:
		return nil, fmt.Errorf("unrecognized prefix")
	}
}

// NewPublicAddress parses adrStr and returns the correct address type as an
// Address interface. If adrStr is not a public address then an "invalid
// prefix" error is returned. This is useful when the address type isn't known
// prior to parsing adrStr, but must be a public address. If the address type
// is known ahead of time, it is generally better to just use the appropriate
// concrete type.
func NewPublicAddress(adrStr string) (Address, error) {
	if len(adrStr) != adrStrLen {
		return nil, fmt.Errorf("invalid length")
	}
	switch adrStr[:2] {
	case FAAddress{}.PrefixString():
		return NewFAAddress(adrStr)
	case ECAddress{}.PrefixString():
		return NewECAddress(adrStr)
	case FsAddress{}.PrefixString():
		fallthrough
	case EsAddress{}.PrefixString():
		return nil, fmt.Errorf("invalid prefix")
	default:
		return nil, fmt.Errorf("unrecognized prefix")
	}
}

// NewPrivateAddress parses adrStr and returns the correct address type as a
// PrivateAddress interface. If adrStr is not a private address then an
// "invalid prefix" error is returned. This is useful when the address type
// isn't known prior to parsing adrStr, but must be a private address. If the
// address type is known ahead of time, it is generally better to just use the
// appropriate concrete type.
func NewPrivateAddress(adrStr string) (PrivateAddress, error) {
	if len(adrStr) != adrStrLen {
		return nil, fmt.Errorf("invalid length")
	}
	switch adrStr[:2] {
	case FsAddress{}.PrefixString():
		return NewFsAddress(adrStr)
	case EsAddress{}.PrefixString():
		return NewEsAddress(adrStr)
	case FAAddress{}.PrefixString():
		fallthrough
	case ECAddress{}.PrefixString():
		return nil, fmt.Errorf("invalid prefix")
	default:
		return nil, fmt.Errorf("unrecognized prefix")
	}
}

// GenerateFsAddress generates a secure random private Factoid address using
// crypto/rand.Random as the source of randomness.
func GenerateFsAddress() (FsAddress, error) {
	return generatePrivKey()
}

// GenerateEsAddress generates a secure random private Entry Credit address
// using crypto/rand.Random as the source of randomness.
func GenerateEsAddress() (EsAddress, error) {
	return generatePrivKey()
}
func generatePrivKey() (key [sha256.Size]byte, err error) {
	var priv ed25519.PrivateKey
	if _, priv, err = ed25519.GenerateKey(rand.Reader); err != nil {
		return
	}
	copy(key[:], priv)
	return key, nil
}

// NewFAAddress attempts to parse adrStr into a new FAAddress.
func NewFAAddress(adrStr string) (adr FAAddress, err error) {
	err = adr.Set(adrStr)
	return
}

// NewFsAddress attempts to parse adrStr into a new FsAddress.
func NewFsAddress(adrStr string) (adr FsAddress, err error) {
	err = adr.Set(adrStr)
	return
}

// NewECAddress attempts to parse adrStr into a new ECAddress.
func NewECAddress(adrStr string) (adr ECAddress, err error) {
	err = adr.Set(adrStr)
	return
}

// NewEsAddress attempts to parse adrStr into a new EsAddress.
func NewEsAddress(adrStr string) (adr EsAddress, err error) {
	err = adr.Set(adrStr)
	return
}

// Set attempts to parse adrStr into adr.
func (adr *FAAddress) Set(adrStr string) error {
	return adr.payloadPtr().SetPrefix(adrStr, adr.PrefixString())
}

// Set attempts to parse adrStr into adr.
func (adr *FsAddress) Set(adrStr string) error {
	return adr.payloadPtr().SetPrefix(adrStr, adr.PrefixString())
}

// Set attempts to parse adrStr into adr.
func (adr *ECAddress) Set(adrStr string) error {
	return adr.payloadPtr().SetPrefix(adrStr, adr.PrefixString())
}

// Set attempts to parse adrStr into adr.
func (adr *EsAddress) Set(adrStr string) error {
	return adr.payloadPtr().SetPrefix(adrStr, adr.PrefixString())
}

// UnmarshalJSON decodes a JSON string with a human readable public Factoid
// address into adr.
func (adr *FAAddress) UnmarshalJSON(data []byte) error {
	return adr.payloadPtr().UnmarshalJSONPrefix(data, adr.PrefixString())
}

// UnmarshalJSON decodes a JSON string with a human readable secret Factoid
// address into adr.
func (adr *FsAddress) UnmarshalJSON(data []byte) error {
	return adr.payloadPtr().UnmarshalJSONPrefix(data, adr.PrefixString())
}

// UnmarshalJSON decodes a JSON string with a human readable public Entry
// Credit address into adr.
func (adr *ECAddress) UnmarshalJSON(data []byte) error {
	return adr.payloadPtr().UnmarshalJSONPrefix(data, adr.PrefixString())
}

// UnmarshalJSON decodes a JSON string with a human readable secret Entry
// Credit address into adr.
func (adr *EsAddress) UnmarshalJSON(data []byte) error {
	return adr.payloadPtr().UnmarshalJSONPrefix(data, adr.PrefixString())
}

// GetPrivateAddress queries factom-walletd for the secret address
// corresponding to adr and returns it as a PrivateAddress.
func (adr FAAddress) GetPrivateAddress(c *Client) (PrivateAddress, error) {
	return adr.GetFsAddress(c)
}

// GetPrivateAddress returns adr as a PrivateAddress.
func (adr FsAddress) GetPrivateAddress(_ *Client) (PrivateAddress, error) {
	return adr, nil
}

// GetPrivateAddress queries factom-walletd for the secret address
// corresponding to adr and returns it as a PrivateAddress.
func (adr ECAddress) GetPrivateAddress(c *Client) (PrivateAddress, error) {
	return adr.GetEsAddress(c)
}

// GetPrivateAddress returns adr as a PrivateAddress.
func (adr EsAddress) GetPrivateAddress(_ *Client) (PrivateAddress, error) {
	return adr, nil
}

// GetFsAddress queries factom-walletd for the FsAddress corresponding to adr.
func (adr FAAddress) GetFsAddress(c *Client) (FsAddress, error) {
	var privAdr FsAddress
	err := c.GetAddress(adr, &privAdr)
	return privAdr, err
}

// GetEsAddress queries factom-walletd for the EsAddress corresponding to adr.
func (adr ECAddress) GetEsAddress(c *Client) (EsAddress, error) {
	var privAdr EsAddress
	err := c.GetAddress(adr, &privAdr)
	return privAdr, err
}

type walletAddress struct{ Address Address }

// GetAddress queries factom-walletd for the privAdr corresponding to pubAdr.
// If the returned error is nil, then privAdr is now populated. Note that
// privAdr must be a pointer to a concrete type implementing PrivateAddress.
func (c *Client) GetAddress(pubAdr Address, privAdr PrivateAddress) error {
	params := walletAddress{Address: pubAdr}
	result := struct{ Secret PrivateAddress }{Secret: privAdr}
	if err := c.WalletdRequest("address", params, &result); err != nil {
		return err
	}
	return nil
}

type walletAddressPublic struct{ Public string }
type walletAddressSecret struct{ Secret string }
type walletAddressesPublic struct{ Addresses []walletAddressPublic }
type walletAddressesSecret struct{ Addresses []walletAddressSecret }

// GetAddresses queries factom-walletd for all public addresses.
func (c *Client) GetAddresses() ([]Address, error) {
	var result walletAddressesPublic
	if err := c.WalletdRequest("all-addresses", nil, &result); err != nil {
		return nil, err
	}
	addresses := make([]Address, 0, len(result.Addresses))
	for _, adrStr := range result.Addresses {
		adr, err := NewAddress(adrStr.Public)
		if err != nil {
			return nil, err
		}
		addresses = append(addresses, adr)
	}
	return addresses, nil
}

// GetPrivateAddresses queries factom-walletd for all private addresses.
func (c *Client) GetPrivateAddresses() ([]PrivateAddress, error) {
	var result walletAddressesSecret
	if err := c.WalletdRequest("all-addresses", nil, &result); err != nil {
		return nil, err
	}
	addresses := make([]PrivateAddress, 0, len(result.Addresses))
	for _, adrStr := range result.Addresses {
		adr, err := NewPrivateAddress(adrStr.Secret)
		if err != nil {
			return nil, err
		}
		addresses = append(addresses, adr)
	}
	return addresses, nil
}

// GetFAAddresses queries factom-walletd for all public Factoid addresses.
func (c *Client) GetFAAddresses() ([]FAAddress, error) {
	var result walletAddressesPublic
	if err := c.WalletdRequest("all-addresses", nil, &result); err != nil {
		return nil, err
	}
	addresses := make([]FAAddress, 0, len(result.Addresses))
	for _, adrStr := range result.Addresses {
		adr, err := NewFAAddress(adrStr.Public)
		if err != nil {
			continue
		}
		addresses = append(addresses, adr)
	}
	return addresses, nil
}

// GetFsAddresses queries factom-walletd for all secret Factoid addresses.
func (c *Client) GetFsAddresses() ([]FsAddress, error) {
	var result walletAddressesSecret
	if err := c.WalletdRequest("all-addresses", nil, &result); err != nil {
		return nil, err
	}
	addresses := make([]FsAddress, 0, len(result.Addresses))
	for _, adrStr := range result.Addresses {
		adr, err := NewFsAddress(adrStr.Secret)
		if err != nil {
			continue
		}
		addresses = append(addresses, adr)
	}
	return addresses, nil
}

// GetECAddresses queries factom-walletd for all public Entry Credit addresses.
func (c *Client) GetECAddresses() ([]ECAddress, error) {
	var result walletAddressesPublic
	if err := c.WalletdRequest("all-addresses", nil, &result); err != nil {
		return nil, err
	}
	addresses := make([]ECAddress, 0, len(result.Addresses))
	for _, adrStr := range result.Addresses {
		adr, err := NewECAddress(adrStr.Public)
		if err != nil {
			continue
		}
		addresses = append(addresses, adr)
	}
	return addresses, nil
}

// GetEsAddresses queries factom-walletd for all secret Entry Credit addresses.
func (c *Client) GetEsAddresses() ([]EsAddress, error) {
	var result walletAddressesSecret
	if err := c.WalletdRequest("all-addresses", nil, &result); err != nil {
		return nil, err
	}
	addresses := make([]EsAddress, 0, len(result.Addresses))
	for _, adrStr := range result.Addresses {
		adr, err := NewEsAddress(adrStr.Secret)
		if err != nil {
			continue
		}
		addresses = append(addresses, adr)
	}
	return addresses, nil
}

// Save adr with factom-walletd.
func (adr FsAddress) Save(c *Client) error {
	return c.SavePrivateAddresses(adr)
}

// Save adr with factom-walletd.
func (adr EsAddress) Save(c *Client) error {
	return c.SavePrivateAddresses(adr)
}

// SavePrivateAddresses saves many adrs with factom-walletd.
func (c *Client) SavePrivateAddresses(adrs ...PrivateAddress) error {
	var params walletAddressesSecret
	params.Addresses = make([]walletAddressSecret, len(adrs))
	for i, adr := range adrs {
		params.Addresses[i].Secret = adr.String()
	}
	if err := c.WalletdRequest("import-addresses", params, nil); err != nil {
		return err
	}
	return nil
}

// GetBalance queries factomd for the Factoid Balance for adr.
func (adr FAAddress) GetBalance(c *Client) (uint64, error) {
	return c.getBalance("factoid-balance", adr)
}

// GetBalance queries factomd for the Factoid Balance for adr.
func (adr FsAddress) GetBalance(c *Client) (uint64, error) {
	return adr.PublicAddress().GetBalance(c)
}

// GetBalance queries factomd for the Entry Credit Balance for adr.
func (adr ECAddress) GetBalance(c *Client) (uint64, error) {
	return c.getBalance("entry-credit-balance", adr)
}

// GetBalance queries factomd for the Entry Credit Balance for adr.
func (adr EsAddress) GetBalance(c *Client) (uint64, error) {
	return adr.PublicAddress().GetBalance(c)
}

type getBalanceParams struct {
	Adr Address `json:"address"`
}
type balanceResult struct{ Balance uint64 }

func (c *Client) getBalance(method string, adr Address) (uint64, error) {
	var result balanceResult
	params := getBalanceParams{Adr: adr}
	if err := c.FactomdRequest(method, params, &result); err != nil {
		return 0, err
	}
	return result.Balance, nil
}

// Remove adr from factom-walletd. WARNING: THIS IS DESTRUCTIVE.
func (adr FAAddress) Remove(c *Client) error {
	return c.RemoveAddress(adr)
}

// Remove adr from factom-walletd. WARNING: THIS IS DESTRUCTIVE.
func (adr FsAddress) Remove(c *Client) error {
	return adr.PublicAddress().Remove(c)
}

// Remove adr from factom-walletd. WARNING: THIS IS DESTRUCTIVE.
func (adr ECAddress) Remove(c *Client) error {
	return c.RemoveAddress(adr)
}

// Remove adr from factom-walletd. WARNING: THIS IS DESTRUCTIVE.
func (adr EsAddress) Remove(c *Client) error {
	return adr.PublicAddress().Remove(c)
}

// RemoveAddress removes adr from factom-walletd. WARNING: THIS IS DESTRUCTIVE.
func (c *Client) RemoveAddress(adr Address) error {
	params := walletAddress{Address: adr.PublicAddress()}
	if err := c.WalletdRequest("remove-address", params, nil); err != nil {
		return err
	}
	return nil
}

// PublicAddress returns adr as an Address.
func (adr FAAddress) PublicAddress() Address {
	return adr
}

// PublicAddress returns the FAAddress corresponding to adr as an Address.
func (adr FsAddress) PublicAddress() Address {
	return adr.FAAddress()
}

// PublicAddress returns adr as an Address.
func (adr ECAddress) PublicAddress() Address {
	return adr
}

// PublicAddress returns the ECAddress corresponding to adr as an Address.
func (adr EsAddress) PublicAddress() Address {
	return adr.ECAddress()
}

// FAAddress returns the FAAddress corresponding to adr.
func (adr FsAddress) FAAddress() FAAddress {
	return adr.RCDHash()
}

// ECAddress returns the ECAddress corresponding to adr.
func (adr EsAddress) ECAddress() (ec ECAddress) {
	copy(ec[:], adr.PublicKey())
	return
}

// RCDHash returns the RCD hash encoded in adr.
func (adr FAAddress) RCDHash() [sha256.Size]byte {
	return adr
}

// RCDHash computes the RCD hash corresponding to adr.
func (adr FsAddress) RCDHash() [sha256.Size]byte {
	return sha256d(adr.RCD())
}

// sha256( sha256( data ) )
func sha256d(data []byte) [sha256.Size]byte {
	hash := sha256.Sum256(data)
	return sha256.Sum256(hash[:])
}

const (
	// RCDType is the magic number identifying the currenctly accepted RCD.
	RCDType byte = 0x01
	// RCDSize is the size of the RCD.
	RCDSize = ed25519.PublicKeySize + 1
	// SignatureSize is the size of the ed25519 signatures.
	SignatureSize = ed25519.SignatureSize
)

// RCD computes the RCD for adr.
func (adr FsAddress) RCD() []byte {
	return append([]byte{RCDType}, adr.PublicKey()[:]...)
}

// PublicKey returns the ed25519.PublicKey for adr.
func (adr ECAddress) PublicKey() ed25519.PublicKey {
	return adr[:]
}

// PublicKey computes the ed25519.PublicKey for adr.
func (adr EsAddress) PublicKey() ed25519.PublicKey {
	return adr.PrivateKey().Public().(ed25519.PublicKey)
}

// PublicKey computes the ed25519.PublicKey for adr.
func (adr FsAddress) PublicKey() ed25519.PublicKey {
	return adr.PrivateKey().Public().(ed25519.PublicKey)
}

// PrivateKey returns the ed25519.PrivateKey for adr.
func (adr FsAddress) PrivateKey() ed25519.PrivateKey {
	return ed25519.NewKeyFromSeed(adr[:])
}

// PrivateKey returns the ed25519.PrivateKey for adr.
func (adr EsAddress) PrivateKey() ed25519.PrivateKey {
	return ed25519.NewKeyFromSeed(adr[:])
}

// Scan implements sql.Scanner for adr using Bytes32.Scan. The FAAddress type
// is not encoded and is assumed.
func (adr *FAAddress) Scan(v interface{}) error {
	return (*Bytes32)(adr).Scan(v)
}

// Value implements driver.Valuer for adr using Bytes32.Value. The FAAddress
// type is not encoded.
func (adr FAAddress) Value() (driver.Value, error) {
	return (Bytes32)(adr).Value()
}

var _ sql.Scanner = &FAAddress{}
var _ driver.Valuer = &FAAddress{}
