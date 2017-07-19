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
	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/types"
)

// MockAdapter is used to mock a store.
//
// It implements github.com/stratumn/sdk/store.Adapter.
type MockAdapter struct {
	// The mock for the GetInfo function.
	MockGetInfo MockGetInfo

	// The mock for the AddDidSaveChannel function.
	MockAddDidSaveChannel MockAddDidSaveChannel

	// The mock for the SaveSegment function.
	MockSaveSegment MockSaveSegment

	// The mock for the SaveValue function.
	MockSaveValue MockSaveValue

	// The mock for the GetSegment function.
	MockGetSegment MockGetSegment

	// The mock for the GetValue function.
	MockGetValue MockGetValue

	// The mock for the DeleteSegment function.
	MockDeleteSegment MockDeleteSegment

	// The mock for the DeleteValue function.
	MockDeleteValue MockDeleteValue

	// The mock for the FindSegments function.
	MockFindSegments MockFindSegments

	// The mock for the GetMapIDs function.
	MockGetMapIDs MockGetMapIDs

	// The mock for the NewBatch function.
	MockNewBatch MockNewBatch
}

// MockGetInfo mocks the GetInfo function.
type MockGetInfo struct {
	// The number of times the function was called.
	CalledCount int

	// An optional implementation of the function.
	Fn func() (interface{}, error)
}

// MockAddDidSaveChannel mocks the SaveSegment function.
type MockAddDidSaveChannel struct {
	// The number of times the function was called.
	CalledCount int

	// The segment that was passed to each call.
	CalledWith []chan *cs.Segment

	// The last segment that was passed.
	LastCalledWith chan *cs.Segment

	// An optional implementation of the function.
	Fn func(chan *cs.Segment)
}

// MockSaveSegment mocks the SaveSegment function.
type MockSaveSegment struct {
	// The number of times the function was called.
	CalledCount int

	// The segment that was passed to each call.
	CalledWith []*cs.Segment

	// The last segment that was passed.
	LastCalledWith *cs.Segment

	// An optional implementation of the function.
	Fn func(*cs.Segment) error
}

