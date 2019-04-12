package factom

import (
	"crypto/sha256"
	"database/sql/driver"

	"golang.org/x/crypto/ed25519"
)

// Code generated DO NOT EDIT
// Defines IDKeys ID1Key - ID4Key and corresponding SKKeys SK1Key - SK4Key.

var (
	id1PrefixBytes = [...]byte{0x3f, 0xbe, 0xba}
	id2PrefixBytes = [...]byte{0x3f, 0xbe, 0xd8}
	id3PrefixBytes = [...]byte{0x3f, 0xbe, 0xf6}
	id4PrefixBytes = [...]byte{0x3f, 0xbf, 0x14}

	sk1PrefixBytes = [...]byte{0x4d, 0xb6, 0xc9}
	sk2PrefixBytes = [...]byte{0x4d, 0xb6, 0xe7}
	sk3PrefixBytes = [...]byte{0x4d, 0xb7, 0x05}
	sk4PrefixBytes = [...]byte{0x4d, 0xb7, 0x23}
)

const (
	id1PrefixStr = "id1"
	id2PrefixStr = "id2"
	id3PrefixStr = "id3"
	id4PrefixStr = "id4"

	sk1PrefixStr = "sk1"
	sk2PrefixStr = "sk2"
	sk3PrefixStr = "sk3"
	sk4PrefixStr = "sk4"
)

var (
	_ IDKey = ID1Key{}
	_ IDKey = ID2Key{}
	_ IDKey = ID3Key{}
	_ IDKey = ID4Key{}

	_ SKKey = SK1Key{}
	_ SKKey = SK2Key{}
	_ SKKey = SK3Key{}
	_ SKKey = SK4Key{}
)

// ID1Key is the id1 public key for an identity.
type ID1Key [sha256.Size]byte

// SK1Key is the sk1 secret key for an identity.
type SK1Key [sha256.Size]byte

// Payload returns key as a byte array.
func (key ID1Key) Payload() [sha256.Size]byte {
	return key
}

// Payload returns key as a byte array.
func (key SK1Key) Payload() [sha256.Size]byte {
	return key
}

// payload returns adr as payload. This is syntactic sugar useful in other
// methods that leverage payload.
func (key ID1Key) payload() payload {
	return payload(key)
}
func (key SK1Key) payload() payload {
	return payload(key)
}

// payloadPtr returns adr as *payload. This is syntactic sugar useful in other
// methods that leverage *payload.
func (key *ID1Key) payloadPtr() *payload {
	return (*payload)(key)
}
func (key *SK1Key) payloadPtr() *payload {
	return (*payload)(key)
}

// PrefixBytes returns the two byte prefix for the address type as a byte
// array. Note that the prefix for a given address type is always the same and
// does not depend on the address value. Returns []byte{0x3f, 0xbe, 0xba}.
func (ID1Key) PrefixBytes() []byte {
	prefix := id1PrefixBytes
	return prefix[:]
}

// PrefixBytes returns the two byte prefix for the address type as a byte
// array. Note that the prefix for a given address type is always the same and
// does not depend on the address value. Returns []byte{0x4d, 0xb6, 0xc9}.
func (SK1Key) PrefixBytes() []byte {
	prefix := sk1PrefixBytes
	return prefix[:]
}

// PrefixString returns the two prefix bytes for the address type as an encoded
// string. Note that the prefix for a given address type is always the same and
// does not depend on the address value. Returns "id1".
func (ID1Key) PrefixString() string {
	return id1PrefixStr
}

// PrefixString returns the two prefix bytes for the address type as an encoded
// string. Note that the prefix for a given address type is always the same and
// does not depend on the address value. Returns "sk1".
func (SK1Key) PrefixString() string {
	return sk1PrefixStr
}

// String encodes key into its human readable form: a base58check string with
// key.PrefixBytes().
func (key ID1Key) String() string {
	return key.payload().StringPrefix(key.PrefixBytes())
}

// String encodes key into its human readable form: a base58check string with
// key.PrefixBytes().
func (key SK1Key) String() string {
	return key.payload().StringPrefix(key.PrefixBytes())
}

// MarshalJSON encodes key as a JSON string using key.String().
func (key ID1Key) MarshalJSON() ([]byte, error) {
	return key.payload().MarshalJSONPrefix(key.PrefixBytes())
}

