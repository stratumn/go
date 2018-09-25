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

package storetesting

import (
	"context"

	"github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-core/store"
	"github.com/stratumn/go-core/types"
)

// MockAdapter is used to mock a store.
// It implements github.com/stratumn/go-core/store.Adapter.
type MockAdapter struct {
	// The mock for the GetInfo function.
	MockGetInfo MockGetInfo

	// The mock for the MockAddStoreEventChannel function.
	MockAddStoreEventChannel MockAddStoreEventChannel

	// The mock for the CreateLink function
	MockCreateLink MockCreateLink

	// The mock for the AddEvidence function
	MockAddEvidence MockAddEvidence

	// The mock for the GetSegment function.
	MockGetSegment MockGetSegment

	// The mock for the GetEvidences function
	MockGetEvidences MockGetEvidences

	// The mock for the FindSegments function.
	MockFindSegments MockFindSegments

	// The mock for the GetMapIDs function.
	MockGetMapIDs MockGetMapIDs

	// The mock for the NewBatch function.
	MockNewBatch MockNewBatch
}

// MockKeyValueStore is used to mock a key-value store.
// It implements github.com/stratumn/go-core/store.KeyValueStore.
type MockKeyValueStore struct {
	// The mock for the SetValue function.
	MockSetValue MockSetValue

	// The mock for the GetValue function.
	MockGetValue MockGetValue

	// The mock for the DeleteValue function.
	MockDeleteValue MockDeleteValue
}

// MockGetInfo mocks the GetInfo function.
type MockGetInfo struct {
	// The number of times the function was called.
	CalledCount int

	// An optional implementation of the function.
	Fn func() (interface{}, error)
}

// MockAddStoreEventChannel mocks the AddStoreEventChannel function.
type MockAddStoreEventChannel struct {
	// The number of times the function was called.
	CalledCount int

	// The event that was passed to each call.
	CalledWith []chan *store.Event

	// The last event that was passed.
	LastCalledWith chan *store.Event

	// An optional implementation of the function.
	Fn func(chan *store.Event)
}

// MockCreateLink mocks the CreateLink function.
type MockCreateLink struct {
	// The number of times the function was called.
	CalledCount int

	// The link that was passed to each call.
	CalledWith []*chainscript.Link

	// The last link that was passed.
	LastCalledWith *chainscript.Link

	// An optional implementation of the function.
	Fn func(*chainscript.Link) (chainscript.LinkHash, error)
}

// MockAddEvidence mocks the AddEvidence function.
type MockAddEvidence struct {
	// The number of times the function was called.
	CalledCount int

	// The evidence that was passed to each call.
	CalledWith []*chainscript.Evidence

	// The last evidence that was passed.
	LastCalledWith *chainscript.Evidence

	// An optional implementation of the function.
	Fn func(linkHash chainscript.LinkHash, evidence *chainscript.Evidence) error
}

// MockGetSegment mocks the GetSegment function.
type MockGetSegment struct {
	// The number of times the function was called.
	CalledCount int

	// The link hash that was passed to each call.
	CalledWith []chainscript.LinkHash

	// The last link hash that was passed.
	LastCalledWith chainscript.LinkHash

	// An optional implementation of the function.
	Fn func(chainscript.LinkHash) (*chainscript.Segment, error)
}

// MockGetEvidences mocks the GetEvidences function.
type MockGetEvidences struct {
	// The number of times the function was called.
	CalledCount int

	// The link hash that was passed to each call.
	CalledWith []chainscript.LinkHash

	// The last link hash that was passed.
	LastCalledWith chainscript.LinkHash

	// An optional implementation of the function.
	Fn func(chainscript.LinkHash) (types.EvidenceSlice, error)
}

// MockFindSegments mocks the FindSegments function.
type MockFindSegments struct {
	// The number of times the function was called.
	CalledCount int

	// The filter that was passed to each call.
	CalledWith []*store.SegmentFilter

	// The last filter that was passed.
	LastCalledWith *store.SegmentFilter

	// An optional implementation of the function.
	Fn func(*store.SegmentFilter) (*types.PaginatedSegments, error)
}

