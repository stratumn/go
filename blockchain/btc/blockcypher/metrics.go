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

package blockcypher

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/stratumn/go-core/monitoring"
)

// Private labels used only inside this package.
const (
	requestType    = "request_type"
	accountAddress = "address"
)

// Private metrics used only inside this package.
var (
	requestCount   *prometheus.CounterVec
	requestErr     *prometheus.CounterVec
	requestLatency *prometheus.HistogramVec

	accountBalance *prometheus.GaugeVec
)

func init() {
	requestCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "stratumn",
			Subsystem: "btc",
			Name:      "request_count",
			Help:      "number of requests to the bitcoin blockchain",
		},
		[]string{requestType},
	)

	requestErr = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "stratumn",
			Subsystem: "btc",
			Name:      "request_error",
			Help:      "number of request errors",
		},
		[]string{requestType},
	)

	requestLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "stratumn",
			Subsystem: "btc",
			Name:      "request_latency_ms",
			Help:      "latency of requests to the bitcoin blockchain",
			Buckets:   monitoring.DefaultLatencyBuckets,
		},
		[]string{requestType},
	)

	accountBalance = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "stratumn",
			Subsystem: "btc",
			Name:      "account_balance_satoshi",
			Help:      "balance of the bitcoin addresses used",
		},
		[]string{accountAddress},
	)
}
