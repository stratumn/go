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

package dummyexporter

import (
	"context"
	"encoding/json"

	"github.com/stratumn/go-core/fossilizer"
	"github.com/stratumn/go-core/monitoring"
)

// DummyEventExporter exports fossilizer events to the console.
type DummyEventExporter struct{}

// New creates a new exporter that prints events to the console.
func New() fossilizer.EventExporter {
	return &DummyEventExporter{}
}

// Push prints the event to the console.
func (e *DummyEventExporter) Push(ctx context.Context, event *fossilizer.Event) error {
	b, err := json.Marshal(event)
	if err != nil {
		return err
	}

	monitoring.TxLogEntry(ctx).WithField("event", string(b)).Info(event.EventType)
	return nil
}
