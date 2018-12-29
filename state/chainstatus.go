package state

type ChainStatus uint

const (
	ChainStatusUnknown ChainStatus = 0
	ChainStatusTracked ChainStatus = 1
	ChainStatusIssued  ChainStatus = 3
	ChainStatusIgnored ChainStatus = 4
)

func (status ChainStatus) IsUnknown() bool {
	return status == ChainStatusUnknown
}
func (status ChainStatus) IsIgnored() bool {
	return status == ChainStatusIgnored
}
func (status ChainStatus) IsTracked() bool {
	return status&ChainStatusTracked == ChainStatusTracked
}
func (status ChainStatus) IsIssued() bool {
	return status&ChainStatusIssued == ChainStatusIssued
}

func (status ChainStatus) String() string {
	s := "Unknown"
	switch status {
	case ChainStatusTracked:
		s = "Tracked"
	case ChainStatusIssued:
		s = "Issued"
	case ChainStatusIgnored:
		s = "Ignored"
	}
	return s
}
