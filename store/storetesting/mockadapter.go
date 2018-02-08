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
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/types"
)

// MockAdapter is used to mock a store.
// It implements github.com/stratumn/go-indigocore/store.Adapter.
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
// It implements github.com/stratumn/go-indigocore/store.KeyValueStore.
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
	CalledWith []*cs.Link

	// The last link that was passed.
	LastCalledWith *cs.Link

	// An optional implementation of the function.
	Fn func(*cs.Link) (*types.Bytes32, error)
}

// MockAddEvidence mocks the AddEvidence function.
type MockAddEvidence struct {
	// The number of times the function was called.
	CalledCount int

	// The evidence that was passed to each call.
	CalledWith []*cs.Evidence

	// The last evidence that was passed.
	LastCalledWith *cs.Evidence

	// An optional implementation of the function.
	Fn func(linkHash *types.Bytes32, evidence *cs.Evidence) error
}

// MockGetSegment mocks the GetSegment function.
type MockGetSegment struct {
	// The number of times the function was called.
	CalledCount int

	// The link hash that was passed to each call.
	CalledWith []*types.Bytes32

	// The last link hash that was passed.
	LastCalledWith *types.Bytes32

	// An optional implementation of the function.
	Fn func(*types.Bytes32) (*cs.Segment, error)
}

// MockGetEvidences mocks the GetEvidences function.
type MockGetEvidences struct {
	// The number of times the function was called.
	CalledCount int

	// The link hash that was passed to each call.
	CalledWith []*types.Bytes32

	// The last link hash that was passed.
	LastCalledWith *types.Bytes32

	// An optional implementation of the function.
	Fn func(*types.Bytes32) (*cs.Evidences, error)
}

