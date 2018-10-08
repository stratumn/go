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
	"github.com/stratumn/go-core/dummystore"
	"github.com/stratumn/go-core/validation/validators"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransitionValidator(t *testing.T) {
	t.Run("Validate()", func(t *testing.T) {
		ctx := context.Background()
		store := dummystore.New(nil)
		require.NotNil(t, store)

		appendLink := func(parent *chainscript.Link, step string) *chainscript.Link {
			l := chainscripttest.NewLinkBuilder(t).
				Branch(t, parent).
				WithStep(step).
				Build()
			_, err := store.CreateLink(ctx, l)
			require.NoError(t, err)
			return l
		}

		testProcess := "product"
		initStep := "init"
		updateStep := "update"
		signStep := "sign"

		initLink := chainscripttest.NewLinkBuilder(t).
			WithProcess(testProcess).
			WithStep(initStep).
			WithoutParent().
			Build()
		_, err := store.CreateLink(ctx, initLink)
		require.NoError(t, err)

		updateLink := appendLink(initLink, updateStep)
		signLink := appendLink(updateLink, signStep)

		testCases := []struct {
			name        string
			link        *chainscript.Link
			transitions []string
			valid       bool
			err         error
		}{{
			name:        "valid init",
			link:        initLink,
			transitions: []string{""},
			valid:       true,
		}, {
			name:        "invalid init",
			link:        initLink,
			transitions: nil,
			valid:       false,
			err:         validators.ErrInvalidTransition,
		}, {
			name:        "no possible transition",
			link:        updateLink,
			transitions: nil,
			valid:       false,
			err:         validators.ErrInvalidTransition,
		}, {
			name:        "valid update transition",
			link:        updateLink,
			transitions: []string{updateStep, initStep},
			valid:       true,
		}, {
			name:        "valid sign transition",
			link:        signLink,
			transitions: []string{updateStep},
			valid:       true,
		}, {
			name:        "invalid sign transition",
			link:        signLink,
			transitions: []string{initStep},
			valid:       false,
			err:         validators.ErrInvalidTransition,
		}, {
			name:        "missing parent",
			link:        chainscripttest.NewLinkBuilder(t).From(t, updateLink).WithParentHash(chainscripttest.RandomHash()).Build(),
			transitions: []string{initStep},
			valid:       false,
		}}

		for _, tt := range testCases {
			t.Run(tt.name, func(t *testing.T) {
				psv, err := validators.NewProcessStepValidator(testProcess, tt.link.Meta.Step)
				require.NoError(t, err)

				v := validators.NewTransitionValidator(psv, tt.transitions)
				err = v.Validate(context.Background(), store, tt.link)

				if tt.valid {
					assert.NoError(t, err)
				} else {
					if tt.err != nil {
						assert.EqualError(t, errors.Cause(err), tt.err.Error())
					} else {
						assert.Error(t, err)
					}
				}
			})
		}
	})

	t.Run("ShouldValidate()", func(t *testing.T) {
		psv, err := validators.NewProcessStepValidator("p", "s")
		require.NoError(t, err)

		tv := validators.NewTransitionValidator(psv, nil)

		testCases := []struct {
			name  string
			link  *chainscript.Link
			match bool
		}{{
			"process mismatch",
			chainscripttest.NewLinkBuilder(t).WithProcess("some_process").WithStep("s").Build(),
			false,
		}, {
			"step mismatch",
			chainscripttest.NewLinkBuilder(t).WithProcess("p").WithStep("some_step").Build(),
			false,
		}, {
			"match",
			chainscripttest.NewLinkBuilder(t).WithProcess("p").WithStep("s").Build(),
			true,
		}}

		for _, tt := range testCases {
			t.Run(tt.name, func(t *testing.T) {
				match := tv.ShouldValidate(tt.link)
				assert.Equal(t, tt.match, match)
			})
		}
	})

	t.Run("Hash()", func(t *testing.T) {
		psv1, err := validators.NewProcessStepValidator("p1", "s1")
		require.NoError(t, err)

		psv2, err := validators.NewProcessStepValidator("p2", "s2")
		require.NoError(t, err)

		v1 := validators.NewTransitionValidator(psv1, []string{"a", "b"})
		v2 := validators.NewTransitionValidator(psv1, []string{"a", "c"})
		v3 := validators.NewTransitionValidator(psv2, []string{"a", "b"})

		h1, err := v1.Hash()
		require.NoError(t, err)

		h2, err := v2.Hash()
		require.NoError(t, err)

		h3, err := v3.Hash()
		require.NoError(t, err)

		assert.NotEqual(t, h1, h2)
		assert.NotEqual(t, h1, h3)
		assert.NotEqual(t, h2, h3)
	})
}
