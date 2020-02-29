package main

import (
	"context"
	"fmt"

	"github.com/Factom-Asset-Tokens/factom"
	flag "github.com/spf13/pflag"
)

// TODO: refactor issue and transact commands to use EntryCreator

// EntryCreator contains the flags required to create entries. Creating entries
// requires a funded EsAddress. This can be supplied directly with the
// --ecadr/-e flag. Alternatively, an ECAddress may be provided, and
// PopulateEsAddress will query the factom-walletd service for the EsAddress.
// See ECEsAddress for more info.
//
// If --force is given, then CheckECBalance will skip the check.
//
// If --curl is given, then the curl commands required to submit the entries
// will be printed to stdout. Otherwise, they will be submitted directly to
// factomd.
type EntryCreator struct {
	ECEsAddress      // --ecadr
	Curl        bool // --curl
	Force       bool // --force
}

// Flags returns a new FlagSet with the EntryCreator flags for ec.
func (ec *EntryCreator) Flags() *flag.FlagSet {
	flags := flag.NewFlagSet("Entry Creation", flag.ContinueOnError)
	flags.VarPF(&ec.ECEsAddress, "ecadr", "e",
		"EC or Es address to pay for entries").DefValue = ""
	flags.BoolVar(&ec.Force, "force", false,
		"Skip sanity checks for balances, chain status, and sk1 key")
	flags.BoolVar(&ec.Curl, "curl", false,
		"Do not submit Factom entry; print curl commands")
	return flags
}

func (ec *EntryCreator) ValidateFlagStructure(flags *flag.FlagSet) error {
	return required(flags, "ecadr")
}

func (ec *EntryCreator) CheckECBalance(cost uint) error {
	if ec.Force {
		vrbLog.Println("Skipping EC balance check.")
		return nil
	}
	vrbLog.Printf("Checking EC balance... ")
	ecBalance, err := ec.EC.GetBalance(context.Background(), FactomClient)
	if err != nil {
		return err
	}
	if uint64(cost) > ecBalance {
		return fmt.Errorf("Insufficient EC balance %v: needs at least %v",
			ecBalance, cost)
	}
	return nil
}

func (ec *EntryCreator) ComposeCreate(entry *factom.Entry) (factom.Bytes32, error) {
	vrbLog.Println("Composing entry...")
	commit, reveal, txID, err := entry.Compose(ec.Es)
	if err != nil {
		return txID, err
	}
	return txID, ec.Create(commit, reveal, entry.Hash)
}
func (ec *EntryCreator) Create(commit, reveal []byte, hash *factom.Bytes32) error {
	if ec.Curl {
		ec.PrintCurl(commit, reveal)
		return nil
	}
	return ec.Submit(commit, reveal, hash)
}
func (ec *EntryCreator) Submit(commit, reveal []byte, hash *factom.Bytes32) error {
	vrbLog.Printf("Submitting entry %v...", hash)
	vrbLog.Printf("Committing...")
	if err := FactomClient.Commit(context.Background(), commit); err != nil {
		return err
	}
	vrbLog.Println("Revealing...")
	return FactomClient.Reveal(context.Background(), reveal)
}
func (ec *EntryCreator) PrintCurl(commit, reveal []byte) {
	commitMethod := "commit"
	revealMethod := "reveal"
	switch len(commit) {
	case factom.EntryCommitSize:
		commitMethod += "-entry"
		revealMethod += "-entry"
	case factom.ChainCommitSize:
		commitMethod += "-chain"
		revealMethod += "-chain"
	}

	vrbLog.Println("Curl commands:")
	fmt.Printf(`curl -X POST --data-binary '{"jsonrpc":"2.0","id":0,"method":%q,"params":{"message":%q}}' -H 'content-type:text/plain;' %v`+"\n",
		commitMethod, factom.Bytes(commit), FactomClient.FactomdServer)

	fmt.Printf(`curl -X POST --data-binary '{"jsonrpc":"2.0","id": 0,"method":%q,"params":{"entry":%q}}' -H 'content-type:text/plain;' %v`+"\n",
		revealMethod, factom.Bytes(reveal), FactomClient.FactomdServer)
}
