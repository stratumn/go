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

package batchfossilizer

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/stratumn/go-core/monitoring"
)

var (
	batchCount           prometheus.Counter
	fossilizedLinksCount prometheus.Counter
)

func init() {
	batchCount = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: monitoring.Stratumn,
			Subsystem: "batchfossilizer",
			Name:      "batch_count",
			Help:      "number of batches sent",
		},
	)

	fossilizedLinksCount = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: monitoring.Stratumn,
			Subsystem: "batchfossilizer",
			Name:      "fossilized_links_count",
			Help:      "number of links fossilized",
		},
	)
}
