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

package storetestcases

import (
	"context"
	"testing"

	"github.com/stratumn/go-chainscript/chainscripttest"
	"github.com/stratumn/go-core/monitoring/errorcode"
	"github.com/stratumn/go-core/store"
	"github.com/stratumn/go-core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAdapterConfig tests the implementation of the AdapterConfig interface.
// Stores that don't implement this interface are skipped.
func (f Factory) TestAdapterConfig(t *testing.T) {
	a := f.initAdapter(t)
	defer f.freeAdapter(a)

	cfg, ok := a.(store.AdapterConfig)
	if !ok {
		t.Skip("tested store doesn't support advanced adapter configuration")
	}

	t.Run("enforce unique map entry", func(t *testing.T) {
		ctx := context.Background()
		l1 := chainscripttest.NewLinkBuilder(t).
			WithProcess("test_process").
			WithMapID("test_map_1").
			WithoutParent().
			WithAction("init1").
			Build()

		_, err := a.CreateLink(ctx, l1)
		require.NoError(t, err)

		l2 := chainscripttest.NewLinkBuilder(t).
			WithProcess("test_process").
			WithMapID("test_map_1").
			WithoutParent().
			WithAction("init2").
			Build()

		_, err = a.CreateLink(ctx, l2)
		require.NoError(t, err)

		err = cfg.EnforceUniqueMapEntry()
		require.NoError(t, err)

		ll1 := chainscripttest.NewLinkBuilder(t).
			WithProcess("test_process").
			WithMapID("test_map_2").
			WithoutParent().
			WithAction("init1").
			Build()

		llh1, err := a.CreateLink(ctx, ll1)
		require.NoError(t, err)

		s1, err := a.GetSegment(ctx, llh1)
		require.NoError(t, err)
		chainscripttest.LinksEqual(t, ll1, s1.Link)

		ll2 := chainscripttest.NewLinkBuilder(t).
			WithProcess("test_process").
			WithMapID("test_map_2").
			WithoutParent().
			WithAction("init2").
			Build()

		llh2, err := a.CreateLink(ctx, ll2)
		require.NotNil(t, err)
		require.IsType(t, &types.Error{}, err)
		assert.Equal(t, errorcode.FailedPrecondition, err.(*types.Error).Code)
		require.EqualError(t, err.(*types.Error).Wrapped, store.ErrUniqueMapEntry.Error())

		s2, err := a.GetSegment(ctx, llh2)
		require.NoError(t, err)
		assert.Nil(t, s2)
	})
}
