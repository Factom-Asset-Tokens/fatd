package fat103

import (
	"crypto/sha512"
	"math/rand"
	"strconv"
	"time"

	"github.com/Factom-Asset-Tokens/factom"
	"github.com/Factom-Asset-Tokens/fatd/internal/jsonlen"
)

// Sign the RCD/Sig ID Salt + Timestamp Salt + Chain ID Salt + Content of the
// factom.Entry and add the RCD + signature pairs for the given addresses to
// the ExtIDs. This clears any existing ExtIDs.
func Sign(e factom.Entry, signingSet ...factom.RCDSigner) factom.Entry {
	// Set the Entry's timestamp so that the signatures will verify against
	// this time salt.
	timeSalt := newTimestampSalt()
	e.Timestamp = time.Now()

	// Compose the signed message data using exactly allocated bytes.
	maxRcdSigIDSaltStrLen := jsonlen.Uint64(uint64(len(signingSet)))
	maxMsgLen := maxRcdSigIDSaltStrLen +
		len(timeSalt) +
		len(e.ChainID) +
		len(e.Content)
	msg := make(factom.Bytes, maxMsgLen)
	i := maxRcdSigIDSaltStrLen
	i += copy(msg[i:], timeSalt[:])
	i += copy(msg[i:], e.ChainID[:])
	copy(msg[i:], e.Content)

	// Generate the ExtIDs for each address in the signing set.
	e.ExtIDs = make([]factom.Bytes, 1, len(signingSet)*2+1)
	e.ExtIDs[0] = timeSalt
	for rcdSigID, a := range signingSet {
		// Compose the RcdSigID salt and prepend it to the message.
		rcdSigIDSalt := strconv.FormatUint(uint64(rcdSigID), 10)
		start := maxRcdSigIDSaltStrLen - len(rcdSigIDSalt)
		copy(msg[start:], rcdSigIDSalt)

		msgHash := sha512.Sum512(msg[start:])

		e.ExtIDs = append(e.ExtIDs, a.RCD(), a.Sign(msgHash[:]))
	}
	return e
}
func newTimestampSalt() []byte {
	timestamp := time.Now().Add(time.Duration(-rand.Int63n(int64(1 * time.Hour))))
	return []byte(strconv.FormatInt(timestamp.Unix(), 10))
}
