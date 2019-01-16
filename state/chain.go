package state

import (
	"fmt"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat"
	"github.com/jinzhu/gorm"
)

type Chain struct {
	ID *factom.Bytes32
	ChainStatus
	fat.Identity
	fat.Issuance
	Metadata
	*gorm.DB
}

func (chain Chain) String() string {
	return fmt.Sprintf("{ChainStatus:%v, ID:%v, Metadata:%+v, "+
		"fat.Identity:%+v, fat.Issuance:%+v}",
		chain.ChainStatus, chain.ID, chain.Metadata,
		chain.Identity, chain.Issuance)
}

func (chain *Chain) ignore() {
	chain.ID = nil
	chain.ChainStatus = ChainStatusIgnored
}
func (chain *Chain) track(first factom.Entry) error {
	chain.ChainStatus = ChainStatusTracked
	chain.Identity.ChainID = factom.NewBytes32(first.ExtIDs[3])
	chain.Metadata.Token = string(first.ExtIDs[1])
	chain.Metadata.Issuer = chain.Identity.ChainID
	chain.Metadata.Height = first.Height

	if err := chain.setupDB(); err != nil {
		return err
	}
	log.Debugf("Tracked: %v", chain)
	return nil
}
func (chain *Chain) issue(issuance fat.Issuance) error {
	chain.ChainStatus = ChainStatusIssued
	chain.Issuance = issuance

	if err := chain.saveIssuance(); err != nil {
		return err
	}
	log.Debugf("Issued: %v", chain)
	return nil
}
