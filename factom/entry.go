package factom

import (
	"bytes"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"golang.org/x/crypto/ed25519"
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
	Timestamp Time     `json:"timestamp,omitempty"`
	ChainID   *Bytes32 `json:"chainid,omitempty"`
	Height    uint64   `json:"-"`

	// Entry.Get populates the Content and ExtIDs.
	ExtIDs  []Bytes `json:"extids"`
	Content Bytes   `json:"content"`
}

// IsPopulated returns true if e has already been successfully populated by a
// call to Get. IsPopulated returns false if e.ExtIDs, e.Content, or e.Hash are
// nil, or if e.Timestamp is zero.
func (e Entry) IsPopulated() bool {
	return e.ExtIDs != nil &&
		e.Content != nil &&
		e.ChainID != nil &&
		e.Hash != nil &&
		e.Timestamp.Time != time.Time{}
}

// Get queries factomd for the entry corresponding to e.Hash, which must be not
// nil. After a successful call e.Content, e.ExtIDs, and e.Timestamp will be
// populated.
func (e *Entry) Get(c *Client) error {
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
	if err := c.FactomdRequest("raw-data", params, &result); err != nil {
		return err
	}
	return e.UnmarshalBinary(result.Data)
}

type chainFirstEntryParams struct {
	*Entry `json:"firstentry"`
}
type composeChainParams struct {
	Chain chainFirstEntryParams `json:"chain"`
	EC    ECAddress             `json:"ecpub"`
}
type composeEntryParams struct {
	*Entry `json:"entry"`
	EC     ECAddress `json:"ecpub"`
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

// Create queries factom-walletd to compose and factomd to commit and reveal a
// new Entry or new Chain, if e.ChainID is nil. ec must exist in
// factom-walletd's keystore.
func (e *Entry) Create(c *Client, ec ECAddress) (*Bytes32, error) {
	var params interface{}
	var method string
	if e.ChainID == nil {
		method = "compose-chain"
		params = composeChainParams{
			Chain: chainFirstEntryParams{Entry: e},
			EC:    ec,
		}
	} else {
		method = "compose-entry"
		params = composeEntryParams{Entry: e, EC: ec}
	}
	result := composeResult{}
	if err := c.WalletdRequest(method, params, &result); err != nil {
		return nil, err
	}
	if len(result.Commit.Method) == 0 {
		return nil, fmt.Errorf("Wallet request error: method: %#v", method)
	}

	var commit commitResult
	if err := c.FactomdRequest(result.Commit.Method, result.Commit.Params,
		&commit); err != nil {
		return nil, err
	}
	if err := c.FactomdRequest(result.Reveal.Method, result.Reveal.Params,
		e); err != nil {
		return nil, err
	}
	return commit.TxID, nil
}

// ComposeCreate Composes e locally and then Commit and Reveals it using
// factomd. This does not make any calls to factom-walletd. The Transaction ID
// is returned.
func (e *Entry) ComposeCreate(c *Client, es EsAddress) (*Bytes32, error) {
	var commit, reveal []byte
	commit, reveal, txID, err := e.Compose(es)
	if err != nil {
		return nil, err
	}
	if err := c.Commit(commit); err != nil {
		return txID, err
	}
	if err := c.Reveal(reveal); err != nil {
		return txID, err
	}
	return txID, nil
}

// Commit sends an entry or new chain commit to factomd.
func (c *Client) Commit(commit []byte) error {
	var method string
	switch len(commit) {
	case commitSize:
		method = "commit-entry"
	case chainCommitSize:
		method = "commit-chain"
	default:
		return fmt.Errorf("invalid length")
	}

	params := struct {
		Commit Bytes `json:"message"`
	}{Commit: commit}
	if err := c.FactomdRequest(method, params, nil); err != nil {
		return err
	}
	return nil
}

// Reveal reveals an entry or new chain entry to factomd.
func (c *Client) Reveal(reveal []byte) error {
	params := struct {
		Reveal Bytes `json:"entry"`
	}{Reveal: reveal}
	if err := c.FactomdRequest("reveal-entry", params, nil); err != nil {
		return err
	}
	return nil
}

const (
	commitSize = 1 + // version
		6 + // timestamp
		32 + // entry hash
		1 + // ec cost
		32 + // ec pub
		64 // sig
	chainCommitSize = 1 + // version
		6 + // timestamp
		32 + // chain id hash
		32 + // commit weld
		32 + // entry hash
		1 + // ec cost
		32 + // ec pub
		64 // sig
)

// Compose generates the commit and reveal data required to submit an entry to
// factomd. If e.ChainID is nil, then the ChainID is computed from the e.ExtIDs
// and a new chain commit is created.
func (e *Entry) Compose(es EsAddress) (commit []byte, reveal []byte, txID *Bytes32,
	err error) {
	var newChain bool
	if e.ChainID == nil {
		newChain = true
	}
	reveal, err = e.MarshalBinary() // Populates ChainID and Hash
	if err != nil {
		return
	}

	size := commitSize
	if newChain {
		size = chainCommitSize
	}
	commit = make([]byte, size)

	// Timestamp
	ms := time.Now().
		Add(time.Duration(-rand.Int63n(int64(1*time.Hour)))).
		UnixNano() / 1e6
	buf := bytes.NewBuffer(make([]byte, 0, 8))
	binary.Write(buf, binary.BigEndian, ms)
	i := 1 // Skip version byte
	i += copy(commit[i:], buf.Bytes()[2:])

	if newChain {
		// ChainID Hash
		chainIDHash := sha256d(e.ChainID[:])
		i += copy(commit[i:], chainIDHash[:])

		// Commit Weld sha256d(entryhash | chainid)
		weld := sha256d(append(e.Hash[:], e.ChainID[:]...))
		i += copy(commit[i:], weld[:])
	}

	// Entry Hash
	i += copy(commit[i:], e.Hash[:])

	// Cost
	cost, _ := EntryCost(len(reveal))
	if newChain {
		cost += NewChainCost
	}
	commit[i] = byte(cost)
	i++
	signedDataEndIndex := i
	txID = new(Bytes32)
	*txID = sha256.Sum256(commit[:i])

	// Public Key
	i += copy(commit[i:], es.PublicKey())

	// Signature
	sig := ed25519.Sign(es.PrivateKey(), commit[:signedDataEndIndex])
	copy(commit[i:], sig)

	return
}

// NewChainCost is the fixed added cost of creating a new chain.
const NewChainCost = 10

// EntryCost returns the required Entry Credit cost for an entry with encoded
// length equal to size. An error is returned if size exceeds 10275.
func EntryCost(size int) (int8, error) {
	size -= 35
	if size > 10240 {
		return 0, fmt.Errorf("Entry cannot be larger than 10KB")
	}
	cost := int8(size / 1024)
	if size%1024 > 0 {
		cost++
	}
	if cost < 1 {
		cost = 1
	}
	return cost, nil
}

// MarshalBinary marshals the entry to its binary representation. See
// UnmarshalBinary for encoding details. MarshalBinary populates e.ChainID if
// nil, and always overwrites e.Hash with the computed EntryHash. This is also
// the reveal data.
func (e *Entry) MarshalBinary() ([]byte, error) {
	extIDTotalLen := len(e.ExtIDs) * 2 // Two byte len(ExtID) per ExtID
	for _, extID := range e.ExtIDs {
		extIDTotalLen += len(extID)
	}
	if extIDTotalLen+len(e.Content) > 10240 {
		return nil, fmt.Errorf("Entry cannot be larger than 10KB")
	}
	if e.ChainID == nil {
		e.ChainID = new(Bytes32)
		*e.ChainID = ChainID(e.ExtIDs)
	}
	// Header, version byte 0x00
	data := make([]byte, headerLen+extIDTotalLen+len(e.Content))
	i := 1
	i += copy(data[i:], e.ChainID[:])
	i += copy(data[i:], bigEndian(extIDTotalLen))

	// Payload
	for _, extID := range e.ExtIDs {
		n := len(extID)
		i += copy(data[i:], bigEndian(n))
		i += copy(data[i:], extID)
	}
	copy(data[i:], e.Content)
	// Compute and save entry hash for later use
	e.Hash = new(Bytes32)
	*e.Hash = EntryHash(data)
	return data, nil
}

const (
	headerLen = 1 + // version
		32 + // chain id
		2 // total len

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
	if len(data) > headerLen+10240 {
		return fmt.Errorf("invalid length")
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
func (e Entry) ComputeHash() (Bytes32, error) {
	data, err := e.MarshalBinary()
	return EntryHash(data), err
}

// EntryHash returns the Entry hash of data. Entry's are hashed via:
// sha256(sha512(data) + data).
func EntryHash(data []byte) Bytes32 {
	sum := sha512.Sum512(data)
	saltedSum := make([]byte, len(sum)+len(data))
	i := copy(saltedSum, sum[:])
	copy(saltedSum[i:], data)
	return sha256.Sum256(saltedSum)
}
