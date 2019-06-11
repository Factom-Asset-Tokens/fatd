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
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"time"

	merkle "github.com/AdamSLevy/go-merkle"
)

var (
	mainnetID = [...]byte{0xFA, 0x92, 0xE5, 0xA2}
	testnetID = [...]byte{0xFA, 0x92, 0xE5, 0xA3}

	adminBlockChainID       = Bytes32{31: 0x0a}
	entryCreditBlockChainID = Bytes32{31: 0x0c}
	factoidBlockChainID     = Bytes32{31: 0x0f}
)

// DBlock represents a Factom Directory Block.
type DBlock struct {
	KeyMR *Bytes32 `json:"keymr"`

	FullHash *Bytes32 `json:"dbhash"`

	Header DBlockHeader `json:"header"`

	Height uint32 `json:"dbheight"`

	// DBlock.Get populates EBlocks with their ChainID and KeyMR.
	EBlocks []EBlock `json:"dbentries,omitempty"`
}

type DBlockHeader struct {
	NetworkID    [4]byte  `json:"networkid"`
	BodyMR       *Bytes32 `json:"bodymr"`
	PrevKeyMR    *Bytes32 `json:"prevkeymr"`
	PrevFullHash *Bytes32 `json:"prevfullhash"`

	Timestamp time.Time `json:"timestamp"`
}

func (dbh *DBlockHeader) UnmarshalJSON(data []byte) error {
	type dBlockHeader DBlockHeader
	d := struct {
		*dBlockHeader
		NetworkID uint32 `json:"networkid"`
		Timestamp int64  `json:"timestamp"`
	}{dBlockHeader: (*dBlockHeader)(dbh)}
	if err := json.Unmarshal(data, &d); err != nil {
		return err
	}
	dbh.Timestamp = time.Unix(d.Timestamp*60, 0)
	binary.BigEndian.PutUint32(dbh.NetworkID[:], d.NetworkID)
	return nil
}

// IsPopulated returns true if db has already been successfully populated by a
// call to Get. IsPopulated returns false if db.EBlocks is nil.
func (db DBlock) IsPopulated() bool {
	return db.EBlocks != nil &&
		db.Header.BodyMR != nil &&
		db.Header.PrevKeyMR != nil &&
		db.Header.PrevFullHash != nil
}

// Get queries factomd for the Directory Block at db.Height. After a successful
// call, the EBlocks will all have their ChainID and KeyMR, but not their
// Entries. Call Get on the EBlocks individually to populate their Entries.
func (db *DBlock) Get(c *Client) error {
	if db.IsPopulated() {
		return nil
	}

	if db.KeyMR != nil {
		params := struct {
			Hash *Bytes32 `json:"hash"`
		}{Hash: db.KeyMR}
		var result struct {
			Data Bytes `json:"data"`
		}
		if err := c.FactomdRequest("raw-data", params, &result); err != nil {
			return err
		}
		return db.UnmarshalBinary(result.Data)
	}

	params := struct {
		Height uint32 `json:"height"`
	}{db.Height}
	result := struct {
		DBlock  *DBlock
		RawData Bytes `json:"rawdata"`
	}{DBlock: db}
	if err := c.FactomdRequest("dblock-by-height", params, &result); err != nil {
		return err
	}

	return nil
}

const (
	DBlockHeaderLen = 1 + // [Version byte (0x00)]
		4 + // NetworkID
		32 + // BodyMR
		32 + // PrevKeyMR
		32 + // PrevFullHash
		4 + // Timestamp
		4 + // DB Height
		4 // EBlock Count

	DBlockEBlockLen = 32 + // ChainID
		32 // KeyMR

	DBlockMinBodyLen = DBlockEBlockLen + // Admin Block
		DBlockEBlockLen + // EC Block
		DBlockEBlockLen // FCT Block
	DBlockMinTotalLen = DBlockHeaderLen + DBlockMinBodyLen

	DBlockMaxBodyLen  = math.MaxUint32 * DBlockEBlockLen
	DBlockMaxTotalLen = DBlockHeaderLen + DBlockMaxBodyLen
)

