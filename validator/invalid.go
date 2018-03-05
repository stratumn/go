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
	"crypto/sha256"

	"github.com/pkg/errors"
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/types"
)

// invalidValidator never validates a link.
type invalidValidator struct {
	Config  *validatorBaseConfig
	Message error
}

func newInvalidValidator(baseConfig *validatorBaseConfig, msg string) Validator {
	return &invalidValidator{
		Config:  baseConfig,
		Message: errors.New(msg),
	}
}

func (iv invalidValidator) Hash() (*types.Bytes32, error) {
	hash := types.Bytes32(sha256.Sum256([]byte(iv.Message.Error())))
	return &hash, nil
}

func (iv invalidValidator) ShouldValidate(link *cs.Link) bool {
	return iv.Config.ShouldValidate(link)
}

// Validate always return false.
func (iv invalidValidator) Validate(store store.SegmentReader, link *cs.Link) error {
	return iv.Message
}
