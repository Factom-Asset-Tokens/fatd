package engine

import (
	"context"

	"github.com/Factom-Asset-Tokens/factom"
	"github.com/Factom-Asset-Tokens/fatd/internal/state"
)

type State interface {
	ApplyEBlock(context.Context, *factom.Bytes32, factom.EBlock) error
	ApplyPendingEntries(context.Context, []factom.Entry) error
	SetSync(context.Context, uint32, *factom.Bytes32) error
	GetSync() uint32
	TrackedIDs() []*factom.Bytes32
	IssuedIDs() []*factom.Bytes32
	Get(context.Context, *factom.Bytes32, bool) (state.Chain, func(), error)
	Close()
}

var openState = func(ctx context.Context, c *factom.Client,
	dbPath string,
	networkID factom.NetworkID,
	whitelist, blacklist []factom.Bytes32,
	skipDBValidation, repair bool,
) (_ State, _ context.Context, err error) {

	return state.Open(ctx, c,
		dbPath, networkID,
		whitelist, blacklist,
		skipDBValidation, repair)
}

var _state State

func Get(ctx context.Context, chainID *factom.Bytes32, includePending bool) (
	state.Chain, func(), error) {
	return _state.Get(ctx, chainID, includePending)
}

func TrackedIDs() []*factom.Bytes32 {
	return _state.TrackedIDs()
}

func IssuedIDs() []*factom.Bytes32 {
	return _state.IssuedIDs()
}
