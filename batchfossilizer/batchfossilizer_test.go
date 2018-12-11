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

package batchfossilizer_test

import (
	"context"
	"encoding/hex"
	"testing"
	"time"

	json "github.com/gibson042/canonicaljson-go"
	"github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-core/batchfossilizer"
	"github.com/stratumn/go-core/batchfossilizer/evidences"
	"github.com/stratumn/go-core/dummyfossilizer"
	dummyevidences "github.com/stratumn/go-core/dummyfossilizer/evidences"
	"github.com/stratumn/go-core/fossilizer"
	"github.com/stratumn/go-core/fossilizer/dummyqueue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func batchProof(t *testing.T, e *chainscript.Evidence) *evidences.BatchProof {
	var proof evidences.BatchProof
	err := json.Unmarshal(e.Proof, &proof)
	require.NoError(t, err)

	return &proof
}

func TestGetInfo(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	f := batchfossilizer.New(
		ctx,
		&batchfossilizer.Config{
			Version: "1.0.0",
			Commit:  "abcdef",
		},
		dummyfossilizer.New(&dummyfossilizer.Config{}),
		dummyqueue.New(),
	)

	i, err := f.GetInfo(ctx)
	require.NoError(t, err)
	assert.Equal(t, "1.0.0", i.(*batchfossilizer.Info).Version)
	assert.Equal(t, "abcdef", i.(*batchfossilizer.Info).Commit)
}

func TestFossilize(t *testing.T) {
	t.Run("nothing to fossilize", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		f := dummyfossilizer.New(&dummyfossilizer.Config{})
		q := dummyqueue.New()
		batch := batchfossilizer.New(
			ctx,
			&batchfossilizer.Config{Interval: 3 * time.Millisecond},
			f,
			q,
		)

		eventChan := make(chan *fossilizer.Event, 1)
		batch.AddFossilizerEventChan(eventChan)

		select {
		case <-eventChan:
			assert.Fail(t, "unexpected fossilizer event")
		case <-time.After(5 * time.Millisecond):
		}
	})

	t.Run("single element", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		f := dummyfossilizer.New(&dummyfossilizer.Config{})
		q := dummyqueue.New()
		batch := batchfossilizer.New(
			ctx,
			&batchfossilizer.Config{Interval: 3 * time.Millisecond},
			f,
			q,
		)

		eventChan := make(chan *fossilizer.Event)
		batch.AddFossilizerEventChan(eventChan)

		err := batch.Fossilize(ctx, []byte("b4tm4n"), []byte("r0b1n"))
		require.NoError(t, err)

		e := <-eventChan
		assert.Equal(t, fossilizer.DidFossilize, e.EventType)

		r, ok := e.Data.(*fossilizer.Result)
		require.True(t, ok)
		assert.Equal(t, []byte("b4tm4n"), r.Data)
		assert.Equal(t, []byte("r0b1n"), r.Meta)

		// The queue should have been emptied.
		assert.Empty(t, q.Fossils())
	})

	t.Run("max leaves", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		f := dummyfossilizer.New(&dummyfossilizer.Config{})
		q := dummyqueue.New()
		batch := batchfossilizer.New(
			ctx,
			&batchfossilizer.Config{
				Interval:  3 * time.Millisecond,
				MaxLeaves: 2,
			},
			f,
			q,
		)

		eventChan := make(chan *fossilizer.Event, 3)
		batch.AddFossilizerEventChan(eventChan)

		for i := byte(0); i < 3; i++ {
			err := batch.Fossilize(ctx, []byte{i}, []byte{i + 10})
			require.NoError(t, err)
		}

		trees := make(map[string]struct{})
		for i := byte(0); i < 3; i++ {
			e := <-eventChan
			r, ok := e.Data.(*fossilizer.Result)
			require.True(t, ok)

			p := batchProof(t, &r.Evidence)
			trees[hex.EncodeToString(p.Root)] = struct{}{}
		}

		// The fossils should have been split into two merkle trees.
		assert.Len(t, trees, 2)
		assert.Empty(t, q.Fossils())
	})

	t.Run("simultaneous batches", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		f := dummyfossilizer.New(&dummyfossilizer.Config{})
		q := dummyqueue.New()
		batch := batchfossilizer.New(
			ctx,
			&batchfossilizer.Config{
				Interval:      3 * time.Millisecond,
				MaxLeaves:     2,
				MaxSimBatches: 2,
			},
			f,
			q,
		)

		eventChan := make(chan *fossilizer.Event, 5)
		batch.AddFossilizerEventChan(eventChan)

		for i := byte(0); i < 5; i++ {
			err := batch.Fossilize(ctx, []byte{i}, []byte{i + 10})
			require.NoError(t, err)
		}

		trees := make(map[string]struct{})
		for i := byte(0); i < 5; i++ {
			e := <-eventChan
			r, ok := e.Data.(*fossilizer.Result)
			require.True(t, ok)

			p := batchProof(t, &r.Evidence)
			trees[hex.EncodeToString(p.Root)] = struct{}{}
		}

		// The fossils should have been split into three merkle trees.
		assert.Len(t, trees, 3)
		assert.Empty(t, q.Fossils())
	})

	t.Run("fossilizer evidence", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		f := dummyfossilizer.New(&dummyfossilizer.Config{})
		q := dummyqueue.New()
		batch := batchfossilizer.New(
			ctx,
			&batchfossilizer.Config{
				Interval:  3 * time.Millisecond,
				MaxLeaves: 10,
			},
			f,
			q,
		)

		eventChan := make(chan *fossilizer.Event, 3)
		batch.AddFossilizerEventChan(eventChan)

		for i := byte(0); i < 3; i++ {
			err := batch.Fossilize(ctx, []byte{i}, []byte{i + 10})
			require.NoError(t, err)
		}

		now := time.Now().Unix()
		var root []byte
		for i := 0; i < 3; i++ {
			e := <-eventChan
			r, ok := e.Data.(*fossilizer.Result)
			require.True(t, ok)

			p := batchProof(t, &r.Evidence)
			if root == nil {
				root = p.Root
			}

			// All fossils should be in the same merkle tree.
			assert.Equal(t, root, p.Root)
			assert.InDelta(t, now, p.Timestamp, 10)
			assert.NoError(t, p.Path.Validate())

			// We should find an evidence produced by the wrapped fossilizer.
			var proof dummyevidences.DummyProof
			err := json.Unmarshal(p.Proof, &proof)
			require.NoError(t, err)
			assert.InDelta(t, now, proof.Timestamp, 10)
		}
	})
}
