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

package monitoring

import (
	"context"
	"fmt"

	"github.com/stratumn/go-core/fossilizer"
)

// FossilizerAdapter is a decorator for the Fossilizer interface.
// It wraps a real Fossilizer implementation and adds instrumentation.
type FossilizerAdapter struct {
	f    fossilizer.Adapter
	name string
}

// NewFossilizerAdapter decorates an existing fossilizer.
func NewFossilizerAdapter(f fossilizer.Adapter, name string) fossilizer.Adapter {
	return &FossilizerAdapter{f: f, name: name}
}

// GetInfo instruments the call and delegates to the underlying fossilizer.
func (a *FossilizerAdapter) GetInfo(ctx context.Context) (res interface{}, err error) {
	span, ctx := StartSpanIncomingRequest(ctx, fmt.Sprintf("%s/GetInfo", a.name))
	defer func() {
		SetSpanStatusAndEnd(span, err)
	}()

	res, err = a.f.GetInfo(ctx)
	return
}

// AddFossilizerEventChan instruments the call and delegates to the underlying fossilizer.
func (a *FossilizerAdapter) AddFossilizerEventChan(c chan *fossilizer.Event) {
	span, _ := StartSpanIncomingRequest(context.Background(), fmt.Sprintf("%s/AddFossilizerEventChan", a.name))
	defer span.End()

	a.f.AddFossilizerEventChan(c)
}

// Fossilize instruments the call and delegates to the underlying fossilizer.
func (a *FossilizerAdapter) Fossilize(ctx context.Context, data []byte, meta []byte) (err error) {
	tracker := newFossilizerRequestTracker("Fossilize")
	span, ctx := StartSpanIncomingRequest(ctx, fmt.Sprintf("%s/Fossilize", a.name))
	defer func() {
		SetSpanStatusAndEnd(span, err)
		tracker.End(err)
	}()

	err = a.f.Fossilize(ctx, data, meta)
	return
}
