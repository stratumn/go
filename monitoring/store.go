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

package monitoring

import (
	"context"
	"fmt"

	"github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-core/store"
	"github.com/stratumn/go-core/types"
)

// StoreAdapter is a decorator for the store.Adapter interface.
// It wraps a real store.Adapter implementation and adds instrumentation.
type StoreAdapter struct {
	s    store.Adapter
	name string
}

// WrapStore wraps an existing store adapter to add monitoring.
func WrapStore(s store.Adapter, name string) store.Adapter {
	return &StoreAdapter{s: s, name: name}
}

// GetInfo instruments the call and delegates to the underlying store.
func (a *StoreAdapter) GetInfo(ctx context.Context) (res interface{}, err error) {
	tracker := newStoreRequestTracker("GetInfo")
	span, ctx := StartSpanIncomingRequest(ctx, fmt.Sprintf("%s/GetInfo", a.name))
	defer func() {
		SetSpanStatusAndEnd(span, err)
		tracker.End(err)
	}()

	res, err = a.s.GetInfo(ctx)
	return
}

// AddStoreEventChannel instruments the call and delegates to the underlying store.
func (a *StoreAdapter) AddStoreEventChannel(c chan *store.Event) {
	a.s.AddStoreEventChannel(c)
}

// NewBatch instruments the call and delegates to the underlying store.
func (a *StoreAdapter) NewBatch(ctx context.Context) (b store.Batch, err error) {
	span, ctx := StartSpanIncomingRequest(ctx, fmt.Sprintf("%s/NewBatch", a.name))
	defer func() {
		SetSpanStatusAndEnd(span, err)
	}()

	b, err = a.s.NewBatch(ctx)
	return
}

// AddEvidence instruments the call and delegates to the underlying store.
func (a *StoreAdapter) AddEvidence(ctx context.Context, linkHash chainscript.LinkHash, evidence *chainscript.Evidence) (err error) {
	tracker := newStoreRequestTracker("AddEvidence")
	span, ctx := StartSpanIncomingRequest(ctx, fmt.Sprintf("%s/AddEvidence", a.name))
	defer func() {
		SetSpanStatusAndEnd(span, err)
		tracker.End(err)
	}()

	err = a.s.AddEvidence(ctx, linkHash, evidence)
	return
}

// GetEvidences instruments the call and delegates to the underlying store.
func (a *StoreAdapter) GetEvidences(ctx context.Context, linkHash chainscript.LinkHash) (e types.EvidenceSlice, err error) {
	tracker := newStoreRequestTracker("GetEvidences")
	span, ctx := StartSpanIncomingRequest(ctx, fmt.Sprintf("%s/GetEvidences", a.name))
	defer func() {
		SetSpanStatusAndEnd(span, err)
		tracker.End(err)
	}()

	e, err = a.s.GetEvidences(ctx, linkHash)
	return
}

// CreateLink instruments the call and delegates to the underlying store.
func (a *StoreAdapter) CreateLink(ctx context.Context, link *chainscript.Link) (lh chainscript.LinkHash, err error) {
	tracker := newStoreRequestTracker("CreateLink")
	span, ctx := StartSpanIncomingRequest(ctx, fmt.Sprintf("%s/CreateLink", a.name))
	defer func() {
		SetSpanStatusAndEnd(span, err)
		tracker.End(err)
	}()

	lh, err = a.s.CreateLink(ctx, link)
	return
}

// GetSegment instruments the call and delegates to the underlying store.
func (a *StoreAdapter) GetSegment(ctx context.Context, linkHash chainscript.LinkHash) (s *chainscript.Segment, err error) {
	tracker := newStoreRequestTracker("GetSegment")
	span, ctx := StartSpanIncomingRequest(ctx, fmt.Sprintf("%s/GetSegment", a.name))
	defer func() {
		SetSpanStatusAndEnd(span, err)
		tracker.End(err)
	}()

	s, err = a.s.GetSegment(ctx, linkHash)
	return
}

// FindSegments instruments the call and delegates to the underlying store.
func (a *StoreAdapter) FindSegments(ctx context.Context, filter *store.SegmentFilter) (ss *types.PaginatedSegments, err error) {
	tracker := newStoreRequestTracker("FindSegments")
	span, ctx := StartSpanIncomingRequest(ctx, fmt.Sprintf("%s/FindSegments", a.name))
	defer func() {
		SetSpanStatusAndEnd(span, err)
		tracker.End(err)
	}()

	ss, err = a.s.FindSegments(ctx, filter)
	return
}

// GetMapIDs instruments the call and delegates to the underlying store.
func (a *StoreAdapter) GetMapIDs(ctx context.Context, filter *store.MapFilter) (mids []string, err error) {
	tracker := newStoreRequestTracker("GetMapIDs")
	span, ctx := StartSpanIncomingRequest(ctx, fmt.Sprintf("%s/GetMapIDs", a.name))
	defer func() {
		SetSpanStatusAndEnd(span, err)
		tracker.End(err)
	}()

	mids, err = a.s.GetMapIDs(ctx, filter)
	return
}

// KeyValueStoreAdapter is a decorator for the store.KeyValueStore interface.
// It wraps a real store.KeyValueStore implementation and adds instrumentation.
type KeyValueStoreAdapter struct {
	s    store.KeyValueStore
	name string
}

// WrapKeyValueStore wraps an existing key value store adapter to add
// monitoring.
func WrapKeyValueStore(s store.KeyValueStore, name string) store.KeyValueStore {
	return &KeyValueStoreAdapter{s: s, name: name}
}

// GetValue instruments the call and delegates to the underlying store.
func (a *KeyValueStoreAdapter) GetValue(ctx context.Context, key []byte) (v []byte, err error) {
	tracker := newStoreRequestTracker("GetValue")
	span, ctx := StartSpanIncomingRequest(ctx, fmt.Sprintf("%s/GetValue", a.name))
	defer func() {
		SetSpanStatusAndEnd(span, err)
		tracker.End(err)
	}()

	v, err = a.s.GetValue(ctx, key)
	return
}

// SetValue instruments the call and delegates to the underlying store.
func (a *KeyValueStoreAdapter) SetValue(ctx context.Context, key []byte, value []byte) (err error) {
	tracker := newStoreRequestTracker("SetValue")
	span, ctx := StartSpanIncomingRequest(ctx, fmt.Sprintf("%s/SetValue", a.name))
	defer func() {
		SetSpanStatusAndEnd(span, err)
		tracker.End(err)
	}()

	err = a.s.SetValue(ctx, key, value)
	return
}

// DeleteValue instruments the call and delegates to the underlying store.
func (a *KeyValueStoreAdapter) DeleteValue(ctx context.Context, key []byte) (v []byte, err error) {
	tracker := newStoreRequestTracker("DeleteValue")
	span, ctx := StartSpanIncomingRequest(ctx, fmt.Sprintf("%s/DeleteValue", a.name))
	defer func() {
		SetSpanStatusAndEnd(span, err)
		tracker.End(err)
	}()

	v, err = a.s.DeleteValue(ctx, key)
	return
}
