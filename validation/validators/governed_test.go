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

package validators_test

import (
	"testing"

	"github.com/stratumn/go-chainscript/chainscripttest"
	"github.com/stratumn/go-core/validation/validators"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGovernedProcessValidator(t *testing.T) {
	v := validators.NewGovernedProcessValidator()

	t.Run("ShouldValidate()", func(t *testing.T) {
		t.Run("returns false for governance segments", func(t *testing.T) {
			l := chainscripttest.NewLinkBuilder(t).
				WithProcess(validators.GovernanceProcess).
				Build()
			assert.False(t, v.ShouldValidate(l))
		})

		t.Run("returns true for normal segment", func(t *testing.T) {
			l := chainscripttest.NewLinkBuilder(t).
				WithProcess("test_process").
				WithMapID("test_map").
				Build()
			assert.True(t, v.ShouldValidate(l))
		})
	})

	t.Run("Hash()", func(t *testing.T) {
		h, err := v.Hash()
		require.NoError(t, err)
		assert.Nil(t, h)
	})
}
