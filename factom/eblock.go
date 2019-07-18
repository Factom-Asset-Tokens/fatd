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
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"time"

	merkle "github.com/AdamSLevy/go-merkle"
	jrpc "github.com/AdamSLevy/jsonrpc2/v11"
)

// EBlock represents a Factom Entry Block.
type EBlock struct {
	// DBlock.Get populates the ChainID, KeyMR, and Height.
	ChainID *Bytes32 `json:"chainid,omitempty"`
	KeyMR   *Bytes32 `json:"keymr,omitempty"`

	PrevKeyMR *Bytes32 `json:"-"`

	PrevFullHash *Bytes32 `json:"-"`
	BodyMR       *Bytes32 `json:"-"`

	Height      uint32 `json:"-"`
	Sequence    uint32 `json:"-"`
	ObjectCount uint32 `json:"-"`

	Timestamp time.Time `json:"-"`

	// EBlock.Get populates the EBlockHeader.PrevKeyMR and the Entries with
	// their Hash and Timestamp.
	Entries []Entry `json:"-"`
}

func (eb EBlock) IsPopulated() bool {
	return len(eb.Entries) > 0 &&
		eb.ChainID != nil &&
		eb.PrevKeyMR != nil &&
		eb.PrevFullHash != nil &&
		eb.BodyMR != nil &&
		eb.ObjectCount > 1
}

// Get queries factomd for the Entry Block corresponding to eb.KeyMR, if not
// nil, and otherwise the Entry Block chain head for eb.ChainID. Either
// eb.KeyMR or eb.ChainID must be not nil or else Get will fail to populate the
// EBlock. After a successful call, EBlockHeader and Entries will be populated.
// Each Entry will be populated with its Hash, Timestamp, ChainID, and Height,
// but not its Content or ExtIDs. Call Get on the individual Entries to
// populate their Content and ExtIDs.
func (eb *EBlock) Get(c *Client) error {
	// If the EBlock is already populated then there is nothing to do.
	if eb.IsPopulated() {
		return nil
	}

	// If we don't have a KeyMR, fetch the chain head KeyMR.
	if eb.KeyMR == nil {
		// If the KeyMR and ChainID are both nil we have nothing to
		// query for.
		if eb.ChainID == nil {
			return fmt.Errorf("no KeyMR or ChainID specified")
		}
		if err := eb.GetChainHead(c); err != nil {
			return err
		}
	}

	// Make RPC request for this Entry Block.
	params := struct {
		KeyMR *Bytes32 `json:"hash"`
	}{KeyMR: eb.KeyMR}
	var result struct {
		Data Bytes `json:"data"`
	}
	if err := c.FactomdRequest("raw-data", params, &result); err != nil {
		return err
	}
	height := eb.Height
	if err := eb.UnmarshalBinary(result.Data); err != nil {
		return err
	}
	// Verify height if it was initialized
	if height > 0 && height != eb.Height {
		return fmt.Errorf("height does not match")
	}
	keyMR, err := eb.ComputeKeyMR()
	if err != nil {
		return err
	}
	if *eb.KeyMR != keyMR {
		return fmt.Errorf("invalid key merkle root")
	}
	return nil
}

// GetChainHead queries factomd for the latest eb.KeyMR for chain eb.ChainID.
func (eb *EBlock) GetChainHead(c *Client) error {
	params := eb
	result := struct {
		KeyMR              *Bytes32 `json:"chainhead"`
		ChainInProcessList bool     `json:"chaininprocesslist"`
	}{}
	if err := c.FactomdRequest("chain-head", params, &result); err != nil {
		return err
	}
	var zero Bytes32
	if *result.KeyMR == zero {
		if result.ChainInProcessList {
			return jrpc.Error{Message: "new chain in process list"}
		} else {
			return jrpc.Error{Code: -32009, Message: "Missing Chain Head"}
		}
	}
	eb.KeyMR = result.KeyMR
	return nil
}

// IsFirst returns true if this is the first EBlock in its chain, indicated by
// the PrevKeyMR being all zeroes. IsFirst returns false if eb is not populated
// or if the PrevKeyMR is not all zeroes.
func (eb EBlock) IsFirst() bool {
	return eb.IsPopulated() && *eb.PrevKeyMR == zeroBytes32
}

// Prev returns the an EBlock with its KeyMR initialized to eb.PrevKeyMR and
// ChainID initialized to eb.ChainID. If eb is the first Entry Block in the
// chain, then eb is returned.
func (eb EBlock) Prev() EBlock {
	if eb.IsFirst() {
		return eb
	}
	return EBlock{ChainID: eb.ChainID, KeyMR: eb.PrevKeyMR}
}

