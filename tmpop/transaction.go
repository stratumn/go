// Copyright 2017 Stratumn SAS. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package tmpop

import (
	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/types"
)

// TxType represents the type of a Transaction (Write Segment/Value, Delete Segment/Value)
type TxType byte

const (
	// SaveSegment characterizes a transaction that saves a new segment
	SaveSegment TxType = iota

	// DeleteSegment characterizes a transaction that deletes a segment
	DeleteSegment

	// SaveValue characterizes a transaction that saves a new value
	SaveValue

	// DeleteValue characterizes a transaction that deletes a value
	DeleteValue
)

// Tx represents a TMPoP transaction
type Tx struct {
	TxType   TxType         `json:"type"`
	Segment  *cs.Segment    `json:"segment"`
	LinkHash *types.Bytes32 `json:"linkhash"`
	Key      []byte         `json:"key"`
	Value    []byte         `json:"value"`
}
