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
	"context"
	"testing"

	"github.com/stratumn/go-chainscript/chainscripttest"
	"github.com/stratumn/go-core/testutil"
	"github.com/stratumn/go-core/validation/validators"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParticipantsValidator(t *testing.T) {
	v := validators.NewParticipantsValidator()

	t.Run("Validate()", func(t *testing.T) {
		t.Run("rejects unknown step", func(t *testing.T) {
			l := chainscripttest.NewLinkBuilder(t).
				WithProcess(validators.GovernanceProcess).
				WithMapID(validators.ParticipantsMap).
				WithStep("pwn").
				Build()
			err := v.Validate(context.Background(), nil, l)
			testutil.AssertWrappedErrorEqual(t, err, validators.ErrInvalidParticipantStep)
		})
	})

	t.Run("ShouldValidate()", func(t *testing.T) {
		t.Run("ignores non-governance process", func(t *testing.T) {
			l := chainscripttest.NewLinkBuilder(t).
				WithProcess("not-governance").
				WithMapID(validators.ParticipantsMap).
				Build()
			assert.False(t, v.ShouldValidate(l))
		})

		t.Run("ignores non-participants map", func(t *testing.T) {
			l := chainscripttest.NewLinkBuilder(t).
				WithProcess(validators.GovernanceProcess).
				WithMapID("not-participants").
				Build()
			assert.False(t, v.ShouldValidate(l))
		})

		t.Run("validates governance participants", func(t *testing.T) {
			l := chainscripttest.NewLinkBuilder(t).
				WithProcess(validators.GovernanceProcess).
				WithMapID(validators.ParticipantsMap).
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
