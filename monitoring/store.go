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

package monitoring

import (
	"context"
	"fmt"

	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/types"

	"go.opencensus.io/trace"
)

// StoreAdapter is a decorator for the store.Adapter interface.
// It wraps a real store.Adapter implementation and adds instrumentation.
type StoreAdapter struct {
	s    store.Adapter
	name string
}

// NewStoreAdapter decorates an existing store adapter.
func NewStoreAdapter(s store.Adapter, name string) store.Adapter {
	return &StoreAdapter{s: s, name: name}
}

// GetInfo instruments the call and delegates to the underlying store.
func (a *StoreAdapter) GetInfo(ctx context.Context) (res interface{}, err error) {
	ctx, span := trace.StartSpan(ctx, fmt.Sprintf("%s/GetInfo", a.name))
	defer SetSpanStatusAndEnd(span, err)

	res, err = a.s.GetInfo(ctx)
	return
}

// AddStoreEventChannel instruments the call and delegates to the underlying store.
func (a *StoreAdapter) AddStoreEventChannel(c chan *store.Event) {
	a.s.AddStoreEventChannel(c)
}

// NewBatch instruments the call and delegates to the underlying store.
func (a *StoreAdapter) NewBatch(ctx context.Context) (b store.Batch, err error) {
	ctx, span := trace.StartSpan(ctx, fmt.Sprintf("%s/NewBatch", a.name))
	defer SetSpanStatusAndEnd(span, err)

	b, err = a.s.NewBatch(ctx)
	return
}

// AddEvidence instruments the call and delegates to the underlying store.
func (a *StoreAdapter) AddEvidence(ctx context.Context, linkHash *types.Bytes32, evidence *cs.Evidence) (err error) {
	ctx, span := trace.StartSpan(ctx, fmt.Sprintf("%s/AddEvidence", a.name))
	defer SetSpanStatusAndEnd(span, err)

	err = a.s.AddEvidence(ctx, linkHash, evidence)
	return
}

// GetEvidences instruments the call and delegates to the underlying store.
func (a *StoreAdapter) GetEvidences(ctx context.Context, linkHash *types.Bytes32) (e *cs.Evidences, err error) {
	ctx, span := trace.StartSpan(ctx, fmt.Sprintf("%s/GetEvidences", a.name))
	defer SetSpanStatusAndEnd(span, err)

	e, err = a.s.GetEvidences(ctx, linkHash)
	return
}

// CreateLink instruments the call and delegates to the underlying store.
func (a *StoreAdapter) CreateLink(ctx context.Context, link *cs.Link) (lh *types.Bytes32, err error) {
	ctx, span := trace.StartSpan(ctx, fmt.Sprintf("%s/CreateLink", a.name))
	defer SetSpanStatusAndEnd(span, err)

	lh, err = a.s.CreateLink(ctx, link)
	return
}

// GetSegment instruments the call and delegates to the underlying store.
func (a *StoreAdapter) GetSegment(ctx context.Context, linkHash *types.Bytes32) (s *cs.Segment, err error) {
	ctx, span := trace.StartSpan(ctx, fmt.Sprintf("%s/GetSegment", a.name))
	defer SetSpanStatusAndEnd(span, err)

	s, err = a.s.GetSegment(ctx, linkHash)
	return
}

// FindSegments instruments the call and delegates to the underlying store.
func (a *StoreAdapter) FindSegments(ctx context.Context, filter *store.SegmentFilter) (ss cs.SegmentSlice, err error) {
	ctx, span := trace.StartSpan(ctx, fmt.Sprintf("%s/FindSegments", a.name))
	defer SetSpanStatusAndEnd(span, err)

	ss, err = a.s.FindSegments(ctx, filter)
	return
}

// GetMapIDs instruments the call and delegates to the underlying store.
func (a *StoreAdapter) GetMapIDs(ctx context.Context, filter *store.MapFilter) (mids []string, err error) {
	ctx, span := trace.StartSpan(ctx, fmt.Sprintf("%s/GetMapIDs", a.name))
	defer SetSpanStatusAndEnd(span, err)

	mids, err = a.s.GetMapIDs(ctx, filter)
	return
}

// KeyValueStoreAdapter is a decorator for the store.KeyValueStore interface.
// It wraps a real store.KeyValueStore implementation and adds instrumentation.
type KeyValueStoreAdapter struct {
	s    store.KeyValueStore
	name string
}

// NewKeyValueStoreAdapter decorates an existing key value store adapter.
func NewKeyValueStoreAdapter(s store.KeyValueStore, name string) store.KeyValueStore {
	return &KeyValueStoreAdapter{s: s, name: name}
}

// GetValue instruments the call and delegates to the underlying store.
func (a *KeyValueStoreAdapter) GetValue(ctx context.Context, key []byte) (v []byte, err error) {
	ctx, span := trace.StartSpan(ctx, fmt.Sprintf("%s/GetValue", a.name))
	defer SetSpanStatusAndEnd(span, err)

	v, err = a.s.GetValue(ctx, key)
	return
}

// SetValue instruments the call and delegates to the underlying store.
func (a *KeyValueStoreAdapter) SetValue(ctx context.Context, key []byte, value []byte) (err error) {
	ctx, span := trace.StartSpan(ctx, fmt.Sprintf("%s/SetValue", a.name))
	defer SetSpanStatusAndEnd(span, err)

	err = a.s.SetValue(ctx, key, value)
	return
}

// DeleteValue instruments the call and delegates to the underlying store.
func (a *KeyValueStoreAdapter) DeleteValue(ctx context.Context, key []byte) (v []byte, err error) {
	ctx, span := trace.StartSpan(ctx, fmt.Sprintf("%s/DeleteValue", a.name))
	defer SetSpanStatusAndEnd(span, err)

	v, err = a.s.DeleteValue(ctx, key)
	return
}
