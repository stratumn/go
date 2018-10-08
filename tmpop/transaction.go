// Copyright 2016-2018 Stratumn SAS. All rights reserved.
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

	"github.com/stratumn/go-chainscript"
)

// TxType represents the type of a Transaction
type TxType byte

const (
	// CreateLink characterizes a transaction that creates a new link
	CreateLink TxType = iota
)

// Tx represents a TMPoP transaction
type Tx struct {
	TxType   TxType               `json:"type"`
	Link     *chainscript.Link    `json:"link"`
	LinkHash chainscript.LinkHash `json:"linkhash"`
}

func unmarshallTx(txBytes []byte) (*Tx, *ABCIError) {
	tx := &Tx{}

	if err := json.Unmarshal(txBytes, tx); err != nil {
		return nil, &ABCIError{
			Code: CodeTypeValidation,
			Log:  err.Error(),
		}
	}

	return tx, nil
}
