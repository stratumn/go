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

package storetestcases

import (
	"context"
	"io/ioutil"
	"log"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-chainscript/chainscripttest"
	"github.com/stratumn/go-core/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCreateLink tests what happens when you create a new link.
func (f Factory) TestCreateLink(t *testing.T) {
	a := f.initAdapter(t)
	defer f.freeAdapter(a)

	t.Run("should not produce an error", func(t *testing.T) {
		ctx := context.Background()
		l := chainscripttest.RandomLink(t)
		_, err := a.CreateLink(ctx, l)
		assert.NoError(t, err, "a.CreateLink()")
	})

	t.Run("with no priority should not produce an error", func(t *testing.T) {
		ctx := context.Background()
		l := chainscripttest.RandomLink(t)
		l.Meta.Priority = 0.0

		_, err := a.CreateLink(ctx, l)
		assert.NoError(t, err, "a.CreateLink()")
	})

	t.Run("update data should not produce an error", func(t *testing.T) {
		ctx := context.Background()
		l := chainscripttest.RandomLink(t)
		_, err := a.CreateLink(ctx, l)
		require.NoError(t, err, "a.CreateLink()")

		err = l.SetData(chainscripttest.RandomString(32))
		require.NoError(t, err)

		_, err = a.CreateLink(ctx, l)
		require.NoError(t, err, "a.CreateLink()")
	})

	t.Run("update map ID should not produce an error", func(t *testing.T) {
		ctx := context.Background()
		l1 := chainscripttest.RandomLink(t)
		_, err := a.CreateLink(ctx, l1)
		require.NoError(t, err, "a.CreateLink()")

		l1.Meta.MapId = chainscripttest.RandomString(12)
		_, err = a.CreateLink(ctx, l1)
		require.NoError(t, err, "a.CreateLink()")
	})

	t.Run("with previous link hash should not produce an error", func(t *testing.T) {
		ctx := context.Background()
		l := chainscripttest.RandomLink(t)
		_, err := a.CreateLink(ctx, l)
		require.NoError(t, err, "a.CreateLink()")

		l = chainscripttest.NewLinkBuilder(t).Branch(t, l).Build()
		_, err = a.CreateLink(ctx, l)
		require.NoError(t, err, "a.CreateLink()")
	})

	t.Run("out degree", func(t *testing.T) {
		t.Run("0 prevents children", func(t *testing.T) {
			ctx := context.Background()
			l := chainscripttest.NewLinkBuilder(t).WithRandomData().WithDegree(0).Build()

			_, err := a.CreateLink(ctx, l)
			if err == store.ErrOutDegreeNotSupported {
				t.Skip("tested store doesn't support out degree yet")
			}

			require.NoError(t, err)

			child := chainscripttest.NewLinkBuilder(t).Branch(t, l).Build()
			childHash, err := child.Hash()
			require.NoError(t, err)

			_, err = a.CreateLink(ctx, child)
			assert.EqualError(t, err, chainscript.ErrOutDegree.Error())

			found, _ := a.GetSegment(ctx, childHash)
			assert.Nil(t, found)
		})

		t.Run("1 allows only one child", func(t *testing.T) {
			ctx := context.Background()
			l := chainscripttest.NewLinkBuilder(t).WithRandomData().WithDegree(1).Build()

			lh, err := a.CreateLink(ctx, l)
			if err == store.ErrOutDegreeNotSupported {
				t.Skip("tested store doesn't support out degree yet")
			}

			require.NoError(t, err)

			child1 := chainscripttest.NewLinkBuilder(t).WithRandomData().Branch(t, l).Build()
			child2 := chainscripttest.NewLinkBuilder(t).WithRandomData().Branch(t, l).Build()

			successChan := make(chan struct{})
			errChan := make(chan error)

			for _, child := range []*chainscript.Link{child1, child2} {
				go func(child *chainscript.Link) {
					_, err := a.CreateLink(ctx, child)
					if err != nil {
						errChan <- err
					} else {
						successChan <- struct{}{}
					}
				}(child)
			}

			select {
			case <-successChan:
			case <-time.After(100 * time.Millisecond):
				assert.Fail(t, "timeout before link created")
			}

			select {
			case err := <-errChan:
				assert.EqualError(t, err, chainscript.ErrOutDegree.Error())
			case <-time.After(100 * time.Millisecond):
				assert.Fail(t, "timeout before link creation failure")
			}

			children, err := a.FindSegments(ctx, &store.SegmentFilter{
				Pagination:   store.Pagination{Limit: 10},
				PrevLinkHash: lh,
			})
			require.NoError(t, err)
			require.Equal(t, 1, children.TotalCount)
			require.Len(t, children.Segments, 1)
		})

		t.Run("multiple children", func(t *testing.T) {
			ctx := context.Background()
			l := chainscripttest.NewLinkBuilder(t).WithRandomData().WithDegree(2).Build()

			lh, err := a.CreateLink(ctx, l)
			if err == store.ErrOutDegreeNotSupported {
				t.Skip("tested store doesn't support out degree yet")
			}

			require.NoError(t, err)

			child1 := chainscripttest.NewLinkBuilder(t).WithRandomData().Branch(t, l).WithPriority(1.).Build()
			child2 := chainscripttest.NewLinkBuilder(t).WithRandomData().Branch(t, l).WithPriority(2.).Build()
			child3 := chainscripttest.NewLinkBuilder(t).WithRandomData().Branch(t, l).Build()

			lh1, err := a.CreateLink(ctx, child1)
			require.NoError(t, err)

			// Trying to add a duplicate link should not increment the children count.
			// It should still be possible to add a second child after this call.
			_, _ = a.CreateLink(ctx, child1)

			lh2, err := a.CreateLink(ctx, child2)
			require.NoError(t, err)

			_, err = a.CreateLink(ctx, child3)
			require.EqualError(t, err, chainscript.ErrOutDegree.Error())

			children, err := a.FindSegments(ctx, &store.SegmentFilter{
				Pagination:   store.Pagination{Limit: 10},
				PrevLinkHash: lh,
			})
			require.NoError(t, err)
			require.Equal(t, 2, children.TotalCount)
			require.Len(t, children.Segments, 2)
			assert.Equal(t, lh2, children.Segments[0].LinkHash())
			assert.Equal(t, lh1, children.Segments[1].LinkHash())
		})
	})
}

// BenchmarkCreateLink benchmarks creating new links.
func (f Factory) BenchmarkCreateLink(b *testing.B) {
	a := f.initAdapterB(b)
	defer f.freeAdapter(a)

	slice := make([]*chainscript.Link, b.N)
	for i := 0; i < b.N; i++ {
		slice[i] = RandomLink(b, b.N, i)
	}

	b.ResetTimer()
	log.SetOutput(ioutil.Discard)

	for i := 0; i < b.N; i++ {
		if _, err := a.CreateLink(context.Background(), slice[i]); err != nil {
			b.Error(err)
		}
	}
}

// BenchmarkCreateLinkParallel benchmarks creating new links in parallel.
func (f Factory) BenchmarkCreateLinkParallel(b *testing.B) {
	a := f.initAdapterB(b)
	defer f.freeAdapter(a)

	slice := make([]*chainscript.Link, b.N)
	for i := 0; i < b.N; i++ {
		slice[i] = RandomLink(b, b.N, i)
	}

	var counter uint64

	b.ResetTimer()
	log.SetOutput(ioutil.Discard)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i := atomic.AddUint64(&counter, 1) - 1
			if _, err := a.CreateLink(context.Background(), slice[i]); err != nil {
				b.Error(err)
			}
		}
	})
}
