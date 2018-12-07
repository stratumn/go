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

// Package batchfossilizer implements a fossilizer that fossilizes batches of
// data using a merkle tree.
// The evidence will contain the merkle root, the merkle path, and a timestamp.
package batchfossilizer

import (
	"context"
	"sync"

	"github.com/stratumn/go-core/fossilizer"
)

const (
	// Name is the name set in the fossilizer's information.
	Name = "batchfossilizer"

	// Description is the description set in the fossilizer's information.
	Description = "Stratumn Batch Fossilizer"
)

// Info is the info returned by GetInfo.
type Info struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Commit      string `json:"commit"`
}

// Fossilizer is the type that
// implements github.com/stratumn/go-core/fossilizer.Adapter.
type Fossilizer struct {
	config *Config

	fossilizerEventMutex sync.Mutex
	fossilizerEventChans []chan *fossilizer.Event
}

// New creates an instance of a Fossilizer.
func New(config *Config) fossilizer.Adapter {
	a := &Fossilizer{config: config}
	return a
}

// GetInfo implements github.com/stratumn/go-core/fossilizer.Adapter.GetInfo.
func (a *Fossilizer) GetInfo(ctx context.Context) (interface{}, error) {
	return &Info{
		Name:        Name,
		Description: Description,
		Version:     a.config.Version,
		Commit:      a.config.Commit,
	}, nil
}

// AddFossilizerEventChan implements
// github.com/stratumn/go-core/fossilizer.Adapter.AddFossilizerEventChan.
func (a *Fossilizer) AddFossilizerEventChan(fossilizerEventChan chan *fossilizer.Event) {
	a.fossilizerEventMutex.Lock()
	defer a.fossilizerEventMutex.Unlock()

	a.fossilizerEventChans = append(a.fossilizerEventChans, fossilizerEventChan)
}

// Fossilize implements github.com/stratumn/go-core/fossilizer.Adapter.Fossilize.
func (a *Fossilizer) Fossilize(ctx context.Context, data []byte, meta []byte) error {
	// TODO: just add to the queue.
	// TODO: go routine to dequeue every config.GetInterval() and actually
	// fossilize with the wrapped fossilizer.
	return nil
}
