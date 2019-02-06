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

package bufferedbatch

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/stratumn/go-core/monitoring"
)

const (
	writeStatus = "write_status"
)

var (
	batchCount    prometheus.Counter
	linksPerBatch prometheus.Histogram
	writeCount    *prometheus.CounterVec
)

func init() {
	batchCount = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: monitoring.Stratumn,
			Subsystem: "bufferedbatch",
			Name:      "batch_count",
			Help:      "number of batches created",
		},
	)

	linksPerBatch = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: monitoring.Stratumn,
			Subsystem: "bufferedbatch",
			Name:      "links_per_batch",
			Help:      "number of links per batch",
			Buckets:   []float64{1, 5, 10, 50, 100, 1000},
		},
	)

	writeCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: monitoring.Stratumn,
			Subsystem: "bufferedbatch",
			Name:      "write_count",
			Help:      "number of batch writes",
		},
		[]string{writeStatus},
	)
}
