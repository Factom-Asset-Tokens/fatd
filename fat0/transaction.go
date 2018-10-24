package fat0

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/FactomProject/ed25519"
)

type Transaction struct {
	Inputs  AddressAmountMap `json:"inputs"`
	Outputs AddressAmountMap `json:"outputs"`
	Height  uint64           `json:"blockheight"`
	Salt    string           `json:"salt,omitempty"`
	Entry
}

func NewTransaction(entry *factom.Entry) *Transaction {
	return &Transaction{Entry: Entry{Entry: entry}}
}

func (t *Transaction) Coinbase() bool {
	_, ok := t.Inputs[coinbase.RCDHash()]
	return ok
}

func (t *Transaction) Valid(idKey *factom.Bytes32) bool {
	if t.Unmarshal() != nil {
		return false
	}
	if t.ValidData() != nil {
		return false
	}
	if !t.ValidExtIDs() {
		return false
	}
	if t.Coinbase() {
		if t.RCDHash() != *idKey {
			return false
		}
	} else {
		if !t.ValidRCDHashes() {
			return false
		}
	}
	if !t.ValidSignatures() {
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

func (t *Transaction) ValidData() error {
	if t.Height > t.Entry.Height ||
		t.Entry.Height-t.Height < MaxHeightDifference {
		return fmt.Errorf("invalid height")
	}
	if len(t.Inputs) == 0 {
		return fmt.Errorf("no inputs")
	}
	if len(t.Outputs) == 0 {
		return fmt.Errorf("no outputs")
	}
	if t.SumInputs() != t.SumOutputs() {
		return fmt.Errorf("sum(inputs) != sum(outputs)")
	}
	if t.Coinbase() && len(t.Inputs) != 1 {
		return fmt.Errorf("invalid coinbase transaction")
	}
	a := t.Inputs
	b := t.Outputs
	if len(t.Outputs) < len(t.Inputs) {
		a = t.Outputs
		b = t.Inputs
	}
	for rcdHash := range a {
		if _, ok := b[rcdHash]; ok {
			return fmt.Errorf("%v appears in both inputs and outputs",
				factom.NewAddress(&rcdHash))
		}
	}
	return nil
}

func (t *Transaction) SumInputs() uint64 {
	return sum(t.Inputs)
}

func (t *Transaction) SumOutputs() uint64 {
	return sum(t.Inputs)
}

func sum(aam AddressAmountMap) uint64 {
	var sum uint64
	for _, amount := range aam {
		sum += amount
	}
	return sum
}

func (t *Transaction) ValidExtIDs() bool {
	if len(t.ExtIDs) >= 2*len(t.Inputs) {
		for i := 0; i < len(t.Inputs); i++ {
			if len(t.ExtIDs[i*2]) != RCDSize ||
				len(t.ExtIDs[i*2+1]) != SignatureSize {
				return false
			}
		}
		return true
	}
	return false
}

func (t *Transaction) ValidSignatures() bool {
	msg := append(t.ChainID[:], t.Content...)
	pubKey := new([ed25519.PublicKeySize]byte)
	sig := new([ed25519.SignatureSize]byte)
	for i := 0; i < len(t.Inputs); i++ {
		copy(pubKey[:], t.ExtIDs[i*2][1:])
		copy(sig[:], t.ExtIDs[i*2+1])
		if !ed25519.VerifyCanonical(pubKey, msg, sig) {
			return false
		}
	}
	return true
}

func (t *Transaction) ValidRCDHashes() bool {
	i := 0
	for rcdHash, _ := range t.Inputs {
		if rcdHash != sha256d(t.ExtIDs[i*2]) {
			return false
		}
		i++
	}
	return true
}

func (t *Transaction) RCDHash() [sha256.Size]byte {
	return sha256d(t.ExtIDs[0])
}

type AddressAmountMap map[factom.Bytes32]uint64

type addressAmount struct {
	Address factom.Address `json:"address"`
	Amount  uint64         `json:"amount"`
}

func (aP *AddressAmountMap) UnmarshalJSON(data []byte) error {
	a := make(AddressAmountMap)
	var aaS []addressAmount
	if err := json.Unmarshal(data, &aaS); err != nil {
		return err
	}
	for _, aa := range aaS {
		if aa.Amount == 0 {
			return fmt.Errorf("invalid amount (0) for address: %v", aa)
		}
		if _, duplicate := a[aa.Address.RCDHash()]; duplicate {
			return fmt.Errorf("duplicate address: %v", aa)
		}
		a[aa.Address.RCDHash()] = aa.Amount
	}
	*aP = a
	return nil
}

func (a AddressAmountMap) MarshalJSON() ([]byte, error) {
	aaS := make([]addressAmount, len(a))
	i := 0
	for rcdHash, amount := range a {
		// Skip addresses with 0 amounts
		if amount == 0 {
			continue
		}
		aaS[i].Address = *factom.NewAddress(&rcdHash)
		aaS[i].Amount = amount
		i++
	}
	return json.Marshal(aaS)
}
