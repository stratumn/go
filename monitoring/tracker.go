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
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stratumn/go-core/monitoring/errorcode"
	"github.com/stratumn/go-core/types"
)

type requestTracker struct {
	start          time.Time
	request        string
	requestCounter *prometheus.CounterVec
	errorCounter   *prometheus.CounterVec
	latencyCounter *prometheus.HistogramVec
}

func newStoreRequestTracker(request string) *requestTracker {
	return newRequestTracker(request, storeRequestCount, storeRequestErr, storeRequestLatency)
}

func newFossilizerRequestTracker(request string) *requestTracker {
	return newRequestTracker(request, fossilizerRequestCount, fossilizerRequestErr, fossilizerRequestLatency)
}

func newRequestTracker(
	request string,
	requestCounter *prometheus.CounterVec,
	errorCounter *prometheus.CounterVec,
	latencyCounter *prometheus.HistogramVec,
) *requestTracker {
	return &requestTracker{
		start:          time.Now(),
		request:        request,
		requestCounter: requestCounter,
		errorCounter:   errorCounter,
		latencyCounter: latencyCounter,
	}
}

func (t *requestTracker) End(err error) {
	if err != nil {
		e, ok := err.(*types.Error)
		if ok {
			t.errorCounter.With(prometheus.Labels{
				adapterRequest:      t.request,
				ErrorCodeLabel:      errorcode.Text(e.Code),
				ErrorComponentLabel: e.Component,
			}).Inc()
		} else {
			t.errorCounter.With(prometheus.Labels{
				adapterRequest:      t.request,
				ErrorCodeLabel:      errorcode.Text(errorcode.Unknown),
				ErrorComponentLabel: "Unknown",
			}).Inc()
		}
	}

	t.requestCounter.With(prometheus.Labels{adapterRequest: t.request}).Inc()
	t.latencyCounter.With(prometheus.Labels{adapterRequest: t.request}).Observe(float64(time.Since(t.start)) / float64(time.Millisecond))
}
