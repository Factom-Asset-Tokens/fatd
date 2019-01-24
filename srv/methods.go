package srv

import (
	"bytes"
	"encoding/json"
	"fmt"

	jrpc "github.com/AdamSLevy/jsonrpc2/v10"
	"github.com/jinzhu/gorm"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat"
	"github.com/Factom-Asset-Tokens/fatd/fat/fat0"
	"github.com/Factom-Asset-Tokens/fatd/fat/fat1"
	"github.com/Factom-Asset-Tokens/fatd/flag"
	"github.com/Factom-Asset-Tokens/fatd/state"
)

var jrpcMethods = jrpc.MethodMap{
	"get-issuance":           getIssuance(false),
	"get-issuance-entry":     getIssuance(true),
	"get-transaction":        getTransaction(false),
	"get-transaction-entry":  getTransaction(true),
	"get-transactions":       getTransactions(false),
	"get-transactions-entry": getTransactions(true),
	"get-balance":            getBalance,
	"get-stats":              getStats,
	"get-nf-token":           getNFToken,

	"send-transaction": sendTransaction,

	"get-daemon-tokens":     getDaemonTokens,
	"get-daemon-properties": getDaemonProperties,
}

type ResultGetIssuance struct {
	ParamsToken
	Hash      *factom.Bytes32 `json:"entryhash"`
	Timestamp *factom.Time    `json:"timestamp"`
	Issuance  fat.Issuance    `json:"issuance"`
}

func getIssuance(entry bool) jrpc.MethodFunc {
	return func(data json.RawMessage) interface{} {
		params := ParamsToken{}
		chain, err := validate(data, &params)
		if err != nil {
			return err
		}

		if entry {
			return chain.Issuance.Entry.Entry
		}
		return ResultGetIssuance{
			ParamsToken: ParamsToken{
				ChainID:       chain.ID,
				TokenID:       chain.Token,
				IssuerChainID: chain.Identity.ChainID,
			},
			Hash:      chain.Hash,
			Timestamp: chain.Issuance.Timestamp,
			Issuance:  chain.Issuance,
		}
	}
}

type ResultGetTransaction struct {
	Hash      *factom.Bytes32  `json:"entryhash"`
	Timestamp *factom.Time     `json:"timestamp"`
	Tx        fat0.Transaction `json:"data"`
}

func getTransaction(entry bool) jrpc.MethodFunc {
	return func(data json.RawMessage) interface{} {
		params := ParamsGetTransaction{}
		chain, err := validate(data, &params)
		if err != nil {
			return err
		}

		transaction, err := chain.GetTransaction(params.Hash)
		if err != nil {
			panic(err)
		}
		if !transaction.IsPopulated() {
			return ErrorTransactionNotFound
		}

		if entry {
			return transaction.Entry.Entry
		}
		if err := transaction.UnmarshalEntry(); err != nil {
			panic(err)
		}
		return ResultGetTransaction{
			Hash:      transaction.Hash,
			Timestamp: transaction.Timestamp,
			Tx:        transaction,
		}
	}
}

func getTransactions(entry bool) jrpc.MethodFunc {
	return func(data json.RawMessage) interface{} {
		params := ParamsGetTransactions{}
		chain, err := validate(data, &params)
		if err != nil {
			return err
		}

		// Lookup Txs
		transactions, err := chain.GetTransactions(params.Hash,
			params.FactoidAddress, params.ToFrom,
			*params.Start, *params.Limit)
		if err != nil {
			log.Debug(err)
			panic(err)
		}
		if len(transactions) == 0 {
			return ErrorTransactionNotFound
		}
		if entry {
			txs := make([]factom.Entry, len(transactions))
			for i := range txs {
				txs[i] = transactions[i].Entry.Entry
				txs[i].ChainID = nil
			}
			return txs
		}

		txs := make([]ResultGetTransaction, len(transactions))
		for i := range txs {
			txs[i].Hash = transactions[i].Hash
			txs[i].Timestamp = transactions[i].Timestamp
			txs[i].Tx = transactions[i]
		}

		return txs
	}
}

