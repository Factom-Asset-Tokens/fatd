package factom

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/json"
	"fmt"
)

// ChainID returns the chain ID for a set of NameIDs.
func ChainID(nameIDs []Bytes) Bytes32 {
	hash := sha256.New()
	for _, id := range nameIDs {
		idSum := sha256.Sum256(id)
		hash.Write(idSum[:])
	}
	c := hash.Sum(nil)
	var chainID Bytes32
	copy(chainID[:], c)
	return chainID
}

// Entry represents a Factom Entry.
type Entry struct {
	// EBlock.Get populates the Hash, Timestamp, ChainID, and Height.
	Hash      *Bytes32 `json:"entryhash,omitempty"`
	Timestamp *Time    `json:"timestamp,omitempty"`
	ChainID   *Bytes32 `json:"chainid,omitempty"`
	Height    uint64   `json:"-"`

	// Entry.Get populates the Content and ExtIDs.
	ExtIDs  []Bytes `json:"extids"`
	Content Bytes   `json:"content"`
}

// IsPopulated returns true if e has already been successfully populated by a
// call to Get. IsPopulated returns false if both e.ExtIDs and e.Content are
// nil.
func (e Entry) IsPopulated() bool {
	return e.ExtIDs != nil || e.Content != nil
}

// Get queries factomd for the entry corresponding to e.Hash.
//
// Get returns any networking or marshaling errors, but not JSON RPC errors. To
// check if the Entry has been successfully populated, call IsPopulated().
func (e *Entry) Get() error {
	// If the Hash is nil then we have nothing to query for.
	if e.Hash == nil {
		return fmt.Errorf("Hash is nil")
	}
	// If the Entry is already populated then there is nothing to do. If
	// the Hash is nil, we cannot populate it anyway.
	if e.IsPopulated() {
		return nil
	}
	params := struct {
		Hash *Bytes32 `json:"hash"`
	}{Hash: e.Hash}
	var result struct {
		Data Bytes `json:"data"`
	}
	if err := factomdRequest("raw-data", params, &result); err != nil {
		return err
	}
	return e.UnmarshalBinary(result.Data)
}

type chainFirstEntryParams struct {
	*Entry `json:"firstentry"`
}
type composeChainParams struct {
	Chain chainFirstEntryParams `json:"chain"`
	ECPub string                `json:"ecpub"`
}
type composeEntryParams struct {
	*Entry `json:"entry"`
	ECPub  string `json:"ecpub"`
}

type composeJRPC struct {
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}
type composeResult struct {
	Commit composeJRPC `json:"commit"`
	Reveal composeJRPC `json:"reveal"`
}
type commitResult struct {
	TxID *Bytes32
}

func (e *Entry) Create(ecpub string) (*Bytes32, error) {
	var params interface{}
	var method string
	if e.ChainID == nil {
		method = "compose-chain"
		params = composeChainParams{
			Chain: chainFirstEntryParams{Entry: e},
			ECPub: ecpub,
		}
	} else {
		method = "compose-entry"
		params = composeEntryParams{
			Entry: e,
			ECPub: ecpub,
		}
	}
	result := composeResult{}
	if err := walletRequest(method, params, &result); err != nil {
		return nil, err
	}
	if len(result.Commit.Method) == 0 {
		return nil, fmt.Errorf("Wallet request error: method: %#v", method)
	}

	var commit commitResult
	if err := factomdRequest(result.Commit.Method, result.Commit.Params,
		&commit); err != nil {
		return nil, err
	}
	if err := factomdRequest(result.Reveal.Method, result.Reveal.Params,
		e); err != nil {
		return nil, err
	}
	return commit.TxID, nil
}

// MarshalBinary marshals the entry to its binary representation. See
// UnmarshalBinary for encoding details.
func (e Entry) MarshalBinary() []byte {
	extIDTotalLen := len(e.ExtIDs) * 2 // Two byte len(ExtID) per ExtID
	for _, extID := range e.ExtIDs {
		extIDTotalLen += len(extID)
	}
	// Header, version byte 0x00
	data := make([]byte, 1, headerLen+extIDTotalLen+len(e.Content))
	data = append(data, e.ChainID[:]...)
	data = append(data, bigEndian(extIDTotalLen)...)

	// Payload
	for _, extID := range e.ExtIDs {
		n := len(extID)
		data = append(data, bigEndian(n)...)
		data = append(data, extID...)
	}
	return append(data, e.Content...)
}

const (
	// Version byte, Chain ID, ExtIDs Total Encoded Length
	headerLen = len([...]byte{0x00}) + len(Bytes32{}) + len([...]byte{0x00, 0x00})
)

// UnmarshalBinary unmarshals raw entry data. It does not populate the
// Entry.Hash. Entries are encoded as follows and use big endian uint16:
//
// [Version byte (0x00)] +
// [ChainID (Bytes32)] +
// [Total ExtID encoded length (uint16)] +
// [ExtID 0 length (uint16)] + [ExtID 0 (Bytes)] +
// ... +
// [ExtID X length (uint16)] + [ExtID X (Bytes)] +
// [Content (Bytes)]
//
// https://github.com/FactomProject/FactomDocs/blob/master/factomDataStructureDetails.md#entry
func (e *Entry) UnmarshalBinary(data []byte) error {
	if len(data) < headerLen {
		return fmt.Errorf("insufficient length")
	}
	if data[0] != 0x00 {
		return fmt.Errorf("invalid version byte")
	}
	chainID := data[1:33]
	extIDTotalLen := parseBigEndian(data[33:35])
	if extIDTotalLen == 1 || headerLen+extIDTotalLen > len(data) {
		return fmt.Errorf("invalid ExtIDs length")
	}

	extIDs := []Bytes{}
	pos := headerLen
	for pos < headerLen+extIDTotalLen {
		extIDLen := parseBigEndian(data[pos : pos+2])
		if pos+2+extIDLen > headerLen+extIDTotalLen {
			return fmt.Errorf("error parsing ExtIDs")
		}
		pos += 2
		extIDs = append(extIDs, Bytes(data[pos:pos+extIDLen]))
		pos += extIDLen
	}
	e.Content = data[pos:]
	e.ExtIDs = extIDs
	e.ChainID = NewBytes32(chainID)
	return nil
}
func bigEndian(x int) []byte {
	return []byte{byte(x >> 8), byte(x)}
}
func parseBigEndian(data []byte) int {
	return int(data[0])<<8 + int(data[1])
}

// ComputeHash returns the Entry's hash as computed by hashing the binary
// representation of the Entry.
func (e Entry) ComputeHash() Bytes32 {
	data := e.MarshalBinary()
	return EntryHash(data)
}

// EntryHash returns the Entry hash of data. Entry's are hashed via:
// sha256(sha512(data) + data).
func EntryHash(data []byte) Bytes32 {
	sum := sha512.Sum512(data)
	return sha256.Sum256(append(sum[:], data...))

}
