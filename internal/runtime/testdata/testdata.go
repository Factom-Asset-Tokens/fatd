package testdata

// #include "./src/runtime_test.h"
import "C"
import (
	"time"

	"github.com/Factom-Asset-Tokens/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat0"
	"github.com/Factom-Asset-Tokens/fatd/internal/runtime"
)

// Context returns a runtime.Context populated with the test data expected by
// the api_test.wasm.
func Context() runtime.Context {
	var tx fat0.Transaction

	sender := factom.FAAddress(genBytes32(C.GET_SENDER_ERR))
	tx.Inputs = fat0.AddressAmountMap{sender: C.GET_AMOUNT_EXP}
	tx.Outputs = tx.Inputs

	hash := genBytes32(C.GET_ENTRY_HASH_ERR)
	tx.Entry.Hash = &hash
	tx.Entry.Timestamp = time.Unix(C.GET_TIMESTAMP_EXP, 0)

	return runtime.Context{
		DBlock:      factom.DBlock{Height: uint32(C.GET_HEIGHT_EXP)},
		Transaction: tx,
	}
}

var ErrMap = map[int32]string{
	int32(C.SUCCESS):            "success",
	int32(C.GET_HEIGHT_ERR):     "error: ext_get_height",
	int32(C.GET_SENDER_ERR):     "error: ext_get_sender",
	int32(C.GET_AMOUNT_ERR):     "error: ext_get_amount",
	int32(C.GET_ENTRY_HASH_ERR): "error: ext_get_entry_hash",
	int32(C.GET_TIMESTAMP_ERR):  "error: ext_get_timestamp",
}

func genBytes32(val byte) factom.Bytes32 {
	var hash factom.Bytes32
	for i := range hash[:] {
		hash[i] = byte(i) + val
	}
	return hash
}
