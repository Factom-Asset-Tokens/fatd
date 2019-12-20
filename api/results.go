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

package api

import (
	"encoding/json"

	"github.com/Factom-Asset-Tokens/factom"
	"github.com/Factom-Asset-Tokens/factom/fat"
	"github.com/Factom-Asset-Tokens/factom/fat1"
)

const APIVersion = "1"

type ResultGetIssuance struct {
	ParamsToken
	Hash      *factom.Bytes32 `json:"entryhash"`
	Timestamp int64           `json:"timestamp"`
	Issuance  fat.Issuance    `json:"issuance"`
}

type ResultGetTransaction struct {
	Hash      *factom.Bytes32 `json:"entryhash"`
	Timestamp int64           `json:"timestamp"`
	Tx        interface{}     `json:"data"`
	Pending   bool            `json:"pending,omitempty"`
}

type ResultGetBalances map[factom.Bytes32]uint64

func (r ResultGetBalances) MarshalJSON() ([]byte, error) {
	strMap := make(map[string]uint64, len(r))
	for chainID, balance := range r {
		strMap[chainID.String()] = balance
	}
	return json.Marshal(strMap)
}

func (r *ResultGetBalances) UnmarshalJSON(data []byte) error {
	var strMap map[string]uint64
	if err := json.Unmarshal(data, &strMap); err != nil {
		return err
	}
	*r = make(map[factom.Bytes32]uint64, len(strMap))
	var chainID factom.Bytes32
	for str, balance := range strMap {
		if err := chainID.Set(str); err != nil {
			return err
		}
		(*r)[chainID] = balance
	}
	return nil
}

type ResultGetStats struct {
	ParamsToken
	Issuance                 *fat.Issuance
	IssuanceHash             *factom.Bytes32
	CirculatingSupply        uint64 `json:"circulating"`
	Burned                   uint64 `json:"burned"`
	Transactions             int64  `json:"transactions"`
	IssuanceTimestamp        int64  `json:"issuancets"`
	LastTransactionTimestamp int64  `json:"lasttxts,omitempty"`
	NonZeroBalances          int64  `json:"nonzerobalances, omitempty"`
}

type ResultGetNFToken struct {
	NFTokenID  fat1.NFTokenID    `json:"id"`
	Owner      *factom.FAAddress `json:"owner,omitempty"`
	Burned     bool              `json:"burned,omitempty"`
	Metadata   json.RawMessage   `json:"metadata,omitempty"`
	CreationTx *factom.Bytes32   `json:"creationtx"`
}

type ResultGetDaemonProperties struct {
	FatdVersion string           `json:"fatdversion"`
	APIVersion  string           `json:"apiversion"`
	NetworkID   factom.NetworkID `json:"factomnetworkid"`
}

type ResultGetSyncStatus struct {
	Sync    uint32 `json:"syncheight"`
	Current uint32 `json:"factomheight"`
}