// MarshalJSON encodes key as a JSON string using key.String().
func (key SK1Key) MarshalJSON() ([]byte, error) {
	return key.payload().MarshalJSONPrefix(key.PrefixBytes())
}

// NewID1Key attempts to parse keyStr into a new ID1Key.
func NewID1Key(keyStr string) (key ID1Key, err error) {
	err = key.Set(keyStr)
	return
}

// NewSK1Key attempts to parse keyStr into a new SK1Key.
func NewSK1Key(keyStr string) (key SK1Key, err error) {
	err = key.Set(keyStr)
	return
}

// GenerateSK1Key generates a secure random private Entry Credit address using
// crypto/rand.Random as the source of randomness.
func GenerateSK1Key() (SK1Key, error) {
	return generatePrivKey()
}

// Set attempts to parse keyStr into key.
func (key *ID1Key) Set(keyStr string) error {
	return key.payloadPtr().SetPrefix(keyStr, key.PrefixString())
}

// Set attempts to parse keyStr into key.
func (key *SK1Key) Set(keyStr string) error {
	return key.payloadPtr().SetPrefix(keyStr, key.PrefixString())
}

// UnmarshalJSON decodes a JSON string with a human readable id1 key into key.
func (key *ID1Key) UnmarshalJSON(data []byte) error {
	return key.payloadPtr().UnmarshalJSONPrefix(data, key.PrefixString())
}

// UnmarshalJSON decodes a JSON string with a human readable sk1 key into key.
func (key *SK1Key) UnmarshalJSON(data []byte) error {
	return key.payloadPtr().UnmarshalJSONPrefix(data, key.PrefixString())
}

// IDKey returns key as an IDKey.
func (key ID1Key) IDKey() IDKey {
	return key
}

// IDKey returns the ID1Key corresponding to key as an IDKey.
func (key SK1Key) IDKey() IDKey {
	return key.ID1Key()
}

// SKKey returns key as an SKKey.
func (key SK1Key) SKKey() SKKey {
	return key
}

// ID1Key computes the ID1Key corresponding to key.
func (key SK1Key) ID1Key() ID1Key {
	return key.RCDHash()
}

// RCDHash returns the RCD hash encoded in key.
func (key ID1Key) RCDHash() [sha256.Size]byte {
	return key
}

// RCDHash computes the RCD hash corresponding to key.
func (key SK1Key) RCDHash() [sha256.Size]byte {
	return sha256d(key.RCD())
}

// RCD computes the RCD for key.
func (key SK1Key) RCD() []byte {
	return append([]byte{RCDType}, key.PublicKey()[:]...)
}

// PublicKey computes the ed25519.PublicKey for key.
func (key SK1Key) PublicKey() ed25519.PublicKey {
	return key.PrivateKey().Public().(ed25519.PublicKey)
}

// PrivateKey returns the ed25519.PrivateKey for key.
func (key SK1Key) PrivateKey() ed25519.PrivateKey {
	return ed25519.NewKeyFromSeed(key[:])
}

// Scan implements sql.Scanner for key using Bytes32.Scan. The ID1Key type is
// not encoded and is assumed.
func (key *ID1Key) Scan(v interface{}) error {
	return (*Bytes32)(key).Scan(v)
}

// Value implements driver.Valuer for key using Bytes32.Value. The ID1Key type
// is not encoded.
func (key ID1Key) Value() (driver.Value, error) {
	return (Bytes32)(key).Value()
}

// ID2Key is the id2 public key for an identity.
type ID2Key [sha256.Size]byte

// SK2Key is the sk2 secret key for an identity.
type SK2Key [sha256.Size]byte

// Payload returns key as a byte array.
func (key ID2Key) Payload() [sha256.Size]byte {
	return key
}

// Payload returns key as a byte array.
func (key SK2Key) Payload() [sha256.Size]byte {
	return key
}

// payload returns adr as payload. This is syntactic sugar useful in other
// methods that leverage payload.
func (key ID2Key) payload() payload {
	return payload(key)
}
func (key SK2Key) payload() payload {
	return payload(key)
}

// payloadPtr returns adr as *payload. This is syntactic sugar useful in other
// methods that leverage *payload.
func (key *ID2Key) payloadPtr() *payload {
	return (*payload)(key)
}
func (key *SK2Key) payloadPtr() *payload {
	return (*payload)(key)
}

