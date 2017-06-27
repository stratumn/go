// Copyright 2017 Stratumn SAS. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package storetesting

import (
	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/types"
)

// MockBatch is used to mock a batch.
//
// It implements github.com/stratumn/sdk/store.Batch.
type MockBatch struct {
	// The mock for the SaveSegment function.
	MockSaveSegment MockBatchSaveSegment

	// The mock for the SaveValue function.
	MockSaveValue MockBatchSaveValue

	// The mock for the DeleteSegment function.
	MockDeleteSegment MockBatchDeleteSegment

	// The mock for the DeleteValue function.
	MockDeleteValue MockBatchDeleteValue

	// The mock for the Write function.
	MockWrite MockBatchWrite

	// The mock for the GetSegment function.
	MockGetSegment MockBatchGetSegment

	// The mock for the FindSegments function.
	MockFindSegments MockBatchFindSegments

	// The mock for the GetMapIDs function.
	MockGetMapIDs MockBatchGetMapIDs

	// The mock for the GetValue function.
	MockGetValue MockBatchGetValue
}

// MockBatchSaveSegment mocks the SaveSegment function.
type MockBatchSaveSegment struct {
	// The number of times the function was called.
	CalledCount int

	// The segment that was passed to each call.
	CalledWith []*cs.Segment

	// The last segment that was passed.
	LastCalledWith *cs.Segment

	// An optional implementation of the function.
	Fn func(*cs.Segment) error
}

// MockBatchSaveValue mocks the SaveValue function.
type MockBatchSaveValue struct {
	// The number of times the function was called.
	CalledCount int

	// The segment that was passed to each call.
	CalledWith [][][]byte

	// The last segment that was passed.
	LastCalledWith [][]byte

	// An optional implementation of the function.
	Fn func(key, value []byte) error
}

// MockBatchDeleteSegment mocks the DeleteSegment function.
type MockBatchDeleteSegment struct {
	// The number of times the function was called.
	CalledCount int

	// The link hash that was passed to each call.
	CalledWith []*types.Bytes32

	// The last link hash that was passed.
	LastCalledWith *types.Bytes32

	// An optional implementation of the function.
	Fn func(*types.Bytes32) (*cs.Segment, error)
}

// MockBatchDeleteValue mocks the DeleteValue function.
type MockBatchDeleteValue struct {
	// The number of times the function was called.
	CalledCount int

	// The key that was passed to each call.
	CalledWith [][]byte

	// The last link hash that was passed.
	LastCalledWith []byte

	// An optional implementation of the function.
	Fn func([]byte) ([]byte, error)
}

// MockBatchWrite mocks the Write function.
type MockBatchWrite struct {
	// The number of times the function was called.
	CalledCount int

	// An optional implementation of the function.
	Fn func() error
}

// MockBatchGetSegment mocks the GetSegment function.
type MockBatchGetSegment struct {
	// The number of times the function was called.
	CalledCount int

	// An optional implementation of the function.
	Fn func(linkHash *types.Bytes32) (*cs.Segment, error)
}

// MockBatchFindSegments mocks the FindSegments function.
type MockBatchFindSegments struct {
	// The number of times the function was called.
	CalledCount int

	// An optional implementation of the function.
	Fn func(filter *store.Filter) (cs.SegmentSlice, error)
}

// MockBatchGetMapIDs mocks the GetMapIDs function.
type MockBatchGetMapIDs struct {
	// The number of times the function was called.
	CalledCount int

	// An optional implementation of the function.
	Fn func(pagination *store.Pagination) ([]string, error)
}

// MockBatchGetValue mocks the GetValue function.
type MockBatchGetValue struct {
	// The number of times the function was called.
	CalledCount int

	// An optional implementation of the function.
	Fn func(key []byte) ([]byte, error)
}

// SaveSegment implements github.com/stratumn/sdk/store.Batch.SaveSegment.
func (a *MockBatch) SaveSegment(segment *cs.Segment) error {
	a.MockSaveSegment.CalledCount++
	a.MockSaveSegment.CalledWith = append(a.MockSaveSegment.CalledWith, segment)
	a.MockSaveSegment.LastCalledWith = segment

	if a.MockSaveSegment.Fn != nil {
		return a.MockSaveSegment.Fn(segment)
	}

	return nil
}

// SaveValue implements github.com/stratumn/sdk/store.Batch.SaveValue.
func (a *MockBatch) SaveValue(key, value []byte) error {
	a.MockSaveValue.CalledCount++
	calledWith := [][]byte{key, value}
	a.MockSaveValue.CalledWith = append(a.MockSaveValue.CalledWith, calledWith)
	a.MockSaveValue.LastCalledWith = calledWith

	if a.MockSaveValue.Fn != nil {
		return a.MockSaveValue.Fn(key, value)
	}

	return nil
}

// DeleteSegment implements github.com/stratumn/sdk/store.Batch.DeleteSegment.
func (a *MockBatch) DeleteSegment(linkHash *types.Bytes32) (*cs.Segment, error) {
	a.MockDeleteSegment.CalledCount++
	a.MockDeleteSegment.CalledWith = append(a.MockDeleteSegment.CalledWith, linkHash)
	a.MockDeleteSegment.LastCalledWith = linkHash

	if a.MockDeleteSegment.Fn != nil {
		return a.MockDeleteSegment.Fn(linkHash)
	}

	return nil, nil
}

// DeleteValue implements github.com/stratumn/sdk/store.Batch.DeleteValue.
func (a *MockBatch) DeleteValue(key []byte) ([]byte, error) {
	a.MockDeleteValue.CalledCount++
	a.MockDeleteValue.CalledWith = append(a.MockDeleteValue.CalledWith, key)
	a.MockDeleteValue.LastCalledWith = key

	if a.MockDeleteValue.Fn != nil {
		return a.MockDeleteValue.Fn(key)
	}

	return nil, nil
}

// Write implements github.com/stratumn/sdk/store.Batch.Write.
func (a *MockBatch) Write() error {
	a.MockWrite.CalledCount++

	if a.MockWrite.Fn != nil {
		return a.MockWrite.Fn()
	}
	return nil
}

// GetSegment delegates the call to a underlying store
func (a *MockBatch) GetSegment(linkHash *types.Bytes32) (*cs.Segment, error) {
	a.MockGetSegment.CalledCount++

	if a.MockGetSegment.Fn != nil {
		return a.MockGetSegment.Fn(linkHash)
	}
	return nil, nil
}

// FindSegments delegates the call to a underlying store
func (a *MockBatch) FindSegments(filter *store.Filter) (cs.SegmentSlice, error) {
	a.MockFindSegments.CalledCount++

	if a.MockFindSegments.Fn != nil {
		return a.MockFindSegments.Fn(filter)
	}
	return nil, nil
}

// GetMapIDs delegates the call to a underlying store
func (a *MockBatch) GetMapIDs(pagination *store.Pagination) ([]string, error) {
	a.MockGetMapIDs.CalledCount++

	if a.MockGetMapIDs.Fn != nil {
		return a.MockGetMapIDs.Fn(pagination)
	}
	return nil, nil
}

// GetValue delegates the call to a underlying store
func (a *MockBatch) GetValue(key []byte) ([]byte, error) {
	a.MockGetValue.CalledCount++

	if a.MockGetValue.Fn != nil {
		return a.MockGetValue.Fn(key)
	}
	return nil, nil
}
