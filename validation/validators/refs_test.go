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
	"github.com/stratumn/go-core/dummystore"
	"github.com/stratumn/go-core/testutil"
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

	t.Run("with parent valid out degree", func(t *testing.T) {
		parent := chainscripttest.NewLinkBuilder(t).
			WithRandomData().
			WithDegree(3).
			Build()
		_, err := s.CreateLink(ctx, parent)
		require.NoError(t, err)

		child := chainscripttest.NewLinkBuilder(t).
			WithParent(t, parent).
			WithProcess(parent.Meta.Process.Name).
			WithMapID(parent.Meta.MapId).
			Build()

		assert.True(t, v.ShouldValidate(child))
		assert.NoError(t, v.Validate(ctx, s, child))
	})

	t.Run("with missing parent", func(t *testing.T) {
		l := chainscripttest.NewLinkBuilder(t).
			WithParentHash(chainscripttest.RandomHash()).
			Build()

		assert.True(t, v.ShouldValidate(l))
		testutil.AssertWrappedErrorEqual(t, v.Validate(ctx, s, l), validators.ErrParentNotFound)
	})

	t.Run("with invalid parent mapID", func(t *testing.T) {
		l := chainscripttest.NewLinkBuilder(t).
			WithProcess("p1").
			WithMapID("other_map").
			WithParentHash(h1).
			Build()

		assert.True(t, v.ShouldValidate(l))
		testutil.AssertWrappedErrorEqual(t, v.Validate(ctx, s, l), validators.ErrMapIDMismatch)
	})

	t.Run("with invalid parent process", func(t *testing.T) {
		l := chainscripttest.NewLinkBuilder(t).
			WithProcess("other_process").
			WithMapID("m1").
			WithParentHash(h1).
			Build()

		assert.True(t, v.ShouldValidate(l))
		testutil.AssertWrappedErrorEqual(t, v.Validate(ctx, s, l), validators.ErrProcessMismatch)
	})

	t.Run("with parent that wants no children", func(t *testing.T) {
		parent := chainscripttest.NewLinkBuilder(t).
			WithRandomData().
			WithDegree(0).
			Build()
		_, err := s.CreateLink(ctx, parent)
		require.NoError(t, err)

		child := chainscripttest.NewLinkBuilder(t).
			WithParent(t, parent).
			WithProcess(parent.Meta.Process.Name).
			WithMapID(parent.Meta.MapId).
			Build()

		assert.True(t, v.ShouldValidate(child))
		testutil.AssertWrappedErrorEqual(t, v.Validate(ctx, s, child), chainscript.ErrOutDegree)
	})

	t.Run("with parent that has too many children", func(t *testing.T) {
		parent := chainscripttest.NewLinkBuilder(t).
			WithRandomData().
			WithDegree(1).
			Build()
		_, err := s.CreateLink(ctx, parent)
		require.NoError(t, err)

		child1 := chainscripttest.NewLinkBuilder(t).
			WithParent(t, parent).
			WithProcess(parent.Meta.Process.Name).
			WithMapID(parent.Meta.MapId).
			Build()
		_, err = s.CreateLink(ctx, child1)
		require.NoError(t, err)

		child2 := chainscripttest.NewLinkBuilder(t).
			WithRandomData().
			WithParent(t, parent).
			WithProcess(parent.Meta.Process.Name).
			WithMapID(parent.Meta.MapId).
			Build()

		assert.True(t, v.ShouldValidate(child2))
		testutil.AssertWrappedErrorEqual(t, v.Validate(ctx, s, child2), chainscript.ErrOutDegree)
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
		testutil.AssertWrappedErrorEqual(t, v.Validate(ctx, s, l), validators.ErrProcessMismatch)
	})

	t.Run("with missing ref", func(t *testing.T) {
		l := chainscripttest.NewLinkBuilder(t).
			WithProcess("test_proc").
			WithMapID("test_map").
			WithRef(t, l2).
			Build()

		l.Meta.Refs[0].LinkHash = chainscripttest.RandomHash()

		assert.True(t, v.ShouldValidate(l))
		testutil.AssertWrappedErrorEqual(t, v.Validate(ctx, s, l), validators.ErrRefNotFound)
	})
}