// MockSaveValue mocks the SaveValue function.
type MockSaveValue struct {
	// The number of times the function was called.
	CalledCount int

	// The segment that was passed to each call.
	CalledWith [][][]byte

	// The last segment that was passed.
	LastCalledWith [][]byte

	// An optional implementation of the function.
	Fn func(key, value []byte) error
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

// MockFindSegments mocks the FindSegments function.
type MockFindSegments struct {
	// The number of times the function was called.
	CalledCount int

	// The filter that was passed to each call.
	CalledWith []*store.Filter

	// The last filter that was passed.
	LastCalledWith *store.Filter

	// An optional implementation of the function.
	Fn func(*store.Filter) (cs.SegmentSlice, error)
}

// MockGetMapIDs mocks the GetMapIDs function.
type MockGetMapIDs struct {
	// The number of times the function was called.
	CalledCount int

	// The pagination that was passed to each call.
	CalledWith []*store.Pagination

	// The last pagination that was passed.
	LastCalledWith *store.Pagination

	// An optional implementation of the function.
	Fn func(*store.Pagination) ([]string, error)
}

// MockNewBatch mocks the NewBatch function.
type MockNewBatch struct {
	// The number of times the function was called.
	CalledCount int

	// An optional implementation of the function.
	Fn func() store.Batch
}

// GetInfo implements github.com/stratumn/sdk/store.Adapter.GetInfo.
func (a *MockAdapter) GetInfo() (interface{}, error) {
	a.MockGetInfo.CalledCount++

	if a.MockGetInfo.Fn != nil {
		return a.MockGetInfo.Fn()
	}

	return nil, nil
}

// AddDidSaveChannel implements
// github.com/stratumn/sdk/store.Adapter.AddDidSaveChannel.
func (a *MockAdapter) AddDidSaveChannel(saveChan chan *cs.Segment) {
	a.MockAddDidSaveChannel.CalledCount++
	a.MockAddDidSaveChannel.CalledWith = append(a.MockAddDidSaveChannel.CalledWith, saveChan)
	a.MockAddDidSaveChannel.LastCalledWith = saveChan

	if a.MockAddDidSaveChannel.Fn != nil {
		a.MockAddDidSaveChannel.Fn(saveChan)
	}
}

// SaveSegment implements github.com/stratumn/sdk/store.Adapter.SaveSegment.
func (a *MockAdapter) SaveSegment(segment *cs.Segment) error {
	a.MockSaveSegment.CalledCount++
	a.MockSaveSegment.CalledWith = append(a.MockSaveSegment.CalledWith, segment)
	a.MockSaveSegment.LastCalledWith = segment

	if a.MockSaveSegment.Fn != nil {
		return a.MockSaveSegment.Fn(segment)
	}

	return nil
}

// SaveValue implements github.com/stratumn/sdk/store.Adapter.SaveValue.
func (a *MockAdapter) SaveValue(key, value []byte) error {
	a.MockSaveValue.CalledCount++
	calledWith := [][]byte{key, value}
	a.MockSaveValue.CalledWith = append(a.MockSaveValue.CalledWith, calledWith)
	a.MockSaveValue.LastCalledWith = calledWith

	if a.MockSaveValue.Fn != nil {
		return a.MockSaveValue.Fn(key, value)
	}

	return nil
}

// GetSegment implements github.com/stratumn/sdk/store.Adapter.GetSegment.
func (a *MockAdapter) GetSegment(linkHash *types.Bytes32) (*cs.Segment, error) {
	a.MockGetSegment.CalledCount++
	a.MockGetSegment.CalledWith = append(a.MockGetSegment.CalledWith, linkHash)
	a.MockGetSegment.LastCalledWith = linkHash

	if a.MockGetSegment.Fn != nil {
		return a.MockGetSegment.Fn(linkHash)
	}

	return nil, nil
}

// GetValue implements github.com/stratumn/sdk/store.Adapter.GetValue.
func (a *MockAdapter) GetValue(key []byte) ([]byte, error) {
	a.MockGetValue.CalledCount++
	a.MockGetValue.CalledWith = append(a.MockGetValue.CalledWith, key)
	a.MockGetValue.LastCalledWith = key

	if a.MockGetValue.Fn != nil {
		return a.MockGetValue.Fn(key)
	}

	return nil, nil
}

// DeleteSegment implements github.com/stratumn/sdk/store.Adapter.DeleteSegment.
func (a *MockAdapter) DeleteSegment(linkHash *types.Bytes32) (*cs.Segment, error) {
	a.MockDeleteSegment.CalledCount++
	a.MockDeleteSegment.CalledWith = append(a.MockDeleteSegment.CalledWith, linkHash)
	a.MockDeleteSegment.LastCalledWith = linkHash

	if a.MockDeleteSegment.Fn != nil {
		return a.MockDeleteSegment.Fn(linkHash)
	}

	return nil, nil
}

// DeleteValue implements github.com/stratumn/sdk/store.Adapter.DeleteValue.
func (a *MockAdapter) DeleteValue(key []byte) ([]byte, error) {
	a.MockDeleteValue.CalledCount++
	a.MockDeleteValue.CalledWith = append(a.MockDeleteValue.CalledWith, key)
	a.MockDeleteValue.LastCalledWith = key

	if a.MockDeleteValue.Fn != nil {
		return a.MockDeleteValue.Fn(key)
	}

	return nil, nil
}

// FindSegments implements github.com/stratumn/sdk/store.Adapter.FindSegments.
func (a *MockAdapter) FindSegments(filter *store.Filter) (cs.SegmentSlice, error) {
	a.MockFindSegments.CalledCount++
	a.MockFindSegments.CalledWith = append(a.MockFindSegments.CalledWith, filter)
	a.MockFindSegments.LastCalledWith = filter

	if a.MockFindSegments.Fn != nil {
		return a.MockFindSegments.Fn(filter)
	}

	return nil, nil
}

// GetMapIDs implements github.com/stratumn/sdk/store.Adapter.GetMapIDs.
func (a *MockAdapter) GetMapIDs(pagination *store.Pagination) ([]string, error) {
	a.MockGetMapIDs.CalledCount++
	a.MockGetMapIDs.CalledWith = append(a.MockGetMapIDs.CalledWith, pagination)
	a.MockGetMapIDs.LastCalledWith = pagination

	if a.MockGetMapIDs.Fn != nil {
		return a.MockGetMapIDs.Fn(pagination)
	}

	return nil, nil
}

// NewBatch implements github.com/stratumn/sdk/store.Adapter.NewBatch.
func (a *MockAdapter) NewBatch() (store.Batch, error) {
	a.MockNewBatch.CalledCount++

	if a.MockGetMapIDs.Fn != nil {
		return a.MockNewBatch.Fn(), nil
	}

	return &MockBatch{}, nil
}