// PrefixBytes returns the two byte prefix for the address type as a byte
// array. Note that the prefix for a given address type is always the same and
// does not depend on the address value. Returns []byte{0x3f, 0xbe, 0xd8}.
func (ID2Key) PrefixBytes() []byte {
	prefix := id2PrefixBytes
	return prefix[:]
}

// PrefixBytes returns the two byte prefix for the address type as a byte
// array. Note that the prefix for a given address type is always the same and
// does not depend on the address value. Returns []byte{0x4d, 0xb6, 0xe7}.
func (SK2Key) PrefixBytes() []byte {
	prefix := sk2PrefixBytes
	return prefix[:]
}

// PrefixString returns the two prefix bytes for the address type as an encoded
// string. Note that the prefix for a given address type is always the same and
// does not depend on the address value. Returns "id2".
func (ID2Key) PrefixString() string {
	return id2PrefixStr
}

// PrefixString returns the two prefix bytes for the address type as an encoded
// string. Note that the prefix for a given address type is always the same and
// does not depend on the address value. Returns "sk2".
func (SK2Key) PrefixString() string {
	return sk2PrefixStr
}

// String encodes key into its human readable form: a base58check string with
// key.PrefixBytes().
func (key ID2Key) String() string {
	return key.payload().StringPrefix(key.PrefixBytes())
}

// String encodes key into its human readable form: a base58check string with
// key.PrefixBytes().
func (key SK2Key) String() string {
	return key.payload().StringPrefix(key.PrefixBytes())
}

// MarshalJSON encodes key as a JSON string using key.String().
func (key ID2Key) MarshalJSON() ([]byte, error) {
	return key.payload().MarshalJSONPrefix(key.PrefixBytes())
}

// MarshalJSON encodes key as a JSON string using key.String().
func (key SK2Key) MarshalJSON() ([]byte, error) {
	return key.payload().MarshalJSONPrefix(key.PrefixBytes())
}

// NewID2Key attempts to parse keyStr into a new ID2Key.
func NewID2Key(keyStr string) (key ID2Key, err error) {
	err = key.Set(keyStr)
	return
}

// NewSK2Key attempts to parse keyStr into a new SK2Key.
func NewSK2Key(keyStr string) (key SK2Key, err error) {
	err = key.Set(keyStr)
	return
}

// GenerateSK2Key generates a secure random private Entry Credit address using
// crypto/rand.Random as the source of randomness.
func GenerateSK2Key() (SK2Key, error) {
	return generatePrivKey()
}

// Set attempts to parse keyStr into key.
func (key *ID2Key) Set(keyStr string) error {
	return key.payloadPtr().SetPrefix(keyStr, key.PrefixString())
}

// Set attempts to parse keyStr into key.
func (key *SK2Key) Set(keyStr string) error {
	return key.payloadPtr().SetPrefix(keyStr, key.PrefixString())
}

// UnmarshalJSON decodes a JSON string with a human readable id2 key into key.
func (key *ID2Key) UnmarshalJSON(data []byte) error {
	return key.payloadPtr().UnmarshalJSONPrefix(data, key.PrefixString())
}

// UnmarshalJSON decodes a JSON string with a human readable sk2 key into key.
func (key *SK2Key) UnmarshalJSON(data []byte) error {
	return key.payloadPtr().UnmarshalJSONPrefix(data, key.PrefixString())
}

// IDKey returns key as an IDKey.
func (key ID2Key) IDKey() IDKey {
	return key
}

// IDKey returns the ID2Key corresponding to key as an IDKey.
func (key SK2Key) IDKey() IDKey {
	return key.ID2Key()
}

// SKKey returns key as an SKKey.
func (key SK2Key) SKKey() SKKey {
	return key
}

// ID2Key computes the ID2Key corresponding to key.
func (key SK2Key) ID2Key() ID2Key {
	return key.RCDHash()
}

// RCDHash returns the RCD hash encoded in key.
func (key ID2Key) RCDHash() [sha256.Size]byte {
	return key
}

// RCDHash computes the RCD hash corresponding to key.
func (key SK2Key) RCDHash() [sha256.Size]byte {
	return sha256d(key.RCD())
}

// RCD computes the RCD for key.
func (key SK2Key) RCD() []byte {
	return append([]byte{RCDType}, key.PublicKey()[:]...)
}

