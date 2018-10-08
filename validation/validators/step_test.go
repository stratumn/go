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

	"github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-chainscript/chainscripttest"
	"github.com/stratumn/go-core/validation/validators"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessStepValidator(t *testing.T) {
	t.Run("New()", func(t *testing.T) {
		testCases := []struct {
			name    string
			process string
			step    string
			err     error
		}{{
			"missing process",
			"",
			"init",
			validators.ErrMissingProcess,
		}, {
			"missing step",
			"test_process",
			"",
			validators.ErrMissingLinkStep,
		}, {
			"valid process and step",
			"test_process",
			"init",
			nil,
		}}

		for _, tt := range testCases {
			t.Run(tt.name, func(t *testing.T) {
				v, err := validators.NewProcessStepValidator(tt.process, tt.step)
				if tt.err != nil {
					assert.Nil(t, v)
					assert.EqualError(t, err, tt.err.Error())
				} else {
					assert.NoError(t, err)
					assert.NotNil(t, v)
				}
			})
		}
	})

	t.Run("ShouldValidate()", func(t *testing.T) {
		process := "test_process"
		step := "init"

		v, err := validators.NewProcessStepValidator(process, step)
		require.NoError(t, err)

		testCases := []struct {
			name  string
			link  *chainscript.Link
			match bool
		}{{
			"nil link",
			nil,
			false,
		}, {
			"process mismatch",
			chainscripttest.NewLinkBuilder(t).WithProcess("some_process").WithStep(step).Build(),
			false,
		}, {
			"step mismatch",
			chainscripttest.NewLinkBuilder(t).WithProcess(process).WithStep("some_step").Build(),
			false,
		}, {
			"match",
			chainscripttest.NewLinkBuilder(t).WithProcess(process).WithStep(step).Build(),
			true,
		}}

		for _, tt := range testCases {
			t.Run(tt.name, func(t *testing.T) {
				match := v.ShouldValidate(tt.link)
				assert.Equal(t, tt.match, match)

				err := v.Validate(context.Background(), nil, tt.link)
				if tt.match {
					assert.NoError(t, err)
				} else {
					assert.EqualError(t, err, validators.ErrInvalidProcessOrStep.Error())
				}
			})
		}
	})

	t.Run("Hash()", func(t *testing.T) {
		v1, _ := validators.NewProcessStepValidator("p1", "s1")
		h1, err := v1.Hash()
		require.NoError(t, err)

		v2, _ := validators.NewProcessStepValidator("p1", "s2")
		h2, err := v2.Hash()
		require.NoError(t, err)

		v3, _ := validators.NewProcessStepValidator("p2", "s1")
		h3, err := v3.Hash()
		require.NoError(t, err)

		assert.NotEqual(t, h1, h2)
		assert.NotEqual(t, h1, h3)
		assert.NotEqual(t, h2, h3)
	})
}
