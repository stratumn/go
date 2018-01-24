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

package validator

import (
	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/types"
)

const (
	// DefaultFilename is the default filename for the file with the rules of validation
	DefaultFilename = "/data/validation/rules.json"
)

// validator defines the interface with single Validate() method
type validator interface {
	// Validate runs validations on a link and returns an error
	// if the link is invalid.
	Validate(store.SegmentReader, *cs.Link) error
}

// Validator defines a validator that has an internal state, identified by
// its hash.
type Validator interface {
	validator

	// Hash returns the hash of the validator's state.
	// It can be used to know which set of validations were applied
	// to a block.
	Hash() (*types.Bytes32, error)
}