// GetAllPrev returns a slice of all preceding EBlocks in eb's chain, in order
// from earliest to latest, up to and including eb. So the last element of the
// returned slice is always equal to eb. If eb is the first entry block in its
// chain, then it is the only element in the slice.
//
// If you are only interested in obtaining the first entry block in eb's chain,
// and not all of the intermediary ones, then use GetFirst to reduce network
// calls and memory usage.
func (eb EBlock) GetAllPrev(c *Client) ([]EBlock, error) {
	ebs := []EBlock{eb}
	for ; !ebs[0].IsFirst(); ebs = append([]EBlock{ebs[0].Prev()}, ebs...) {
		if err := ebs[0].Get(c); err != nil {
			return nil, err
		}
	}
	return ebs, nil
}

// GetFirst finds the first Entry Block in eb's chain, and populates eb as
// such.
//
// GetFirst differs from GetAllPrev in that it does not allocate any additional
// EBlocks. GetFirst avoids allocating any new EBlocks by reusing eb to
// traverse up to the first entry block.
func (eb *EBlock) GetFirst(c *Client) error {
	for ; !eb.IsFirst(); *eb = eb.Prev() {
		if err := eb.Get(c); err != nil {
			return err
		}
	}
	return nil
}

const (
	EBlockHeaderLen = 32 + // [ChainID (Bytes32)] +
		32 + // [BodyMR (Bytes32)] +
		32 + // [PrevKeyMR (Bytes32)] +
		32 + // [PrevFullHash (Bytes32)] +
		4 + // [EB Sequence (uint32 BE)] +
		4 + // [DB Height (uint32 BE)] +
		4 // [Entry Count (uint32 BE)]

	EBlockObjectLen = 32 // Entry hash or minute marker

	EBlockMinBodyLen  = EBlockObjectLen * 2 // one entry hash & one minute marker
	EBlockMinTotalLen = EBlockHeaderLen + EBlockMinBodyLen

	EBlockMaxBodyLen  = math.MaxUint32 * EBlockObjectLen
	EBlockMaxTotalLen = EBlockHeaderLen + EBlockMaxBodyLen
)

// UnmarshalBinary unmarshals raw entry block data.
//
// Header
//
// [ChainID (Bytes32)] +
// [BodyMR (Bytes32)] +
// [PrevKeyMR (Bytes32)] +
// [PrevFullHash (Bytes32)] +
// [EB Sequence (uint32 BE)] +
// [DB Height (uint32 BE)] +
// [Object Count (uint32 BE)]
//
// Body
//
// [Object 0 (Bytes32)] // entry hash or minute marker +
// ... +
// [Object N (Bytes32)]
//
// https://github.com/FactomProject/FactomDocs/blob/master/factomDataStructureDetails.md#entry-block
func (eb *EBlock) UnmarshalBinary(data []byte) error {
	if len(data) < EBlockMinTotalLen {
		return fmt.Errorf("insufficient length")
	}
	if len(data) > EBlockMaxTotalLen {
		return fmt.Errorf("invalid length")
	}

	// When the eb.ChainID is already populated, just reuse the data.
	if eb.ChainID == nil {
		eb.ChainID = new(Bytes32)
		copy(eb.ChainID[:], data[:len(eb.ChainID)])
	}
	i := len(eb.ChainID)
	eb.BodyMR = new(Bytes32)
	i += copy(eb.BodyMR[:], data[i:i+len(eb.BodyMR)])
	eb.PrevKeyMR = new(Bytes32)
	i += copy(eb.PrevKeyMR[:], data[i:i+len(eb.PrevKeyMR)])
	eb.PrevFullHash = new(Bytes32)
	i += copy(eb.PrevFullHash[:], data[i:i+len(eb.PrevFullHash)])
	eb.Sequence = binary.BigEndian.Uint32(data[i : i+4])
	i += 4
	eb.Height = binary.BigEndian.Uint32(data[i : i+4])
	i += 4
	eb.ObjectCount = binary.BigEndian.Uint32(data[i : i+4])
	i += 4
	if len(data[i:]) != int(eb.ObjectCount*32) {
		return fmt.Errorf("invalid length")
	}

	// Parse all objects into Bytes32
	objects := make([]Bytes32, eb.ObjectCount)
	maxMinute := Bytes32{31: 10}
	var numMins int
	for oi := range objects {
		obj := &objects[len(objects)-1-oi] // Reverse object order
		i += copy(obj[:], data[i:i+len(obj)])
		if bytes.Compare(obj[:], maxMinute[:]) <= 0 { // if obj <= maxMinute
			numMins++ // obj is a minute marker
		}
	}
	// The last object (which is now index 0) must be a minute marker.
	if bytes.Compare(objects[0][:], maxMinute[:]) > 0 { // if obj > maxMinute
		return fmt.Errorf("invalid minute marker")
	}

	// Populate Entries from objects.
	eb.Entries = make([]Entry, int(eb.ObjectCount)-numMins)
	ei := len(eb.Entries) - 1
	var ts time.Time
	for _, obj := range objects {
		if bytes.Compare(obj[:], maxMinute[:]) <= 0 {
			ts = eb.Timestamp.
				Add(time.Duration(obj[31]) * time.Minute)
			continue
		}
		e := &eb.Entries[ei]
		e.Timestamp = ts
		e.ChainID = eb.ChainID
		e.Height = eb.Height
		obj := obj
		e.Hash = &obj
		ei--
	}
	return nil
}