func getBalance(data json.RawMessage) interface{} {
	params := ParamsGetBalance{}
	chain, err := validate(data, &params)
	if err != nil {
		return err
	}

	// Lookup Txs
	balance, err := chain.GetAddress(params.Address)
	if err != nil {
		panic(err)
	}
	return balance
}

type ResultGetStats struct {
	Supply                   int64        `json:"supply"`
	CirculatingSupply        uint64       `json:"circulating"`
	Burned                   uint64       `json:"burned"`
	Transactions             int          `json:"transactions"`
	IssuanceTimestamp        *factom.Time `json:"issuancets"`
	LastTransactionTimestamp *factom.Time `json:"lasttxts,omitempty"`
}

var coinbaseRCDHash = func() *factom.RCDHash {
	coinbase := factom.Address{}
	return coinbase.RCDHash()
}()

func getStats(data json.RawMessage) interface{} {
	params := ParamsToken{}
	chain, err := validate(data, &params)
	if err != nil {
		return err
	}

	coinbase, err := chain.GetAddress(coinbaseRCDHash)
	if err != nil {
		panic(err)
	}
	burned := coinbase.Balance
	txs, err := chain.GetTransactions(nil, nil, "", 0, 0)
	if err != nil {
		panic(err)
	}

	var lastTxTs *factom.Time
	if len(txs) > 0 {
		lastTxTs = txs[len(txs)-1].Timestamp
	}
	return ResultGetStats{
		Supply:                   chain.Supply,
		CirculatingSupply:        chain.Issued - burned,
		Burned:                   burned,
		Transactions:             len(txs),
		IssuanceTimestamp:        chain.Issuance.Timestamp,
		LastTransactionTimestamp: lastTxTs,
	}
}

type ResultGetNFToken struct {
	NFTokenID fat1.NFTokenID
	Owner     *factom.RCDHash
	Metadata  json.RawMessage
}

func getNFToken(data json.RawMessage) interface{} {
	params := ParamsGetNFToken{}
	chain, err := validate(data, &params)
	if err != nil {
		return err
	}

	if chain.Type != fat1.Type {
		err := ErrorTokenNotFound
		err.Data = "Token Chain is not FAT-1"
		return err
	}

	tkn := state.NFToken{NFTokenID: *params.NFTokenID}
	chain.GetNFToken(&tkn)

	return ResultGetNFToken{
		NFTokenID: tkn.NFTokenID,
		Metadata:  tkn.Metadata,
		Owner:     tkn.Owner.RCDHash,
	}
}

func sendTransaction(data json.RawMessage) interface{} {
	if len(flag.ECPub) == 0 {
		return ErrorNoEC
	}
	params := ParamsSendTransaction{}
	chain, err := validate(data, &params)
	if err != nil {
		return err
	}

	entry := params.Entry()
	hash := entry.ComputeHash()
	transaction, err := chain.GetTransaction(&hash)
	if transaction.IsPopulated() {
		err := ErrorInvalidTransaction
		err.Data = "duplicate transaction"
		return err
	}
	if err != gorm.ErrRecordNotFound {
		log.Error(err)
		panic(err)
	}

	switch chain.Type {
	case fat0.Type:
		if err := validFAT0Transaction(chain, entry); err != nil {
			return err
		}
	case fat1.Type:
		if err := validFAT1Transaction(chain, entry); err != nil {
			return err
		}
	default:
		panic("invalid FAT type")
	}

	txID, err := entry.Create(flag.ECPub)
	if err != nil {
		log.Error(err)
		panic(err)
	}

	return struct {
		ChainID *factom.Bytes32 `json:"chainid"`
		TxID    *factom.Bytes32 `json:"txid"`
		Hash    *factom.Bytes32 `json:"entryhash"`
	}{ChainID: chain.ID, TxID: txID, Hash: entry.Hash}
}

