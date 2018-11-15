package fat0

import (
	"sync"

	"github.com/Factom-Asset-Tokens/fatd/factom"
)

type State struct {
	Signatures   SignatureMap
	Balances     BalanceMap
	Height       uint64
	AmountIssued uint64
	Issuance

	mu *sync.RWMutex
}

type SignatureMap map[uint64]map[[SignatureSize]byte]bool
type BalanceMap map[factom.Bytes32]uint64

func NewState(issuance Issuance) State {
	return State{
		Signatures: make(SignatureMap),
		Balances:   make(BalanceMap),
		Issuance:   issuance,
		mu:         new(sync.RWMutex),
	}
}

func (s *State) Apply(t *Transaction) bool {
	if !s.UniqueSignatures(t) {
		return false
	}
	if !s.SufficientBalances(t) {
		return false
	}
	defer s.mu.Unlock()
	s.mu.Lock()
	for rcdHash, amount := range t.Inputs {
		s.Balances[rcdHash] -= amount
	}
	for rcdHash, amount := range t.Outputs {
		s.Balances[rcdHash] += amount
	}
	sig := new([SignatureSize]byte)
	copy(sig[:], t.ExtIDs[1])
	s.Signatures[t.Height][*sig] = true
	return true
}

func (s *State) SufficientBalances(t *Transaction) bool {
	for rcdHash, amount := range t.Inputs {
		balance := s.Balances[rcdHash]
		if amount > balance {
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

func (s *State) Balance(a *factom.Address) uint64 {
	defer s.mu.RUnlock()
	s.mu.RLock()
	return s.Balances[a.RCDHash()]
}
