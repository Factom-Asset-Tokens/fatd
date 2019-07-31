package factom

var (
	mainnetID = [...]byte{0xFA, 0x92, 0xE5, 0xA2}
	testnetID = [...]byte{0xFA, 0x92, 0xE5, 0xA3}
)

func Mainnet() NetworkID { return mainnetID }
func Testnet() NetworkID { return testnetID }

type NetworkID [4]byte

func (n NetworkID) String() string {
	switch n {
	case mainnetID:
		return "mainnet"
	case testnetID:
		return "testnet"
	default:
		return "custom: 0x" + Bytes(n[:]).String()
	}
}

func (n NetworkID) IsMainnet() bool {
	return n == mainnetID
}

func (n NetworkID) IsTestnet() bool {
	return n == testnetID
}

func (n NetworkID) IsCustom() bool {
	return !n.IsMainnet() && !n.IsTestnet()
}