// PublicKey computes the ed25519.PublicKey for key.
func (key SK2Key) PublicKey() ed25519.PublicKey {
	return key.PrivateKey().Public().(ed25519.PublicKey)
}

// PrivateKey returns the ed25519.PrivateKey for key.
func (key SK2Key) PrivateKey() ed25519.PrivateKey {
	return ed25519.NewKeyFromSeed(key[:])
}

// Scan implements sql.Scanner for key using Bytes32.Scan. The ID2Key type is
// not encoded and is assumed.
func (key *ID2Key) Scan(v interface{}) error {
	return (*Bytes32)(key).Scan(v)
}

// Value implements driver.Valuer for key using Bytes32.Value. The ID2Key type
// is not encoded.
func (key ID2Key) Value() (driver.Value, error) {
	return (Bytes32)(key).Value()
}

// ID3Key is the id3 public key for an identity.
type ID3Key [sha256.Size]byte

// SK3Key is the sk3 secret key for an identity.
type SK3Key [sha256.Size]byte

// Payload returns key as a byte array.
func (key ID3Key) Payload() [sha256.Size]byte {
	return key
}

// Payload returns key as a byte array.
func (key SK3Key) Payload() [sha256.Size]byte {
	return key
}

// payload returns adr as payload. This is syntactic sugar useful in other
// methods that leverage payload.
func (key ID3Key) payload() payload {
	return payload(key)
}
func (key SK3Key) payload() payload {
	return payload(key)
}

// payloadPtr returns adr as *payload. This is syntactic sugar useful in other
// methods that leverage *payload.
func (key *ID3Key) payloadPtr() *payload {
	return (*payload)(key)
}
func (key *SK3Key) payloadPtr() *payload {
	return (*payload)(key)
}

// PrefixBytes returns the two byte prefix for the address type as a byte
// array. Note that the prefix for a given address type is always the same and
// does not depend on the address value. Returns []byte{0x3f, 0xbe, 0xf6}.
func (ID3Key) PrefixBytes() []byte {
	prefix := id3PrefixBytes
	return prefix[:]
}

// PrefixBytes returns the two byte prefix for the address type as a byte
// array. Note that the prefix for a given address type is always the same and
// does not depend on the address value. Returns []byte{0x4d, 0xb7, 0x05}.
func (SK3Key) PrefixBytes() []byte {
	prefix := sk3PrefixBytes
	return prefix[:]
}

// PrefixString returns the two prefix bytes for the address type as an encoded
// string. Note that the prefix for a given address type is always the same and
// does not depend on the address value. Returns "id3".
func (ID3Key) PrefixString() string {
	return id3PrefixStr
}

// PrefixString returns the two prefix bytes for the address type as an encoded
// string. Note that the prefix for a given address type is always the same and
// does not depend on the address value. Returns "sk3".
func (SK3Key) PrefixString() string {
	return sk3PrefixStr
}

// String encodes key into its human readable form: a base58check string with
// key.PrefixBytes().
func (key ID3Key) String() string {
	return key.payload().StringPrefix(key.PrefixBytes())
}

// String encodes key into its human readable form: a base58check string with
// key.PrefixBytes().
func (key SK3Key) String() string {
	return key.payload().StringPrefix(key.PrefixBytes())
}

// MarshalJSON encodes key as a JSON string using key.String().
func (key ID3Key) MarshalJSON() ([]byte, error) {
	return key.payload().MarshalJSONPrefix(key.PrefixBytes())
}

// MarshalJSON encodes key as a JSON string using key.String().
func (key SK3Key) MarshalJSON() ([]byte, error) {
	return key.payload().MarshalJSONPrefix(key.PrefixBytes())
}

// NewID3Key attempts to parse keyStr into a new ID3Key.
func NewID3Key(keyStr string) (key ID3Key, err error) {
	err = key.Set(keyStr)
	return
}

// NewSK3Key attempts to parse keyStr into a new SK3Key.
func NewSK3Key(keyStr string) (key SK3Key, err error) {
	err = key.Set(keyStr)
	return
}

// GenerateSK3Key generates a secure random private Entry Credit address using
// crypto/rand.Random as the source of randomness.
func GenerateSK3Key() (SK3Key, error) {
	return generatePrivKey()
}