// UnmarshalBinary unmarshals raw directory block data.
//
// Header
// [Version byte (0x00)] +
// [NetworkID (4 bytes)] +
// [BodyMR (Bytes32)] +
// [PrevKeyMR (Bytes32)] +
// [PrevFullHash (Bytes32)] +
// [Timestamp (4 bytes)] +
// [DB Height (4 bytes)] +
// [EBlock Count (4 bytes)]
//
// Body
// [Admin Block ChainID (Bytes32{31:0x0a})] +
// [Admin Block LookupHash (Bytes32)] +
// [EC Block ChainID (Bytes32{31:0x0c})] +
// [EC Block HeaderHash (Bytes32)] +
// [FCT Block ChainID (Bytes32{31:0x0f})] +
// [FCT Block KeyMR (Bytes32)] +
// [ChainID 0 (Bytes32)] +
// [KeyMR 0 (Bytes32)] +
// ... +
// [ChainID N (Bytes32)] +
// [KeyMR N (Bytes32)] +
//
// https://github.com/FactomProject/FactomDocs/blob/master/factomDataStructureDetails.md#directory-block
func (db *DBlock) UnmarshalBinary(data []byte) error {
	if len(data) < DBlockMinTotalLen {
		return fmt.Errorf("insufficient length")
	}
	if len(data) > DBlockMaxTotalLen {
		return fmt.Errorf("invalid length")
	}
	if data[0] != 0x00 {
		return fmt.Errorf("invalid version byte")
	}
	i := 1
	i += copy(db.Header.NetworkID[:], data[i:i+len(db.Header.NetworkID)])
	db.Header.BodyMR = new(Bytes32)
	i += copy(db.Header.BodyMR[:], data[i:i+len(db.Header.BodyMR)])
	db.Header.PrevKeyMR = new(Bytes32)
	i += copy(db.Header.PrevKeyMR[:], data[i:i+len(db.Header.PrevKeyMR)])
	db.Header.PrevFullHash = new(Bytes32)
	i += copy(db.Header.PrevFullHash[:], data[i:i+len(db.Header.PrevFullHash)])
	db.Header.Timestamp = time.Unix(int64(binary.BigEndian.Uint32(data[i:i+4]))*60, 0)
	i += 4
	db.Height = binary.BigEndian.Uint32(data[i : i+4])
	i += 4
	ebsLen := int(binary.BigEndian.Uint32(data[i : i+4]))
	i += 4
	if len(data[i:]) < ebsLen*DBlockEBlockLen {
		return fmt.Errorf("insufficient length")
	}
	db.EBlocks = make([]EBlock, ebsLen)
	for ebi := range db.EBlocks {
		eb := &db.EBlocks[ebi]
		eb.ChainID = new(Bytes32)
		i += copy(eb.ChainID[:], data[i:i+len(eb.ChainID)])
		eb.KeyMR = new(Bytes32)
		i += copy(eb.KeyMR[:], data[i:i+len(eb.KeyMR)])
	}
	return nil
}

func (db *DBlock) MarshalBinary() ([]byte, error) {
	data, err := db.MarshalBinaryHeader()
	if err != nil {
		return nil, err
	}
	i := DBlockHeaderLen
	for _, eb := range db.EBlocks {
		i += copy(data[i:], eb.ChainID[:])
		i += copy(data[i:], eb.KeyMR[:])
	}
	return data, nil
}

func (db *DBlock) MarshalBinaryHeader() ([]byte, error) {
	totalLen := db.MarshalBinaryLen()
	if totalLen > DBlockMaxTotalLen {
		return nil, fmt.Errorf("too many EBlocks")
	}
	data := make([]byte, totalLen)
	i := 1 // Skip version byte
	i += copy(data[i:], db.Header.NetworkID[:])
	i += copy(data[i:], db.Header.BodyMR[:])
	i += copy(data[i:], db.Header.PrevKeyMR[:])
	i += copy(data[i:], db.Header.PrevFullHash[:])
	binary.BigEndian.PutUint32(data[i:], uint32(db.Header.Timestamp.Unix()/60))
	i += 4
	binary.BigEndian.PutUint32(data[i:], db.Height)
	i += 4
	binary.BigEndian.PutUint32(data[i:], uint32(len(db.EBlocks)))
	i += 4
	return data, nil
}

func (db *DBlock) MarshalBinaryLen() int {
	return DBlockHeaderLen + len(db.EBlocks)*DBlockEBlockLen
}

func (db DBlock) ComputeBodyMR() (Bytes32, error) {
	var bodyMR Bytes32
	blocks := make([][]byte, len(db.EBlocks))
	for i, eb := range db.EBlocks {
		blocks[i] = make([]byte, len(eb.ChainID)+len(eb.KeyMR))
		j := copy(blocks[i], eb.ChainID[:])
		copy(blocks[i][j:], eb.KeyMR[:])
	}
	tree := merkle.NewTreeWithOpts(merkle.TreeOptions{DoubleOddNodes: true})
	if err := tree.Generate(blocks, sha256.New()); err != nil {
		return bodyMR, err
	}
	root := tree.Root()
	copy(bodyMR[:], root.Hash)
	return bodyMR, nil
}

func (db DBlock) ComputeFullHash() (Bytes32, error) {
	data, err := db.MarshalBinary()
	if err != nil {
		return Bytes32{}, err
	}
	return sha256.Sum256(data), nil
}

func (db DBlock) ComputeHeaderHash() (Bytes32, error) {
	header, err := db.MarshalBinaryHeader()
	if err != nil {
		return Bytes32{}, err
	}
	return sha256.Sum256(header[:DBlockHeaderLen]), nil
}

func (db DBlock) ComputeKeyMR() (Bytes32, error) {
	headerHash, err := db.ComputeHeaderHash()
	if err != nil {
		return Bytes32{}, err
	}
	bodyMR, err := db.ComputeBodyMR()
	if err != nil {
		return Bytes32{}, err
	}
	data := make([]byte, len(headerHash)+len(bodyMR))
	i := copy(data, headerHash[:])
	copy(data[i:], bodyMR[:])
	return sha256.Sum256(data), nil
}
