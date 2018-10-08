// Copyright 2017-2018 Stratumn SAS. All rights reserved.
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

	"github.com/pkg/errors"
	"github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-chainscript/chainscripttest"
	"github.com/stratumn/go-core/validation/validators"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMultiValidator_New(t *testing.T) {
	t.Run("New()", func(t *testing.T) {
		mv := validators.NewMultiValidator(validators.Validators{})
		assert.NotNil(t, mv)
	})

	t.Run("ShouldValidate()", func(t *testing.T) {
		psv1, _ := validators.NewProcessStepValidator("p1", "s1")
		psv2, _ := validators.NewProcessStepValidator("p2", "s2")
		mv := validators.NewMultiValidator(validators.Validators{psv1, psv2})

		testCases := []struct {
			name  string
			link  *chainscript.Link
			match bool
		}{{
			"no match",
			chainscripttest.NewLinkBuilder(t).WithProcess("p3").Build(),
			false,
		}, {
			"match",
			chainscripttest.NewLinkBuilder(t).WithProcess("p2").WithStep("s2").Build(),
			true,
		}}

		for _, tt := range testCases {
			t.Run(tt.name, func(t *testing.T) {
				match := mv.ShouldValidate(tt.link)
				assert.Equal(t, tt.match, match)
			})
		}
	})

	t.Run("Hash()", func(t *testing.T) {
		psv1, _ := validators.NewProcessStepValidator("p1", "s1")
		psv2, _ := validators.NewProcessStepValidator("p2", "s2")

		mv1 := validators.NewMultiValidator(validators.Validators{psv1})
		mv2 := validators.NewMultiValidator(validators.Validators{psv2})
		mv3 := validators.NewMultiValidator(validators.Validators{psv1, psv2})

		h1, err := mv1.Hash()
		require.NoError(t, err)

		h2, err := mv2.Hash()
		require.NoError(t, err)

		h3, err := mv3.Hash()
		require.NoError(t, err)

		assert.NotEqual(t, h1, h2)
		assert.NotEqual(t, h1, h3)
		assert.NotEqual(t, h2, h3)
	})

	t.Run("Validate()", func(t *testing.T) {
		testSchema := []byte(`{
			"type": "object",
			"properties": {
				"seller": {
					"type": "string"
				}
			}, 
			"required": ["seller"]
		}`)

		psv1, _ := validators.NewProcessStepValidator("p1", "s1")
		psv2, _ := validators.NewProcessStepValidator("p2", "s2")
		sv, _ := validators.NewSchemaValidator(psv1, testSchema)

		mv := validators.NewMultiValidator(validators.Validators{psv1, psv2, sv})

		testCases := []struct {
			name string
			link *chainscript.Link
			err  error
		}{{
			"success",
			chainscripttest.NewLinkBuilder(t).WithProcess("p2").WithStep("s2").Build(),
			nil,
		}, {
			"no matching validator",
			chainscripttest.NewLinkBuilder(t).WithProcess("p3").Build(),
			validators.ErrNoMatchingValidator,
		}, {
			"child validation failed",
			chainscripttest.NewLinkBuilder(t).WithProcess("p1").WithStep("s1").WithData(t, map[string]string{
				"buyer": "bob",
			}).Build(),
			validators.ErrInvalidLinkSchema,
		}}

		for _, tt := range testCases {
			t.Run(tt.name, func(t *testing.T) {
				err := mv.Validate(context.Background(), nil, tt.link)
				if tt.err != nil {
					assert.EqualError(t, errors.Cause(err), tt.err.Error())
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})
}
