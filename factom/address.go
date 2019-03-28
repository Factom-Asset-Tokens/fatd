package factom

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/Factom-Asset-Tokens/base58"

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
	PrefixBytes() [2]byte
	// PrefixString returns the encoded prefix string for the Address.
	PrefixString() string

	// String encodes the address to a base58check string with the
	// appropriate prefix.
	String() string
	// Payload returns the address as a byte array.
	Payload() [sha256.Size]byte

	// PublicAddress returns the corresponding public Address. Public
	// addresses return themselves. Private addresses compute the public
	// address.
	PublicAddress() Address
	// GetPrivateAddress returns the corresponding PrivateAddress. Public
	// addresses query factom-walletd for the private address. Private
	// addresses return themselves.
	GetPrivateAddress(*Client) (PrivateAddress, error)

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

	// Save queries factom-walletd to save this public/private Address pair
	// in its database.
	Save(*Client) error
}

// addressPayload implements helper functions used by all address types.
type addressPayload [sha256.Size]byte

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

// payload returns adr as *addressPayload. This is syntactic sugar useful in
// other methods that leverage addressPayload.
func (adr *FAAddress) payload() *addressPayload {
	return (*addressPayload)(adr)
}
func (adr *FsAddress) payload() *addressPayload {
	return (*addressPayload)(adr)
}
func (adr *ECAddress) payload() *addressPayload {
	return (*addressPayload)(adr)
}
func (adr *EsAddress) payload() *addressPayload {
	return (*addressPayload)(adr)
}

var (
	faPrefixBytes = [2]byte{0x5f, 0xb1}
	fsPrefixBytes = [2]byte{0x64, 0x78}
	ecPrefixBytes = [2]byte{0x59, 0x2a}
	esPrefixBytes = [2]byte{0x5d, 0xb6}
)

// PrefixBytes returns the two byte prefix for the address type as a byte
// array. Note that the prefix for a given address type is always the same and
// does not depend on the address value. Returns [2]byte{0x5f, 0xb1}.
func (FAAddress) PrefixBytes() [2]byte {
	return faPrefixBytes
}

// PrefixBytes returns the two byte prefix for the address type as a byte
// array. Note that the prefix for a given address type is always the same and
// does not depend on the address value. Returns [2]byte{0x64, 0x78}.
func (FsAddress) PrefixBytes() [2]byte {
	return fsPrefixBytes
}

// PrefixBytes returns the two byte prefix for the address type as a byte
// array. Note that the prefix for a given address type is always the same and
// does not depend on the address value. Returns [2]byte{0x59, 0x2a}.
func (ECAddress) PrefixBytes() [2]byte {
	return ecPrefixBytes
}

// PrefixBytes returns the two byte prefix for the address type as a byte
// array. Note that the prefix for a given address type is always the same and
// does not depend on the address value. Returns [2]byte{0x5d, 0xb6}.
func (EsAddress) PrefixBytes() [2]byte {
	return esPrefixBytes
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

// StringPrefix encodes payload as a base58check string with the given prefix.
func (payload addressPayload) StringPrefix(prefix [2]byte) string {
	return base58.CheckEncode(payload[:], prefix[:]...)
}

// MarshalJSON encodes adr as a JSON string using adr.String().
func (adr FAAddress) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%#v", adr.String())), nil
}

// MarshalJSON encodes adr as a JSON string using adr.String().
func (adr FsAddress) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%#v", adr.String())), nil
}

// MarshalJSON encodes adr as a JSON string using adr.String().
func (adr ECAddress) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%#v", adr.String())), nil
}

// MarshalJSON encodes adr as a JSON string using adr.String().
func (adr EsAddress) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%#v", adr.String())), nil
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
	return adr.payload().SetPrefix(adrStr, adr.PrefixString())
}

// Set attempts to parse adrStr into adr.
func (adr *FsAddress) Set(adrStr string) error {
	return adr.payload().SetPrefix(adrStr, adr.PrefixString())
}

// Set attempts to parse adrStr into adr.
func (adr *ECAddress) Set(adrStr string) error {
	return adr.payload().SetPrefix(adrStr, adr.PrefixString())
}

// Set attempts to parse adrStr into adr.
func (adr *EsAddress) Set(adrStr string) error {
	return adr.payload().SetPrefix(adrStr, adr.PrefixString())
}

// SetPrefix attempts to parse adrStr into adr enforcing that adrStr
// starts with prefix if prefix is not empty.
func (payload *addressPayload) SetPrefix(adrStr, prefix string) error {
	if len(adrStr) != 52 {
		return fmt.Errorf("invalid length")
	}
	if len(prefix) > 0 && adrStr[:2] != prefix {
		return fmt.Errorf("invalid prefix")
	}
	b, _, err := base58.CheckDecode(adrStr, 2)
	if err != nil {
		return err
	}
	copy(payload[:], b)
	return nil
}

// UnmarshalJSON decodes a JSON string with a human readable public Factoid
// address into adr.
func (adr *FAAddress) UnmarshalJSON(data []byte) error {
	return adr.payload().UnmarshalJSONPrefix(data, adr.PrefixString())
}

// UnmarshalJSON decodes a JSON string with a human readable secret Factoid
// address into adr.
func (adr *FsAddress) UnmarshalJSON(data []byte) error {
	return adr.payload().UnmarshalJSONPrefix(data, adr.PrefixString())
}

// UnmarshalJSON decodes a JSON string with a human readable public Entry
// Credit address into adr.
func (adr *ECAddress) UnmarshalJSON(data []byte) error {
	return adr.payload().UnmarshalJSONPrefix(data, adr.PrefixString())
}

// UnmarshalJSON decodes a JSON string with a human readable secret Entry
// Credit address into adr.
func (adr *EsAddress) UnmarshalJSON(data []byte) error {
	return adr.payload().UnmarshalJSONPrefix(data, adr.PrefixString())
}

// UnmarshalJSON unmarshals a human readable address JSON string with the given
// prefix.
func (payload *addressPayload) UnmarshalJSONPrefix(data []byte, prefix string) error {
	var adrStr string
	if err := json.Unmarshal(data, &adrStr); err != nil {
		return err
	}
	return payload.SetPrefix(adrStr, prefix)
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

// GetAllAddresses queries factom-walletd for all public addresses.
func (c *Client) GetAllAddresses() ([]Address, error) {
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

// GetAllPrivateAddresses queries factom-walletd for all private addresses.
func (c *Client) GetAllPrivateAddresses() ([]PrivateAddress, error) {
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

// GetAllFAAddresses queries factom-walletd for all public Factoid addresses.
func (c *Client) GetAllFAAddresses() ([]FAAddress, error) {
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

// GetAllFsAddresses queries factom-walletd for all secret Factoid addresses.
func (c *Client) GetAllFsAddresses() ([]FsAddress, error) {
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

// GetAllECAddresses queries factom-walletd for all public Entry Credit
// addresses.
func (c *Client) GetAllECAddresses() ([]ECAddress, error) {
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

// GetAllEsAddresses queries factom-walletd for all secret Entry Credit
// addresses.
func (c *Client) GetAllEsAddresses() ([]EsAddress, error) {
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

// PublicKey computes the ed25519.PublicKey for adr.
func (adr FsAddress) PublicKey() ed25519.PublicKey {
	return adr.PrivateKey().Public().(ed25519.PublicKey)
}

// PublicKey returns the ed25519.PublicKey for adr.
func (adr ECAddress) PublicKey() ed25519.PublicKey {
	return adr[:]
}

// PublicKey computes the ed25519.PublicKey for adr.
func (adr EsAddress) PublicKey() ed25519.PublicKey {
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
