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

package storetesting

import (
	"context"

	"github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-core/store"
	"github.com/stratumn/go-core/types"
)

// MockBatch is used to mock a batch.
// It implements github.com/stratumn/go-core/store.Batch
type MockBatch struct {
	// The mock for the CreateLink function.
	MockCreateLink MockBatchCreateLink

	// The mock for the Write function.
	MockWrite MockBatchWrite

	// The mock for the GetSegment function.
	MockGetSegment MockBatchGetSegment

	// The mock for the FindSegments function.
	MockFindSegments MockBatchFindSegments

	// The mock for the GetMapIDs function.
	MockGetMapIDs MockBatchGetMapIDs
}

// MockBatchCreateLink mocks the CreateLink function.
type MockBatchCreateLink struct {
	// The number of times the function was called.
	CalledCount int

	// The link that was passed to each call.
	CalledWith []*chainscript.Link

	// The last link that was passed.
	LastCalledWith *chainscript.Link

	// An optional implementation of the function.
	Fn func(*chainscript.Link) (chainscript.LinkHash, error)
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
	Fn func(linkHash chainscript.LinkHash) (*chainscript.Segment, error)
}

// MockBatchFindSegments mocks the FindSegments function.
type MockBatchFindSegments struct {
	// The number of times the function was called.
	CalledCount int

	// An optional implementation of the function.
	Fn func(filter *store.SegmentFilter) (*types.PaginatedSegments, error)
}

// MockBatchGetMapIDs mocks the GetMapIDs function.
type MockBatchGetMapIDs struct {
	// The number of times the function was called.
	CalledCount int

	// An optional implementation of the function.
	Fn func(filter *store.MapFilter) ([]string, error)
}

// CreateLink implements github.com/stratumn/go-core/store.Batch.CreateLink.
func (a *MockBatch) CreateLink(ctx context.Context, link *chainscript.Link) (chainscript.LinkHash, error) {
	a.MockCreateLink.CalledCount++
	a.MockCreateLink.CalledWith = append(a.MockCreateLink.CalledWith, link)
	a.MockCreateLink.LastCalledWith = link

	if a.MockCreateLink.Fn != nil {
		return a.MockCreateLink.Fn(link)
	}

	return nil, nil
}

// Write implements github.com/stratumn/go-core/store.Batch.Write.
func (a *MockBatch) Write(ctx context.Context) error {
	a.MockWrite.CalledCount++

	if a.MockWrite.Fn != nil {
		return a.MockWrite.Fn()
	}
	return nil
}

// GetSegment delegates the call to a underlying store
func (a *MockBatch) GetSegment(ctx context.Context, linkHash chainscript.LinkHash) (*chainscript.Segment, error) {
	a.MockGetSegment.CalledCount++

	if a.MockGetSegment.Fn != nil {
		return a.MockGetSegment.Fn(linkHash)
	}
	return nil, nil
}

// FindSegments delegates the call to a underlying store
func (a *MockBatch) FindSegments(ctx context.Context, filter *store.SegmentFilter) (*types.PaginatedSegments, error) {
	a.MockFindSegments.CalledCount++

	if a.MockFindSegments.Fn != nil {
		return a.MockFindSegments.Fn(filter)
	}
	return nil, nil
}

// GetMapIDs delegates the call to a underlying store
func (a *MockBatch) GetMapIDs(ctx context.Context, filter *store.MapFilter) ([]string, error) {
	a.MockGetMapIDs.CalledCount++

	if a.MockGetMapIDs.Fn != nil {
		return a.MockGetMapIDs.Fn(filter)
	}
	return nil, nil
}