// MockGetMapIDs mocks the GetMapIDs function.
type MockGetMapIDs struct {
	// The number of times the function was called.
	CalledCount int

	// The pagination that was passed to each call.
	CalledWith []*store.MapFilter

	// The last pagination that was passed.
	LastCalledWith *store.MapFilter

	// An optional implementation of the function.
	Fn func(*store.MapFilter) ([]string, error)
}

// MockNewBatch mocks the NewBatch function.
type MockNewBatch struct {
	// The number of times the function was called.
	CalledCount int

	// An optional implementation of the function.
	Fn func() store.Batch
}

// MockSetValue mocks the SetValue function.
type MockSetValue struct {
	// The number of times the function was called.
	CalledCount int

	// The segment that was passed to each call.
	CalledWith [][][]byte

	// The last segment that was passed.
	LastCalledWith [][]byte

	// An optional implementation of the function.
	Fn func(key, value []byte) error
}

// MockGetValue mocks the GetValue function.
type MockGetValue struct {
	// The number of times the function was called.
	CalledCount int

	// The link hash that was passed to each call.
	CalledWith [][]byte

	// The last link hash that was passed.
	LastCalledWith []byte

	// An optional implementation of the function.
	Fn func([]byte) ([]byte, error)
}

// MockDeleteValue mocks the DeleteValue function.
type MockDeleteValue struct {
	// The number of times the function was called.
	CalledCount int

	// The key that was passed to each call.
	CalledWith [][]byte

	// The last link hash that was passed.
	LastCalledWith []byte

	// An optional implementation of the function.
	Fn func([]byte) ([]byte, error)
}

// GetInfo implements github.com/stratumn/go-core/store.Adapter.GetInfo.
func (a *MockAdapter) GetInfo(ctx context.Context) (interface{}, error) {
	a.MockGetInfo.CalledCount++

	if a.MockGetInfo.Fn != nil {
		return a.MockGetInfo.Fn()
	}

	return nil, nil
}

// AddStoreEventChannel implements
// github.com/stratumn/go-core/store.Adapter.AddStoreEventChannel.
func (a *MockAdapter) AddStoreEventChannel(storeChan chan *store.Event) {
	a.MockAddStoreEventChannel.CalledCount++
	a.MockAddStoreEventChannel.CalledWith = append(a.MockAddStoreEventChannel.CalledWith, storeChan)
	a.MockAddStoreEventChannel.LastCalledWith = storeChan

	if a.MockAddStoreEventChannel.Fn != nil {
		a.MockAddStoreEventChannel.Fn(storeChan)
	}
}

// CreateLink implements github.com/stratumn/go-core/store.Adapter.CreateLink.
func (a *MockAdapter) CreateLink(ctx context.Context, link *chainscript.Link) (chainscript.LinkHash, error) {
	a.MockCreateLink.CalledCount++
	a.MockCreateLink.CalledWith = append(a.MockCreateLink.CalledWith, link)
	a.MockCreateLink.LastCalledWith = link

	if a.MockCreateLink.Fn != nil {
		return a.MockCreateLink.Fn(link)
	}

	return nil, nil
}

// AddEvidence implements github.com/stratumn/go-core/store.Adapter.AddEvidence.
func (a *MockAdapter) AddEvidence(ctx context.Context, linkHash chainscript.LinkHash, evidence *chainscript.Evidence) error {
	a.MockAddEvidence.CalledCount++
	a.MockAddEvidence.CalledWith = append(a.MockAddEvidence.CalledWith, evidence)
	a.MockAddEvidence.LastCalledWith = evidence

	if a.MockAddEvidence.Fn != nil {
		return a.MockAddEvidence.Fn(linkHash, evidence)
	}

	return nil
}

