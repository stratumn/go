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

package monitoring

import (
	"context"
	"fmt"

	"github.com/stratumn/go-indigocore/fossilizer"

	"go.opencensus.io/trace"
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
	ctx, span := trace.StartSpan(ctx, fmt.Sprintf("%s/GetInfo", a.name))
	defer SetSpanStatusAndEnd(span, err)

	res, err = a.f.GetInfo(ctx)
	return
}

// AddFossilizerEventChan instruments the call and delegates to the underlying fossilizer.
func (a *FossilizerAdapter) AddFossilizerEventChan(c chan *fossilizer.Event) {
	_, span := trace.StartSpan(context.Background(), fmt.Sprintf("%s/AddFossilizerEventChan", a.name))
	defer span.End()

	a.AddFossilizerEventChan(c)
}

// Fossilize instruments the call and delegates to the underlying fossilizer.
func (a *FossilizerAdapter) Fossilize(ctx context.Context, data []byte, meta []byte) (err error) {
	ctx, span := trace.StartSpan(ctx, fmt.Sprintf("%s/Fossilize", a.name))
	defer SetSpanStatusAndEnd(span, err)

	err = a.f.Fossilize(ctx, data, meta)
	return
}