// Set attempts to parse keyStr into key.
func (key *ID3Key) Set(keyStr string) error {
	return key.payloadPtr().SetPrefix(keyStr, key.PrefixString())
}

// Set attempts to parse keyStr into key.
func (key *SK3Key) Set(keyStr string) error {
	return key.payloadPtr().SetPrefix(keyStr, key.PrefixString())
}

// UnmarshalJSON decodes a JSON string with a human readable id3 key into key.
func (key *ID3Key) UnmarshalJSON(data []byte) error {
	return key.payloadPtr().UnmarshalJSONPrefix(data, key.PrefixString())
}

// UnmarshalJSON decodes a JSON string with a human readable sk3 key into key.
func (key *SK3Key) UnmarshalJSON(data []byte) error {
	return key.payloadPtr().UnmarshalJSONPrefix(data, key.PrefixString())
}

// IDKey returns key as an IDKey.
func (key ID3Key) IDKey() IDKey {
	return key
}

// IDKey returns the ID3Key corresponding to key as an IDKey.
func (key SK3Key) IDKey() IDKey {
	return key.ID3Key()
}

// SKKey returns key as an SKKey.
func (key SK3Key) SKKey() SKKey {
	return key
}

// ID3Key computes the ID3Key corresponding to key.
func (key SK3Key) ID3Key() ID3Key {
	return key.RCDHash()
}

// RCDHash returns the RCD hash encoded in key.
func (key ID3Key) RCDHash() [sha256.Size]byte {
	return key
}

// RCDHash computes the RCD hash corresponding to key.
func (key SK3Key) RCDHash() [sha256.Size]byte {
	return sha256d(key.RCD())
}

// RCD computes the RCD for key.
func (key SK3Key) RCD() []byte {
	return append([]byte{RCDType}, key.PublicKey()[:]...)
}

// PublicKey computes the ed25519.PublicKey for key.
func (key SK3Key) PublicKey() ed25519.PublicKey {
	return key.PrivateKey().Public().(ed25519.PublicKey)
}

// PrivateKey returns the ed25519.PrivateKey for key.
func (key SK3Key) PrivateKey() ed25519.PrivateKey {
	return ed25519.NewKeyFromSeed(key[:])
}

// Scan implements sql.Scanner for key using Bytes32.Scan. The ID3Key type is
// not encoded and is assumed.
func (key *ID3Key) Scan(v interface{}) error {
	return (*Bytes32)(key).Scan(v)
}

// Value implements driver.Valuer for key using Bytes32.Value. The ID3Key type
// is not encoded.
func (key ID3Key) Value() (driver.Value, error) {
	return (Bytes32)(key).Value()
}

// ID4Key is the id4 public key for an identity.
type ID4Key [sha256.Size]byte

// SK4Key is the sk4 secret key for an identity.
type SK4Key [sha256.Size]byte

// Payload returns key as a byte array.
func (key ID4Key) Payload() [sha256.Size]byte {
	return key
}

// Payload returns key as a byte array.
func (key SK4Key) Payload() [sha256.Size]byte {
	return key
}

// payload returns adr as payload. This is syntactic sugar useful in other
// methods that leverage payload.
func (key ID4Key) payload() payload {
	return payload(key)
}
func (key SK4Key) payload() payload {
	return payload(key)
}

// payloadPtr returns adr as *payload. This is syntactic sugar useful in other
// methods that leverage *payload.
func (key *ID4Key) payloadPtr() *payload {
	return (*payload)(key)
}
func (key *SK4Key) payloadPtr() *payload {
	return (*payload)(key)
}

// PrefixBytes returns the two byte prefix for the address type as a byte
// array. Note that the prefix for a given address type is always the same and
// does not depend on the address value. Returns []byte{0x3f, 0xbf, 0x14}.
func (ID4Key) PrefixBytes() []byte {
	prefix := id4PrefixBytes
	return prefix[:]
}

// PrefixBytes returns the two byte prefix for the address type as a byte
// array. Note that the prefix for a given address type is always the same and
// does not depend on the address value. Returns []byte{0x4d, 0xb7, 0x23}.
func (SK4Key) PrefixBytes() []byte {
	prefix := sk4PrefixBytes
	return prefix[:]
}

// PrefixString returns the two prefix bytes for the address type as an encoded
// string. Note that the prefix for a given address type is always the same and
// does not depend on the address value. Returns "id4".
func (ID4Key) PrefixString() string {
	return id4PrefixStr
}

