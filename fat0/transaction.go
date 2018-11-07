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
	return &Transaction{Entry: Entry{Entry: *entry}}
}

func (t *Transaction) Coinbase() bool {
	_, ok := t.Inputs[coinbase.RCDHash()]
	return ok
}

func (t *Transaction) Valid(idKey *factom.Bytes32) error {
	if err := t.Unmarshal(); err != nil {
		return err
	}
	if err := t.ValidData(); err != nil {
		return err
	}
	if err := t.ValidExtIDs(); err != nil {
		return err
	}
	if t.Coinbase() {
		if t.RCDHash() != *idKey {
			return fmt.Errorf("invalid RCD")
		}
	} else {
		if !t.ValidRCDs() {
			return fmt.Errorf("invalid RCDs")
		}
	}
	if !t.ValidSignatures() {
		return fmt.Errorf("invalid signatures")
	}
	return nil
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
		t.Entry.Height-t.Height > MaxHeightDifference {
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
	// Select the shortest map to range through.
	a := t.Inputs
	b := t.Outputs
	if len(t.Outputs) < len(t.Inputs) {
		a = t.Outputs
		b = t.Inputs
	}
	// Ensure that no address exists in both the Inputs and Outputs.
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
	return sum(t.Outputs)
}

func sum(aam AddressAmountMap) uint64 {
	var sum uint64
	for _, amount := range aam {
		sum += amount
	}
	return sum
}

func (t *Transaction) ValidExtIDs() error {
	if len(t.ExtIDs) < 2*len(t.Inputs) {
		return fmt.Errorf("insufficient number of ExtIDs")
	}
	for i := 0; i < len(t.Inputs); i++ {
		rcd := t.ExtIDs[i*2]
		if len(rcd) != RCDSize {
			return fmt.Errorf("invalid RCD size")
		}
		if rcd[0] != RCDType {
			return fmt.Errorf("invalid RCD type")
		}
		sig := t.ExtIDs[i*2+1]
		if len(sig) != SignatureSize {
			return fmt.Errorf("invalid signature size")
		}
	}
	return nil
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

func (t *Transaction) ValidRCDs() bool {
	rcdHashes := make(AddressAmountMap)
	for i := 0; i < len(t.Inputs); i++ {
		rcdHashes[sha256d(t.ExtIDs[i*2])] = 0
	}
	for inputRCDHash, _ := range t.Inputs {
		if _, ok := rcdHashes[inputRCDHash]; !ok {
			return false
		}
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
