package fat0

import (
	"bytes"
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"strconv"
	"time"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/FactomProject/ed25519"
)

// Entry has variables and methods common to all fat0 entries.
type Entry struct {
	Metadata json.RawMessage `json:"metadata,omitempty"`

	factom.Entry `json:"-"`
}

// unmarshalEntry unmarshals the content of the factom.Entry into the provided
// variable v, disallowing all unknown fields.
func (e Entry) unmarshalEntry(v interface {
	ExpectedJSONLength() int
}) error {
	contentJSONLen := compactJSONLen(e.Content)
	if contentJSONLen == 0 {
		return fmt.Errorf("not a single valid JSON")
	}
	d := json.NewDecoder(bytes.NewReader(e.Content))
	d.DisallowUnknownFields()
	if err := d.Decode(v); err != nil {
		return err
	}
	expectedJSONLen := v.ExpectedJSONLength()
	if contentJSONLen != expectedJSONLen {
		return fmt.Errorf("contentJSONLen (%v) != expectedJSONLen (%v)",
			contentJSONLen, expectedJSONLen)
	}
	return nil
}

func (e Entry) metadataLen() int {
	if e.Metadata == nil {
		return 0
	}
	l := len(`,`)
	l += len(`"metadata":`) + compactJSONLen(e.Metadata)
	return l
}

func compactJSONLen(data []byte) int {
	buf := bytes.NewBuffer(make([]byte, 0, len(data)))
	json.Compact(buf, data)
	cmp, _ := ioutil.ReadAll(buf)
	return len(cmp)
}

func (e *Entry) marshalEntry(v interface {
	ValidData() error
}) error {
	if err := v.ValidData(); err != nil {
		return err
	}
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	e.Content = factom.Bytes(data)
	return nil
}

// ValidExtIDs validates the structure of the ExtIDs of the factom.Entry to
// make sure that it has a valid timestamp salt and a valid set of
// RCD/signature pairs.
func (e Entry) ValidExtIDs() error {
	if len(e.ExtIDs) < 3 || len(e.ExtIDs)%2 != 1 {
		return fmt.Errorf("invalid number of ExtIDs")
	}
	if err := e.validTimestamp(); err != nil {
		return err
	}
	extIDs := e.ExtIDs[1:]
	for i := 0; i < len(extIDs)/2; i++ {
		rcd := extIDs[i*2]
		if len(rcd) != factom.RCDSize {
			return fmt.Errorf("ExtIDs[%v]: invalid RCD size", i+1)
		}
		if rcd[0] != factom.RCDType {
			return fmt.Errorf("ExtIDs[%v]: invalid RCD type", i+1)
		}
		sig := extIDs[i*2+1]
		if len(sig) != factom.SignatureSize {
			return fmt.Errorf("ExtIDs[%v]: invalid signature size", i+1)
		}
	}
	return e.validSignatures()
}

func (e Entry) validTimestamp() error {
	sec, err := strconv.ParseInt(string(e.ExtIDs[0]), 10, 64)
	if err != nil {
		return fmt.Errorf("timestamp salt: %v", err)
	}
	ts := time.Unix(sec, 0)
	diff := e.Timestamp.Sub(ts)
	if -12*time.Hour > diff || diff > 12*time.Hour {
		return fmt.Errorf("timestamp salt expired")
	}
	return nil
}

// validSignatures returns true if the first num RCD/signature pairs in the
// ExtIDs are valid.
func (e Entry) validSignatures() error {
	num := len(e.ExtIDs) / 2
	timeSalt := e.ExtIDs[0]
	salt := append(timeSalt, e.ChainID[:]...)
	msg := append(salt, e.Content...)
	pubKey := new([ed25519.PublicKeySize]byte)
	sig := new([ed25519.SignatureSize]byte)
	extIDs := e.ExtIDs[1:]
	for sigID := 0; sigID < num; sigID++ {
		copy(pubKey[:], extIDs[sigID*2][1:])
		copy(sig[:], extIDs[sigID*2+1])
		extIDSalt := []byte(strconv.FormatInt(int64(sigID), 10))
		msg := append(extIDSalt, msg...)
		msgHash := sha512.Sum512(msg)
		if !ed25519.VerifyCanonical(pubKey, msgHash[:], sig) {
			return fmt.Errorf("ExtIDs[%v]: invalid signature", sigID*2+2)
		}
	}
	return nil
}

// Sign the ExtIDIndex Salt + Timestamp Salt + Chain ID Salt + Content of the
// factom.Entry and add the RCD + signature pairs for the given addresses to
// the ExtIDs. This clears any existing ExtIDs.
func (e *Entry) Sign(as ...factom.Address) {
	e.Timestamp = &factom.Time{Time: time.Now()}
	ts := time.Now().Add(time.Duration(
		-rand.Int63n(int64(12 * time.Hour))))
	timeSalt := []byte(strconv.FormatInt(ts.Unix(), 10))
	salt := append(timeSalt, e.ChainID[:]...)
	msg := append(salt, e.Content...)
	e.ExtIDs = make([]factom.Bytes, 1, len(as)*2+1)
	e.ExtIDs[0] = timeSalt
	for sigID, a := range as {
		extIDSalt := []byte(strconv.FormatInt(int64(sigID), 10))
		msg := append(extIDSalt, msg...)
		msgHash := sha512.Sum512(msg)
		sig := ed25519.Sign(a.PrivateKey, msgHash[:])
		e.ExtIDs = append(e.ExtIDs, a.RCD(), sig[:])
	}
}
