package fat0

import (
	"crypto/sha256"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"golang.org/x/crypto/ed25519"
)

type Transaction struct {
	Inputs  []AddressAmount `json:"inputs"`
	Outputs []AddressAmount `json:"outputs"`
	Height  uint64          `json:"blockheight"`
	Salt    string          `json:"salt,omitempty"`
	Entry

	Coinbase bool `json:"-"`
}

type AddressAmount struct {
	factom.Address `json:"address"`
	Amount         uint64 `json:"amount"`
}

func (t *Transaction) Valid(idKey *factom.Bytes32) bool {
	if t.Unmarshal() != nil {
		return false
	}
	if !t.ValidData() {
		return false
	}
	if !t.ValidExtIDs() {
		return false
	}
	if t.Coinbase {
		if t.RCDHash() != *idKey {
			return false
		}
	} else {
		if !t.VerifyRCDHashes() {
			return false
		}
	}
	if !t.VerifySignatures() {
		return false
	}
	return true
}

func (t *Transaction) Unmarshal() error {
	return t.Entry.Unmarshal(t)
}

var (
	coinbase factom.Address
)

const (
	MaxHeightDifference = uint64(3)
)

func (t *Transaction) ValidData() bool {
	if len(t.Inputs) > 0 && len(t.Outputs) > 0 &&
		t.Height <= t.Entry.Height &&
		t.Entry.Height-t.Height < MaxHeightDifference &&
		t.SumInputs() == t.SumOutputs() {
		inputs := map[factom.Bytes32]bool{}
		for i, a := range t.Inputs {
			if a.Amount == 0 {
				return false
			}
			if a.RCDHash() == coinbase.RCDHash() {
				// A coinbase transaction may only have a
				// single input. No other inputs may be the
				// coinbase address.
				if i == 0 && len(t.Inputs) == 1 {
					t.Coinbase = true
					return true
				}
				return false
			}
			// Enforce input uniqueness
			if inputs[a.RCDHash()] {
				return false
			}
			inputs[a.RCDHash()] = true

		}
		outputs := map[factom.Bytes32]bool{}
		for _, a := range t.Outputs {
			if a.Amount == 0 {
				return false
			}
			// Enforce output uniqueness
			if outputs[a.RCDHash()] {
				return false
			}
			outputs[a.RCDHash()] = true

		}
		return true
	}
	return false
}

func (t *Transaction) SumInputs() uint64 {
	return sum(t.Inputs)
}

func (t *Transaction) SumOutputs() uint64 {
	return sum(t.Inputs)
}

func sum(as []AddressAmount) uint64 {
	var sum uint64
	for i, _ := range as {
		sum += as[i].Amount
	}
	return sum
}

func (t *Transaction) ValidExtIDs() bool {
	if len(t.ExtIDs) >= 2*len(t.Inputs) {
		for i, _ := range t.Inputs {
			if len(t.ExtIDs[i*2]) != RCDSize ||
				len(t.ExtIDs[i*2+1]) != SignatureSize {
				return false
			}
		}
		return true
	}
	return false
}

func (t *Transaction) VerifySignatures() bool {
	msg := append(t.ChainID[:], t.Content...)
	for i, _ := range t.Inputs {
		pubKey := ed25519.PublicKey(t.ExtIDs[i*2][1:])
		if !ed25519.Verify(pubKey, msg, t.ExtIDs[i*2+1]) {
			return false
		}
	}
	return true
}

func (t *Transaction) VerifyRCDHashes() bool {
	for i, _ := range t.Inputs {
		input := &t.Inputs[i]
		if input.RCDHash() != sha256d(t.ExtIDs[i*2]) {
			return false
		}
	}
	return true
}

func (t *Transaction) RCDHash() [sha256.Size]byte {
	return sha256d(t.ExtIDs[0])
}
