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

package dummyqueue

import (
	"context"
	"sync"

	"github.com/stratumn/go-core/fossilizer"
)

// DummyQueue is a basic in-memory queue.
type DummyQueue struct {
	fossilsLock sync.RWMutex
	fossils     []*fossilizer.Fossil
}

// New creates a new dummy queue.
func New() *DummyQueue {
	return &DummyQueue{}
}

// Push a fossil to the queue.
func (q *DummyQueue) Push(_ context.Context, f *fossilizer.Fossil) error {
	q.fossilsLock.Lock()
	defer q.fossilsLock.Unlock()

	q.fossils = append(q.fossils, f)
	return nil
}

// Pop fossils from the queue.
func (q *DummyQueue) Pop(_ context.Context, count int) ([]*fossilizer.Fossil, error) {
	q.fossilsLock.Lock()
	defer q.fossilsLock.Unlock()

	var fossils []*fossilizer.Fossil
	for i := 0; i < count; i++ {
		if len(q.fossils) == 0 {
			break
		}

		fossils = append(fossils, q.fossils[0])
		q.fossils = q.fossils[1:]
	}

	return fossils, nil
}

// Fossils returns all the fossils in the queue for testing purposes.
func (q *DummyQueue) Fossils() []*fossilizer.Fossil {
	q.fossilsLock.RLock()
	defer q.fossilsLock.RUnlock()

	fossils := make([]*fossilizer.Fossil, len(q.fossils))
	for i, f := range q.fossils {
		fossils[i] = f
	}

	return fossils
}
