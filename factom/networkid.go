package factom

import (
	"fmt"
	"strings"
)

var (
	mainnetID  = [...]byte{0xFA, 0x92, 0xE5, 0xA2}
	testnetID  = [...]byte{0x88, 0x3e, 0x09, 0x3b}
	localnetID = [...]byte{0xFA, 0x92, 0xE5, 0xA4}
)

func MainnetID() NetworkID  { return mainnetID }
func TestnetID() NetworkID  { return testnetID }
func LocalnetID() NetworkID { return localnetID }

type NetworkID [4]byte

func (n NetworkID) String() string {
	switch n {
	case mainnetID:
		return "mainnet"
	case testnetID:
		return "testnet"
	case localnetID:
		return "localnet"
	default:
		return "custom: 0x" + Bytes(n[:]).String()
	}
}
func (n *NetworkID) Set(netIDStr string) error {
	switch strings.ToLower(netIDStr) {
	case "main", "mainnet":
		*n = mainnetID
	case "test", "testnet":
		*n = testnetID
	case "local", "localnet":
		*n = localnetID
	default:
		if netIDStr[:2] == "0x" {
			// omit leading 0x
			netIDStr = netIDStr[2:]
		}
		var b Bytes
		if err := b.Set(netIDStr); err != nil {
			return err
		}
		if len(b) != len(n[:]) {
			return fmt.Errorf("invalid length")
		}
		copy(n[:], b)
	}
	return nil
}

func (n NetworkID) IsMainnet() bool {
	return n == mainnetID
}

func (n NetworkID) IsTestnet() bool {
	return n == testnetID
}

func (n NetworkID) IsLocalnet() bool {
	return n == localnetID
}

func (n NetworkID) IsCustom() bool {
	return !n.IsMainnet() && !n.IsTestnet()
}
