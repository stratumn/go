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

package types_test

import (
	"testing"

	"github.com/stratumn/go-chainscript/chainscripttest"
	"github.com/stratumn/go-core/types"
	"github.com/stretchr/testify/assert"
)

func TestSegmentSliceSort_Priority(t *testing.T) {
	t.Run("priority", func(t *testing.T) {
		slice := types.SegmentSlice{
			chainscripttest.NewLinkBuilder(t).WithPriority(2.3).Segmentify(t),
			chainscripttest.NewLinkBuilder(t).WithPriority(1.1).Segmentify(t),
			chainscripttest.NewLinkBuilder(t).WithPriority(3.33).Segmentify(t),
			chainscripttest.NewLinkBuilder(t).Segmentify(t),
		}

		slice.Sort(false)
		assert.Equal(t, 3.33, slice[0].Link.Meta.Priority)
		assert.Equal(t, 2.3, slice[1].Link.Meta.Priority)
		assert.Equal(t, 1.1, slice[2].Link.Meta.Priority)
		assert.Equal(t, 0.0, slice[3].Link.Meta.Priority)

		slice.Sort(true)
		assert.Equal(t, 0.0, slice[0].Link.Meta.Priority)
		assert.Equal(t, 1.1, slice[1].Link.Meta.Priority)
		assert.Equal(t, 2.3, slice[2].Link.Meta.Priority)
		assert.Equal(t, 3.33, slice[3].Link.Meta.Priority)
	})

	t.Run("link hash", func(t *testing.T) {
		slice := types.SegmentSlice{
			chainscripttest.NewLinkBuilder(t).WithPriority(2.0).WithAction("a1").Segmentify(t),
			chainscripttest.NewLinkBuilder(t).WithPriority(2.0).WithAction("a2").Segmentify(t),
		}

		slice.Sort(false)
		assert.True(t, slice[0].LinkHash().String() < slice[1].LinkHash().String())

		slice.Sort(true)
		assert.True(t, slice[0].LinkHash().String() > slice[1].LinkHash().String())
	})
}
