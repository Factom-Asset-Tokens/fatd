package state

type ChainStatus uint

const (
	ChainStatusUnknown ChainStatus = 1 << iota
	ChainStatusTracked ChainStatus = 1 << iota
	ChainStatusIssued  ChainStatus = (1 << iota) | ChainStatusTracked
	ChainStatusIgnored ChainStatus = 1 << iota
)

func (status ChainStatus) IsUnknown() bool {
	return status == ChainStatusUnknown
}
func (status ChainStatus) IsIgnored() bool {
	return status == ChainStatusIgnored
}
func (status ChainStatus) IsTracked() bool {
	return status|ChainStatusTracked == ChainStatusTracked
}
func (status ChainStatus) IsIssued() bool {
	return status|ChainStatusIssued == ChainStatusIssued
}
