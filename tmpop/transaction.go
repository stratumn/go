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
	"encoding/json"

	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/types"
)

// TxType represents the type of a Transaction
type TxType byte

const (
	// CreateLink characterizes a transaction that creates a new link
	CreateLink TxType = iota
)

// Tx represents a TMPoP transaction
type Tx struct {
	TxType   TxType         `json:"type"`
	Link     *cs.Link       `json:"link"`
	LinkHash *types.Bytes32 `json:"linkhash"`
}

func unmarshallTx(txBytes []byte) (*Tx, *ABCIError) {
	tx := &Tx{}

	if err := json.Unmarshal(txBytes, tx); err != nil {
		return nil, &ABCIError{CodeTypeValidation, err.Error()}
	}

	return tx, nil
}
