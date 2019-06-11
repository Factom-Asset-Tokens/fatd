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

package state

import (
	"time"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat/fat1"
	"github.com/jinzhu/gorm"
)

type Metadata struct {
	gorm.Model

	Height uint32

	Token  string
	Issuer *factom.Bytes32

	Issued uint64
}

type entry struct {
	ID        uint64
	Hash      *factom.Bytes32 `gorm:"type:VARCHAR(32); UNIQUE_INDEX; NOT NULL;"`
	Timestamp time.Time       `gorm:"NOT NULL;"`
	Data      factom.Bytes    `gorm:"NOT NULL;"`
}

func newEntry(e factom.Entry) entry {
	b, _ := e.MarshalBinary()
	return entry{
		Hash:      e.Hash,
		Timestamp: e.Timestamp.Time(),
		Data:      b,
	}
}

func (e entry) IsValid() bool {
	return *e.Hash == factom.EntryHash(e.Data)
}

func (e entry) Entry() factom.Entry {
	fe := factom.Entry{Hash: e.Hash, Timestamp: (*factom.Time)(&e.Timestamp)}
	fe.UnmarshalBinary(e.Data)
	return fe
}

type Address struct {
	gorm.Model
	RCDHash *factom.FAAddress `gorm:"type:varchar(32); UNIQUE_INDEX; NOT NULL;"`

	Balance uint64 `gorm:"NOT NULL;"`

	To   []entry `gorm:"many2many:address_transactions_to;"`
	From []entry `gorm:"many2many:address_transactions_from;"`
}

func newAddress(fa factom.FAAddress) Address {
	return Address{RCDHash: &fa}
}

func (a Address) Address() factom.FAAddress {
	return *a.RCDHash
}

type NFToken struct {
	gorm.Model
	NFTokenID fat1.NFTokenID `gorm:"UNIQUE_INDEX"`
	Metadata  []byte
	OwnerID   uint    `gorm:"INDEX"`
	Owner     Address `gorm:"foreignkey:OwnerID"`

	PreviousOwners []Address `gorm:"many2many:nf_token_previousowners;"`
	Transactions   []entry   `gorm:"many2many:nf_token_transactions;"`
}
