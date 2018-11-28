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

func TestGovernanceRulesValidator(t *testing.T) {
	v := validators.NewGovernanceRulesValidator()

	t.Run("ShouldValidate()", func(t *testing.T) {
		t.Run("returns false for participant governance", func(t *testing.T) {
			l := chainscripttest.NewLinkBuilder(t).
				WithProcess(validators.GovernanceProcess).
				WithMapID(validators.ParticipantsMap).
				Build()
			assert.False(t, v.ShouldValidate(l))
		})

		t.Run("returns false for non-governance links", func(t *testing.T) {
			l := chainscripttest.NewLinkBuilder(t).
				WithProcess("test_process").
				Build()
			assert.False(t, v.ShouldValidate(l))
		})

		t.Run("returns true for governance link", func(t *testing.T) {
			l := chainscripttest.NewLinkBuilder(t).
				WithProcess(validators.GovernanceProcess).
				WithMapID("test_process").
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
