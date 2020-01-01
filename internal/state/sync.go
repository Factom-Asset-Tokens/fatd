// MIT License
//
// Copyright 2018 Canonical Ledgers, LLC
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
// IN THE SOFTWARE.

package state

import (
	"context"
	"fmt"

	"github.com/Factom-Asset-Tokens/factom"
)

func SyncEBlocks(ctx context.Context, c *factom.Client, chain Chain,
	eblocks []factom.EBlock) error {
	if err := chain.UpdateSidechainData(ctx, c); err != nil {
		return fmt.Errorf("state.Chain.UpdateSidechainData(): %w", err)
	}
	for i := range eblocks {
		eb := eblocks[len(eblocks)-1-i] // Earliest EBlock first.

		// Get DBlock Timestamp and KeyMR
		var dblock factom.DBlock
		dblock.Height = eb.Height
		if err := dblock.Get(ctx, c); err != nil {
			return fmt.Errorf("factom.DBlock.Get(): %w", err)
		}

		eb.SetTimestamp(dblock.Timestamp)

		if err := eb.GetEntries(ctx, c); err != nil {
			return fmt.Errorf("factom.EBlock.GetEntries(): %w", err)
		}

		if err := Apply(chain, dblock.KeyMR, eb); err != nil {
			return fmt.Errorf("state.Apply(): %w", err)
		}
	}

	chain.ToFactomChain().Log.Infof("Chain synced.")
	return nil
}
