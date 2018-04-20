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

	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/cs/cstesting"
	"github.com/stretchr/testify/assert"
)

// TestCreateLink tests what happens when you create a new link.
func (f Factory) TestCreateLink(t *testing.T) {
	a := f.initAdapter(t)
	defer f.freeAdapter(a)

	t.Run("CreateLink should not produce an error", func(t *testing.T) {
		ctx := context.Background()
		l := cstesting.RandomLink()
		_, err := a.CreateLink(ctx, l)
		assert.NoError(t, err, "a.CreateLink()")
	})

	t.Run("CreateLink with no priority should not produce an error", func(t *testing.T) {
		ctx := context.Background()
		l := cstesting.RandomLink()
		delete(l.Meta.Data, "priority")

		_, err := a.CreateLink(ctx, l)
		assert.NoError(t, err, "a.CreateLink()")
	})

	t.Run("CreateLink and update state should not produce an error", func(t *testing.T) {
		ctx := context.Background()
		l := cstesting.RandomLink()
		_, err := a.CreateLink(ctx, l)
		assert.NoError(t, err, "a.CreateLink()")

		l = cstesting.ChangeState(l)
		_, err = a.CreateLink(ctx, l)
		assert.NoError(t, err, "a.CreateLink()")
	})

	t.Run("CreateLink and update map ID should not produce an error", func(t *testing.T) {
		ctx := context.Background()
		l1 := cstesting.RandomLink()
		_, err := a.CreateLink(ctx, l1)
		assert.NoError(t, err, "a.CreateLink()")

		l2 := cstesting.ChangeMapID(l1)
		_, err = a.CreateLink(ctx, l2)
		assert.NoError(t, err, "a.CreateLink()")

	})

	t.Run("CreateLink with previous link hash should not produce an error", func(t *testing.T) {
		ctx := context.Background()
		l := cstesting.RandomLink()
		_, err := a.CreateLink(ctx, l)
		assert.NoError(t, err, "a.CreateLink()")

		l = cstesting.NewLinkBuilder().Branch(l).Build()
		_, err = a.CreateLink(ctx, l)
		assert.NoError(t, err, "a.CreateLink()")
	})
}

// BenchmarkCreateLink benchmarks creating new links.
func (f Factory) BenchmarkCreateLink(b *testing.B) {
	a := f.initAdapterB(b)
	defer f.freeAdapter(a)

	slice := make([]*cs.Link, b.N)
	for i := 0; i < b.N; i++ {
		slice[i] = cstesting.RandomLink()
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

	slice := make([]*cs.Link, b.N)
	for i := 0; i < b.N; i++ {
		slice[i] = cstesting.RandomLink()
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