// MockDeleteSegment mocks the DeleteSegment function.
type MockDeleteSegment struct {
	// The number of times the function was called.
	CalledCount int

	// The link hash that was passed to each call.
	CalledWith []*types.Bytes32

	// The last link hash that was passed.
	LastCalledWith *types.Bytes32

	// An optional implementation of the function.
	Fn func(*types.Bytes32) (*cs.Segment, error)
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
	Fn func(*store.SegmentFilter) (cs.SegmentSlice, error)
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

// GetInfo implements github.com/stratumn/go-indigocore/store.Adapter.GetInfo.
func (a *MockAdapter) GetInfo() (interface{}, error) {
	a.MockGetInfo.CalledCount++

	if a.MockGetInfo.Fn != nil {
		return a.MockGetInfo.Fn()
	}

	return nil, nil
}

// AddStoreEventChannel implements
// github.com/stratumn/go-indigocore/store.Adapter.AddStoreEventChannel.
func (a *MockAdapter) AddStoreEventChannel(storeChan chan *store.Event) {
	a.MockAddStoreEventChannel.CalledCount++
	a.MockAddStoreEventChannel.CalledWith = append(a.MockAddStoreEventChannel.CalledWith, storeChan)
	a.MockAddStoreEventChannel.LastCalledWith = storeChan

	if a.MockAddStoreEventChannel.Fn != nil {
		a.MockAddStoreEventChannel.Fn(storeChan)
	}
}

// CreateLink implements github.com/stratumn/go-indigocore/store.Adapter.CreateLink.
func (a *MockAdapter) CreateLink(link *cs.Link) (*types.Bytes32, error) {
	a.MockCreateLink.CalledCount++
	a.MockCreateLink.CalledWith = append(a.MockCreateLink.CalledWith, link)
	a.MockCreateLink.LastCalledWith = link

	if a.MockCreateLink.Fn != nil {
		return a.MockCreateLink.Fn(link)
	}

	return nil, nil
}

// AddEvidence implements github.com/stratumn/go-indigocore/store.Adapter.AddEvidence.
func (a *MockAdapter) AddEvidence(linkHash *types.Bytes32, evidence *cs.Evidence) error {
	a.MockAddEvidence.CalledCount++
	a.MockAddEvidence.CalledWith = append(a.MockAddEvidence.CalledWith, evidence)
	a.MockAddEvidence.LastCalledWith = evidence

	if a.MockAddEvidence.Fn != nil {
		return a.MockAddEvidence.Fn(linkHash, evidence)
	}

	return nil
}

// GetSegment implements github.com/stratumn/go-indigocore/store.Adapter.GetSegment.
func (a *MockAdapter) GetSegment(linkHash *types.Bytes32) (*cs.Segment, error) {
	a.MockGetSegment.CalledCount++
	a.MockGetSegment.CalledWith = append(a.MockGetSegment.CalledWith, linkHash)
	a.MockGetSegment.LastCalledWith = linkHash

	if a.MockGetSegment.Fn != nil {
		return a.MockGetSegment.Fn(linkHash)
	}

	return nil, nil
}

// GetEvidences implements github.com/stratumn/go-indigocore/store.Adapter.GetEvidences.
func (a *MockAdapter) GetEvidences(linkHash *types.Bytes32) (*cs.Evidences, error) {
	a.MockGetEvidences.CalledCount++
	a.MockGetEvidences.CalledWith = append(a.MockGetEvidences.CalledWith, linkHash)
	a.MockGetEvidences.LastCalledWith = linkHash

	if a.MockGetEvidences.Fn != nil {
		return a.MockGetEvidences.Fn(linkHash)
	}

	return nil, nil
}

// FindSegments implements github.com/stratumn/go-indigocore/store.Adapter.FindSegments.
func (a *MockAdapter) FindSegments(filter *store.SegmentFilter) (cs.SegmentSlice, error) {
	a.MockFindSegments.CalledCount++
	a.MockFindSegments.CalledWith = append(a.MockFindSegments.CalledWith, filter)
	a.MockFindSegments.LastCalledWith = filter

	if a.MockFindSegments.Fn != nil {
		return a.MockFindSegments.Fn(filter)
	}

	return nil, nil
}

// GetMapIDs implements github.com/stratumn/go-indigocore/store.Adapter.GetMapIDs.
func (a *MockAdapter) GetMapIDs(filter *store.MapFilter) ([]string, error) {
	a.MockGetMapIDs.CalledCount++
	a.MockGetMapIDs.CalledWith = append(a.MockGetMapIDs.CalledWith, filter)
	a.MockGetMapIDs.LastCalledWith = filter

	if a.MockGetMapIDs.Fn != nil {
		return a.MockGetMapIDs.Fn(filter)
	}

	return nil, nil
}

// NewBatch implements github.com/stratumn/go-indigocore/store.Adapter.NewBatch.
func (a *MockAdapter) NewBatch() (store.Batch, error) {
	a.MockNewBatch.CalledCount++

	if a.MockNewBatch.Fn != nil {
		return a.MockNewBatch.Fn(), nil
	}

	return &MockBatch{}, nil
}

// SetValue implements github.com/stratumn/go-indigocore/store.KeyValueStore.SetValue.
func (a *MockKeyValueStore) SetValue(key, value []byte) error {
	a.MockSetValue.CalledCount++
	calledWith := [][]byte{key, value}
	a.MockSetValue.CalledWith = append(a.MockSetValue.CalledWith, calledWith)
	a.MockSetValue.LastCalledWith = calledWith

	if a.MockSetValue.Fn != nil {
		return a.MockSetValue.Fn(key, value)
	}

	return nil
}

// GetValue implements github.com/stratumn/go-indigocore/store.KeyValueStore.GetValue.
func (a *MockKeyValueStore) GetValue(key []byte) ([]byte, error) {
	a.MockGetValue.CalledCount++
	a.MockGetValue.CalledWith = append(a.MockGetValue.CalledWith, key)
	a.MockGetValue.LastCalledWith = key

	if a.MockGetValue.Fn != nil {
		return a.MockGetValue.Fn(key)
	}

	return nil, nil
}

// DeleteValue implements github.com/stratumn/go-indigocore/store.KeyValueStore.DeleteValue.
func (a *MockKeyValueStore) DeleteValue(key []byte) ([]byte, error) {
	a.MockDeleteValue.CalledCount++
	a.MockDeleteValue.CalledWith = append(a.MockDeleteValue.CalledWith, key)
	a.MockDeleteValue.LastCalledWith = key

	if a.MockDeleteValue.Fn != nil {
		return a.MockDeleteValue.Fn(key)
	}

	return nil, nil
}
