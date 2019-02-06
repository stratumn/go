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
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Default buckets used by histograms.
var (
	DefaultLatencyBuckets = []float64{0, 1, 2, 3, 4, 5, 6, 8, 10, 13, 16, 20, 25, 30, 40, 50, 65, 80, 100, 130, 160, 200, 250, 300, 400, 500, 650, 800, 1000, 2000, 5000, 10000, 20000, 50000, 100000}
)

// Metrics and labels available to all packages.
const (
	Stratumn            = "stratumn"
	ErrorLabel          = "error"
	ErrorCodeLabel      = "error_code"
	ErrorComponentLabel = "error_component"
)

// Private labels used only inside this package.
const (
	adapterRequest = "adapter_request"
)

// Store metrics used only inside this package.
var (
	storeRequestCount   *prometheus.CounterVec
	storeRequestErr     *prometheus.CounterVec
	storeRequestLatency *prometheus.HistogramVec
)

// Fossilizer metrics used only inside this package.
var (
	fossilizerRequestCount   *prometheus.CounterVec
	fossilizerRequestErr     *prometheus.CounterVec
	fossilizerRequestLatency *prometheus.HistogramVec
)

func init() {
	storeRequestCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: Stratumn,
			Subsystem: "store",
			Name:      "request_count",
			Help:      "number of requests to the store",
		},
		[]string{adapterRequest},
	)

	storeRequestLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: Stratumn,
			Subsystem: "store",
			Name:      "request_latency_ms",
			Help:      "latency of store requests",
			Buckets:   DefaultLatencyBuckets,
		},
		[]string{adapterRequest},
	)

	storeRequestErr = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: Stratumn,
			Subsystem: "store",
			Name:      "request_error",
			Help:      "number of fossilizer request errors",
		},
		[]string{adapterRequest, ErrorCodeLabel, ErrorComponentLabel},
	)

	fossilizerRequestCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: Stratumn,
			Subsystem: "fossilizer",
			Name:      "request_count",
			Help:      "number of requests to the fossilizer",
		},
		[]string{adapterRequest},
	)

	fossilizerRequestLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: Stratumn,
			Subsystem: "fossilizer",
			Name:      "request_latency_ms",
			Help:      "latency of fossilizer requests",
			Buckets:   DefaultLatencyBuckets,
		},
		[]string{adapterRequest},
	)

	fossilizerRequestErr = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: Stratumn,
			Subsystem: "fossilizer",
			Name:      "request_error",
			Help:      "number of fossilizer request errors",
		},
		[]string{adapterRequest, ErrorCodeLabel, ErrorComponentLabel},
	)
}
