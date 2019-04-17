package state

import (
	"time"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat/fat1"
	"github.com/jinzhu/gorm"
)

type Metadata struct {
	gorm.Model

	Height uint64

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
