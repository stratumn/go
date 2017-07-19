// Copyright 2017 Stratumn SAS. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