// GetSegment implements github.com/stratumn/go-core/store.Adapter.GetSegment.
func (a *MockAdapter) GetSegment(ctx context.Context, linkHash chainscript.LinkHash) (*chainscript.Segment, error) {
	a.MockGetSegment.CalledCount++
	a.MockGetSegment.CalledWith = append(a.MockGetSegment.CalledWith, linkHash)
	a.MockGetSegment.LastCalledWith = linkHash

	if a.MockGetSegment.Fn != nil {
		return a.MockGetSegment.Fn(linkHash)
	}

	return nil, nil
}

// GetEvidences implements github.com/stratumn/go-core/store.Adapter.GetEvidences.
func (a *MockAdapter) GetEvidences(ctx context.Context, linkHash chainscript.LinkHash) (types.EvidenceSlice, error) {
	a.MockGetEvidences.CalledCount++
	a.MockGetEvidences.CalledWith = append(a.MockGetEvidences.CalledWith, linkHash)
	a.MockGetEvidences.LastCalledWith = linkHash

	if a.MockGetEvidences.Fn != nil {
		return a.MockGetEvidences.Fn(linkHash)
	}

	return nil, nil
}

// FindSegments implements github.com/stratumn/go-core/store.Adapter.FindSegments.
func (a *MockAdapter) FindSegments(ctx context.Context, filter *store.SegmentFilter) (*types.PaginatedSegments, error) {
	a.MockFindSegments.CalledCount++
	a.MockFindSegments.CalledWith = append(a.MockFindSegments.CalledWith, filter)
	a.MockFindSegments.LastCalledWith = filter

	if a.MockFindSegments.Fn != nil {
		return a.MockFindSegments.Fn(filter)
	}

	return &types.PaginatedSegments{}, nil
}

// GetMapIDs implements github.com/stratumn/go-core/store.Adapter.GetMapIDs.
func (a *MockAdapter) GetMapIDs(ctx context.Context, filter *store.MapFilter) ([]string, error) {
	a.MockGetMapIDs.CalledCount++
	a.MockGetMapIDs.CalledWith = append(a.MockGetMapIDs.CalledWith, filter)
	a.MockGetMapIDs.LastCalledWith = filter

	if a.MockGetMapIDs.Fn != nil {
		return a.MockGetMapIDs.Fn(filter)
	}

	return nil, nil
}

// NewBatch implements github.com/stratumn/go-core/store.Adapter.NewBatch.
func (a *MockAdapter) NewBatch(ctx context.Context) (store.Batch, error) {
	a.MockNewBatch.CalledCount++

	if a.MockNewBatch.Fn != nil {
		return a.MockNewBatch.Fn(), nil
	}

	return &MockBatch{}, nil
}

// SetValue implements github.com/stratumn/go-core/store.KeyValueStore.SetValue.
func (a *MockKeyValueStore) SetValue(ctx context.Context, key, value []byte) error {
	a.MockSetValue.CalledCount++
	calledWith := [][]byte{key, value}
	a.MockSetValue.CalledWith = append(a.MockSetValue.CalledWith, calledWith)
	a.MockSetValue.LastCalledWith = calledWith

	if a.MockSetValue.Fn != nil {
		return a.MockSetValue.Fn(key, value)
	}

	return nil
}

// GetValue implements github.com/stratumn/go-core/store.KeyValueStore.GetValue.
func (a *MockKeyValueStore) GetValue(ctx context.Context, key []byte) ([]byte, error) {
	a.MockGetValue.CalledCount++
	a.MockGetValue.CalledWith = append(a.MockGetValue.CalledWith, key)
	a.MockGetValue.LastCalledWith = key

	if a.MockGetValue.Fn != nil {
		return a.MockGetValue.Fn(key)
	}

	return nil, nil
}

// DeleteValue implements github.com/stratumn/go-core/store.KeyValueStore.DeleteValue.
func (a *MockKeyValueStore) DeleteValue(ctx context.Context, key []byte) ([]byte, error) {
	a.MockDeleteValue.CalledCount++
	a.MockDeleteValue.CalledWith = append(a.MockDeleteValue.CalledWith, key)
	a.MockDeleteValue.LastCalledWith = key

	if a.MockDeleteValue.Fn != nil {
		return a.MockDeleteValue.Fn(key)
	}

	return nil, nil
}
