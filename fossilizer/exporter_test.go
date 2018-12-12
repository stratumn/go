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

package fossilizer_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stratumn/go-core/dummyfossilizer"
	"github.com/stratumn/go-core/fossilizer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestEventExporter struct {
	eventsLock sync.RWMutex
	Events     []*fossilizer.Event
}

func (e *TestEventExporter) Push(_ context.Context, event *fossilizer.Event) error {
	e.eventsLock.Lock()
	defer e.eventsLock.Unlock()

	e.Events = append(e.Events, event)
	return nil
}

func (e *TestEventExporter) EventsCount() int {
	e.eventsLock.RLock()
	defer e.eventsLock.RUnlock()

	return len(e.Events)
}

func TestRunExporter(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	exporter := &TestEventExporter{}
	endChan := make(chan struct{})
	f := dummyfossilizer.New(&dummyfossilizer.Config{})
	go func() {
		fossilizer.RunExporter(ctx, f, exporter)
		endChan <- struct{}{}
	}()

	waitFor(func() bool { return f.ListenersCount() > 0 })

	err := f.Fossilize(ctx, []byte{42}, []byte{24})
	require.NoError(t, err)

	waitFor(func() bool { return exporter.EventsCount() > 0 })

	require.Len(t, exporter.Events, 1)
	event := exporter.Events[0]
	assert.Equal(t, fossilizer.DidFossilize, event.EventType)

	cancel()
	<-endChan
}

func waitFor(predicate func() bool) {
	for {
		if predicate() {
			break
		}

		<-time.After(2 * time.Millisecond)
	}
}
