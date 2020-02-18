package state

func (chain *FactomChain) Copy() Chain {
	chainCopy := *chain
	return &chainCopy
}

func (chain *FATChain) Copy() Chain {
	chainCopy := *chain
	return &chainCopy
}
