package fat0

import "github.com/Factom-Asset-Tokens/fatd/factom"

type State struct {
	Signatures   SignatureMap
	Balances     BalanceMap
	Height       uint64
	AmountIssued uint64
	*Issuance
}

type SignatureMap map[uint64]map[[SignatureSize]byte]bool
type BalanceMap map[factom.Bytes32]uint64

func NewState(issuance *Issuance) *State {
	return &State{
		Signatures: make(SignatureMap),
		Balances:   make(BalanceMap),
		Issuance:   issuance,
	}
}

func (s *State) Apply(t *Transaction) {
	for i, _ := range t.Inputs {
		input := &t.Inputs[i]
		s.Balances[input.RCDHash()] -= input.Amount
	}
	for i, _ := range t.Outputs {
		output := &t.Outputs[i]
		s.Balances[output.RCDHash()] += output.Amount
	}
	sig := new([SignatureSize]byte)
	copy(sig[:], t.ExtIDs[1])
	s.Signatures[t.Height][*sig] = true
}

func (s *State) SufficientBalances(t *Transaction) bool {
	for _, input := range t.Inputs {
		balance := s.Balances[input.RCDHash()]
		if input.Amount > balance {
			return false
		}
	}
	return true
}

func (s *State) UniqueSignatures(t *Transaction) bool {
	prevSigs := s.Signatures[t.Height]
	sig := new([SignatureSize]byte)
	copy(sig[:], t.ExtIDs[1])
	if _, replay := prevSigs[*sig]; replay {
		return false
	}
	return true
}
