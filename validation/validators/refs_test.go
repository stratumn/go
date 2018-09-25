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

package validators_test

import (
	"context"
	"testing"

	"github.com/stratumn/go-chainscript/chainscripttest"
	"github.com/stratumn/go-core/dummystore"
	"github.com/stratumn/go-core/validation/validators"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRefsValidator(t *testing.T) {
	ctx := context.Background()
	s := dummystore.New(&dummystore.Config{})
	l1 := chainscripttest.NewLinkBuilder(t).
		WithProcess("p1").
		WithMapID("m1").
		Build()
	l2 := chainscripttest.NewLinkBuilder(t).
		WithProcess("p2").
		WithMapID("m2").
		Build()
	l3 := chainscripttest.NewLinkBuilder(t).
		WithProcess("p2").
		WithMapID("m3").
		Build()

	h1, err := s.CreateLink(ctx, l1)
	require.NoError(t, err)
	_, err = s.CreateLink(ctx, l2)
	require.NoError(t, err)
	_, err = s.CreateLink(ctx, l3)
	require.NoError(t, err)

	v := validators.NewRefsValidator()

	t.Run("without references", func(t *testing.T) {
		l := chainscripttest.NewLinkBuilder(t).Build()
		assert.True(t, v.ShouldValidate(l))
		assert.NoError(t, v.Validate(ctx, nil, l))
	})

	t.Run("with parent", func(t *testing.T) {
		l := chainscripttest.NewLinkBuilder(t).
			WithProcess("p1").
			WithMapID("m1").
			WithParentHash(h1).
			Build()

		assert.True(t, v.ShouldValidate(l))
		assert.NoError(t, v.Validate(ctx, s, l))
	})

	t.Run("with missing parent", func(t *testing.T) {
		l := chainscripttest.NewLinkBuilder(t).
			WithParentHash(chainscripttest.RandomHash()).
			Build()

		assert.True(t, v.ShouldValidate(l))
		assert.EqualError(t, v.Validate(ctx, s, l), validators.ErrParentNotFound.Error())
	})

	t.Run("with invalid parent mapID", func(t *testing.T) {
		l := chainscripttest.NewLinkBuilder(t).
			WithProcess("p1").
			WithMapID("other_map").
			WithParentHash(h1).
			Build()

		assert.True(t, v.ShouldValidate(l))
		assert.EqualError(t, v.Validate(ctx, s, l), validators.ErrMapIDMismatch.Error())
	})

	t.Run("with invalid parent process", func(t *testing.T) {
		l := chainscripttest.NewLinkBuilder(t).
			WithProcess("other_process").
			WithMapID("m1").
			WithParentHash(h1).
			Build()

		assert.True(t, v.ShouldValidate(l))
		assert.EqualError(t, v.Validate(ctx, s, l), validators.ErrProcessMismatch.Error())
	})

	t.Run("with references", func(t *testing.T) {
		l := chainscripttest.NewLinkBuilder(t).
			WithProcess("test_proc").
			WithMapID("test_map").
			WithRef(t, l2).
			WithRef(t, l3).
			Build()

		assert.True(t, v.ShouldValidate(l))
		assert.NoError(t, v.Validate(ctx, s, l))
	})

	t.Run("with invalid ref process", func(t *testing.T) {
		l := chainscripttest.NewLinkBuilder(t).
			WithProcess("test_proc").
			WithMapID("test_map").
			WithRef(t, l2).
			Build()

		l.Meta.Refs[0].Process = "other_process"

		assert.True(t, v.ShouldValidate(l))
		assert.EqualError(t, v.Validate(ctx, s, l), validators.ErrProcessMismatch.Error())
	})

	t.Run("with missing ref", func(t *testing.T) {
		l := chainscripttest.NewLinkBuilder(t).
			WithProcess("test_proc").
			WithMapID("test_map").
			WithRef(t, l2).
			Build()

		l.Meta.Refs[0].LinkHash = chainscripttest.RandomHash()

		assert.True(t, v.ShouldValidate(l))
		assert.EqualError(t, v.Validate(ctx, s, l), validators.ErrRefNotFound.Error())
	})
}
