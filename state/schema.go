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
	return entry{
		Hash:      e.Hash,
		Timestamp: e.Timestamp.Time,
		Data:      e.MarshalBinary(),
	}
}

func (e entry) IsValid() bool {
	return *e.Hash == factom.EntryHash(e.Data)
}

func (e entry) Entry() factom.Entry {
	fe := factom.Entry{Hash: e.Hash, Timestamp: &factom.Time{Time: e.Timestamp}}
	fe.UnmarshalBinary(e.Data)
	return fe
}

type address struct {
	gorm.Model
	RCDHash *factom.RCDHash `gorm:"type:varchar(32); UNIQUE_INDEX; NOT NULL;"`

	Balance uint64 `gorm:"NOT NULL;"`

	To   []entry `gorm:"many2many:address_transactions_to;"`
	From []entry `gorm:"many2many:address_transactions_from;"`
}

func newAddress(fa factom.Address) address {
	return address{RCDHash: fa.RCDHash()}
}

func (a address) Address() factom.Address {
	return factom.NewAddress(a.RCDHash)
}

type nftoken struct {
	gorm.Model
	NFTokenID fat1.NFTokenID `gorm:"UNIQUE_INDEX;"`
	OwnerID   uint
	Owner     address `gorm:"foreignkey:OwnerID"`

	PreviousOwners []address `gorm:"many2many:nftoken_previousowners;"`
	Transactions   []entry   `gorm:"many2many:nftoken_transactions;"`
}