func validFAT0Transaction(chain *state.Chain, entry factom.Entry) error {
	tx := fat0.NewTransaction(entry)
	rpcErr := ErrorInvalidTransaction
	if err := tx.Valid(chain.IDKey); err != nil {
		rpcErr.Data = err.Error()
		return rpcErr
	}

	// check balances
	if tx.IsCoinbase() {
		if tx.Inputs.Sum() > uint64(chain.Supply)-chain.Issued {
			rpcErr.Data = "insufficient coinbase supply"
			return rpcErr
		}
		return nil
	}
	for rcdHash, amount := range tx.Inputs {
		adr, err := chain.GetAddress(&rcdHash)
		if err != nil {
			log.Error(err)
			panic(err)
		}
		if amount > adr.Balance {
			rpcErr.Data = fmt.Sprintf("insufficient balance: %v", rcdHash)
			return rpcErr
		}
	}
	return nil
}

func validFAT1Transaction(chain *state.Chain, entry factom.Entry) error {
	tx := fat1.NewTransaction(entry)
	rpcErr := ErrorInvalidTransaction
	if err := tx.Valid(chain.IDKey); err != nil {
		rpcErr.Data = err.Error()
		return rpcErr
	}

	for rcdHash, tkns := range tx.Inputs {
		adr, err := chain.GetAddress(&rcdHash)
		if err != nil {
			log.Error(err)
			panic(err)
		}
		if tx.IsCoinbase() {
			if chain.Supply > 0 &&
				uint64(chain.Supply)-chain.Issued < uint64(len(tkns)) {
				// insufficient coinbase supply
				rpcErr.Data = "insufficient coinbase supply"
				return rpcErr
			}
			for tknID := range tkns {
				tkn := state.NFToken{NFTokenID: tknID}
				err := chain.GetNFToken(&tkn)
				if err == nil {
					rpcErr.Data = fmt.Sprintf(
						"NFTokenID(%v) already exists", tknID)
					return rpcErr
				}
				if err != gorm.ErrRecordNotFound {
					log.Error(err)
					panic(err)
				}
			}
			break
		}
		if adr.Balance < uint64(len(tkns)) {
			rpcErr.Data = fmt.Sprintf("insufficient balance: %v", rcdHash)
			return rpcErr
		}
		for tknID := range tkns {
			tkn := state.NFToken{NFTokenID: tknID, OwnerID: adr.ID}
			err := chain.GetNFToken(&tkn)
			if err == gorm.ErrRecordNotFound {
				rpcErr.Data = fmt.Sprintf(
					"NFTokenID(%v) is not owned by %v",
					tknID, rcdHash)
				return rpcErr
			}
			if err != nil {
				log.Error(err)
				panic(err)
			}
		}
	}

	return nil
}

func getDaemonTokens(data json.RawMessage) interface{} {
	if data != nil {
		return ParamsErrorNoParams
	}

	issuedIDs := state.Chains.GetIssued()
	chains := make([]ParamsToken, len(issuedIDs))
	for i, chainID := range issuedIDs {
		chain := state.Chains.Get(chainID)
		chains[i].ChainID = chainID
		chains[i].TokenID = chain.Token
		chains[i].IssuerChainID = chain.Issuer
	}
	return chains
}

func getDaemonProperties(data json.RawMessage) interface{} {
	if data != nil {
		return ParamsErrorNoParams
	}
	return struct {
		FatdVersion string `json:"fatdversion"`
		APIVersion  string `json:"apiversion"`
	}{FatdVersion: "0.0.0", APIVersion: "v0"}
}

func validate(data json.RawMessage, params Params) (*state.Chain, error) {
	if data == nil {
		return nil, params.Error()
	}
	if err := unmarshalStrict(data, params); err != nil {
		return nil, jrpc.NewInvalidParamsError(err.Error())
	}
	chainID := params.ValidChainID()
	if chainID == nil || !params.IsValid() {
		return nil, params.Error()
	}
	chain := state.Chains.Get(chainID)
	if !chain.IsIssued() {
		return nil, ErrorTokenNotFound
	}
	return &chain, nil
}

func unmarshalStrict(data []byte, v interface{}) error {
	b := bytes.NewBuffer(data)
	d := json.NewDecoder(b)
	d.DisallowUnknownFields()
	return d.Decode(v)
}