func (eb *EBlock) MarshalBinaryHeader() ([]byte, error) {
	data := make([]byte, eb.MarshalBinaryLen())
	i := copy(data, eb.ChainID[:])
	i += copy(data[i:], eb.BodyMR[:])
	i += copy(data[i:], eb.PrevKeyMR[:])
	i += copy(data[i:], eb.PrevFullHash[:])
	binary.BigEndian.PutUint32(data[i:], eb.Sequence)
	i += 4
	binary.BigEndian.PutUint32(data[i:], eb.Height)
	i += 4
	eb.ObjectCount = eb.CountObjects()
	binary.BigEndian.PutUint32(data[i:], eb.ObjectCount)
	i += 4
	return data, nil
}

func (eb *EBlock) MarshalBinary() ([]byte, error) {
	data, err := eb.MarshalBinaryHeader()
	if err != nil {
		return nil, err
	}
	objects, err := eb.Objects()
	if err != nil {
		return nil, err
	}
	i := EBlockHeaderLen
	for _, obj := range objects {
		i += copy(data[i:], obj[:])
	}
	return data, nil
}

func (eb *EBlock) Objects() ([]Bytes32, error) {
	if eb.ObjectCount == 0 {
		eb.ObjectCount = eb.CountObjects()
	}
	objects := make([]Bytes32, eb.ObjectCount)
	var lastMin, oi int
	lastMin = int(eb.Entries[0].Timestamp.Sub(eb.Timestamp).Minutes())
	for _, e := range eb.Entries {
		min := int(e.Timestamp.Sub(eb.Timestamp).Minutes())
		if min > 10 {
			return nil, fmt.Errorf("invalid entry timestamp")
		}
		if min > lastMin {
			objects[oi][31] = byte(lastMin)
			oi++
			lastMin = min
		}
		objects[oi] = *e.Hash
		oi++
	}
	// Insert last minute marker
	lastE := eb.Entries[len(eb.Entries)-1]
	lastMin = int(lastE.Timestamp.Sub(eb.Timestamp).Minutes())
	objects[oi][31] = byte(lastMin)
	return objects, nil
}

func (eb *EBlock) CountObjects() uint32 {
	if len(eb.Entries) == 0 {
		panic("no entries")
	}
	var lastMin int
	numMins := 1 // There is always at least one minute marker.
	for _, e := range eb.Entries {
		min := int(e.Timestamp.Sub(eb.Timestamp).Minutes())
		if min > lastMin {
			numMins++
			lastMin = min
		}
	}
	return uint32(len(eb.Entries) + numMins)
}

func (eb *EBlock) MarshalBinaryLen() int {
	if eb.ObjectCount == 0 {
		eb.ObjectCount = eb.CountObjects()
	}
	return EBlockHeaderLen + int(eb.ObjectCount)*len(Bytes32{})
}

func (eb *EBlock) ComputeBodyMR() (Bytes32, error) {
	var bodyMR Bytes32
	objects, err := eb.Objects()
	if err != nil {
		return Bytes32{}, err
	}
	blocks := make([][]byte, len(objects))
	for i := range objects {
		blocks[i] = objects[i][:]
	}
	tree := merkle.NewTreeWithOpts(merkle.TreeOptions{
		DoubleOddNodes:    true,
		DisableHashLeaves: true})
	if err := tree.Generate(blocks, sha256.New()); err != nil {
		return Bytes32{}, err
	}
	root := tree.Root()
	copy(bodyMR[:], root.Hash)
	return bodyMR, nil
}

func (eb *EBlock) ComputeFullHash() (Bytes32, error) {
	data, err := eb.MarshalBinary()
	if err != nil {
		return Bytes32{}, err
	}
	return sha256.Sum256(data), nil
}

func (eb *EBlock) ComputeHeaderHash() (Bytes32, error) {
	header, err := eb.MarshalBinaryHeader()
	if err != nil {
		return Bytes32{}, err
	}
	return sha256.Sum256(header[:EBlockHeaderLen]), nil
}

func (eb *EBlock) ComputeKeyMR() (Bytes32, error) {
	headerHash, err := eb.ComputeHeaderHash()
	if err != nil {
		return Bytes32{}, err
	}
	bodyMR, err := eb.ComputeBodyMR()
	if err != nil {
		return Bytes32{}, err
	}
	data := make([]byte, len(headerHash)+len(bodyMR))
	i := copy(data, headerHash[:])
	copy(data[i:], bodyMR[:])
	return sha256.Sum256(data), nil
}
