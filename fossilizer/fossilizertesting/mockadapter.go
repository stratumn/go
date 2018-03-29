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

package fossilizertesting

import (
	"context"

	"github.com/stratumn/go-indigocore/fossilizer"
)

// MockAdapter is used to mock a fossilizer.
//
// It implements github.com/stratumn/go-indigocore/fossilizer.Adapter.
type MockAdapter struct {
	// The mock for the GetInfo function.
	MockGetInfo MockGetInfo

	// The mock for the AddFossilizerEventChan function.
	MockAddFossilizerEventChan MockAddFossilizerEventChan

	// The mock for the Fossilize function.
	MockFossilize MockFossilize
}

// MockGetInfo mocks the GetInfo function.
type MockGetInfo struct {
	// The number of times the function was called.
	CalledCount int

	// An optional implementation of the function.
	Fn func() (interface{}, error)
}

// MockAddFossilizerEventChan mocks the AddFossilizerEventChan function.
type MockAddFossilizerEventChan struct {
	// The number of times the function was called.
	CalledCount int

	// The channel that was passed to each call.
	CalledWith []chan *fossilizer.Event

	// The last channel that was passed.
	LastCalledWith chan *fossilizer.Event

	// An optional implementation of the function.
	Fn func(chan *fossilizer.Event)
}

// MockFossilize mocks the Fossilize function.
type MockFossilize struct {
	// The number of times the function was called.
	CalledCount int

	// The data that was passed to each call.
	CalledWithData [][]byte

	// The meta that was passed to each call.
	CalledWithMeta [][]byte

	// The last data that was passed.
	LastCalledWithData []byte

	// The last meta that was passed.
	LastCalledWithMeta []byte

	// An optional implementation of the function.
	Fn func([]byte, []byte) error
}

// GetInfo implements github.com/stratumn/go-indigocore/fossilizer.Adapter.GetInfo.
func (a *MockAdapter) GetInfo(_ context.Context) (interface{}, error) {
	a.MockGetInfo.CalledCount++

	if a.MockGetInfo.Fn != nil {
		return a.MockGetInfo.Fn()
	}

	return nil, nil
}

// AddFossilizerEventChan implements
// github.com/stratumn/go-indigocore/fossilizer.Adapter.AddFossilizerEventChan.
func (a *MockAdapter) AddFossilizerEventChan(eventChan chan *fossilizer.Event) {
	a.MockAddFossilizerEventChan.CalledCount++
	a.MockAddFossilizerEventChan.CalledWith = append(a.MockAddFossilizerEventChan.CalledWith, eventChan)
	a.MockAddFossilizerEventChan.LastCalledWith = eventChan

	if a.MockAddFossilizerEventChan.Fn != nil {
		a.MockAddFossilizerEventChan.Fn(eventChan)
	}
}

// Fossilize implements github.com/stratumn/go-indigocore/fossilizer.Adapter.Fossilize.
func (a *MockAdapter) Fossilize(_ context.Context, data []byte, meta []byte) error {
	a.MockFossilize.CalledCount++
	a.MockFossilize.CalledWithData = append(a.MockFossilize.CalledWithData, data)
	a.MockFossilize.LastCalledWithData = data
	a.MockFossilize.CalledWithMeta = append(a.MockFossilize.CalledWithMeta, meta)
	a.MockFossilize.LastCalledWithMeta = meta

	if a.MockFossilize.Fn != nil {
		return a.MockFossilize.Fn(data, meta)
	}

	return nil
}