// PrefixString returns the two prefix bytes for the address type as an encoded
// string. Note that the prefix for a given address type is always the same and
// does not depend on the address value. Returns "sk4".
func (SK4Key) PrefixString() string {
	return sk4PrefixStr
}

// String encodes key into its human readable form: a base58check string with
// key.PrefixBytes().
func (key ID4Key) String() string {
	return key.payload().StringPrefix(key.PrefixBytes())
}

// String encodes key into its human readable form: a base58check string with
// key.PrefixBytes().
func (key SK4Key) String() string {
	return key.payload().StringPrefix(key.PrefixBytes())
}

// MarshalJSON encodes key as a JSON string using key.String().
func (key ID4Key) MarshalJSON() ([]byte, error) {
	return key.payload().MarshalJSONPrefix(key.PrefixBytes())
}

// MarshalJSON encodes key as a JSON string using key.String().
func (key SK4Key) MarshalJSON() ([]byte, error) {
	return key.payload().MarshalJSONPrefix(key.PrefixBytes())
}

// NewID4Key attempts to parse keyStr into a new ID4Key.
func NewID4Key(keyStr string) (key ID4Key, err error) {
	err = key.Set(keyStr)
	return
}

// NewSK4Key attempts to parse keyStr into a new SK4Key.
func NewSK4Key(keyStr string) (key SK4Key, err error) {
	err = key.Set(keyStr)
	return
}

// GenerateSK4Key generates a secure random private Entry Credit address using
// crypto/rand.Random as the source of randomness.
func GenerateSK4Key() (SK4Key, error) {
	return generatePrivKey()
}

// Set attempts to parse keyStr into key.
func (key *ID4Key) Set(keyStr string) error {
	return key.payloadPtr().SetPrefix(keyStr, key.PrefixString())
}

// Set attempts to parse keyStr into key.
func (key *SK4Key) Set(keyStr string) error {
	return key.payloadPtr().SetPrefix(keyStr, key.PrefixString())
}

// UnmarshalJSON decodes a JSON string with a human readable id4 key into key.
func (key *ID4Key) UnmarshalJSON(data []byte) error {
	return key.payloadPtr().UnmarshalJSONPrefix(data, key.PrefixString())
}

// UnmarshalJSON decodes a JSON string with a human readable sk4 key into key.
func (key *SK4Key) UnmarshalJSON(data []byte) error {
	return key.payloadPtr().UnmarshalJSONPrefix(data, key.PrefixString())
}

// IDKey returns key as an IDKey.
func (key ID4Key) IDKey() IDKey {
	return key
}

// IDKey returns the ID4Key corresponding to key as an IDKey.
func (key SK4Key) IDKey() IDKey {
	return key.ID4Key()
}

// SKKey returns key as an SKKey.
func (key SK4Key) SKKey() SKKey {
	return key
}

// ID4Key computes the ID4Key corresponding to key.
func (key SK4Key) ID4Key() ID4Key {
	return key.RCDHash()
}

// RCDHash returns the RCD hash encoded in key.
func (key ID4Key) RCDHash() [sha256.Size]byte {
	return key
}

// RCDHash computes the RCD hash corresponding to key.
func (key SK4Key) RCDHash() [sha256.Size]byte {
	return sha256d(key.RCD())
}

// RCD computes the RCD for key.
func (key SK4Key) RCD() []byte {
	return append([]byte{RCDType}, key.PublicKey()[:]...)
}

// PublicKey computes the ed25519.PublicKey for key.
func (key SK4Key) PublicKey() ed25519.PublicKey {
	return key.PrivateKey().Public().(ed25519.PublicKey)
}

// PrivateKey returns the ed25519.PrivateKey for key.
func (key SK4Key) PrivateKey() ed25519.PrivateKey {
	return ed25519.NewKeyFromSeed(key[:])
}

// Scan implements sql.Scanner for key using Bytes32.Scan. The ID4Key type is
// not encoded and is assumed.
func (key *ID4Key) Scan(v interface{}) error {
	return (*Bytes32)(key).Scan(v)
}

// Value implements driver.Valuer for key using Bytes32.Value. The ID4Key type
// is not encoded.
func (key ID4Key) Value() (driver.Value, error) {
	return (Bytes32)(key).Value()
}
