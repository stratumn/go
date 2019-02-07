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

package tmpop

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/stratumn/go-core/monitoring"
)

const (
	// DefaultMetricsPort is the default port used to expose metrics.
	DefaultMetricsPort = 5090
)

const (
	txStatus = "tx_status"
)

var (
	blockCount prometheus.Counter
	txCount    *prometheus.CounterVec
	txPerBlock prometheus.Histogram
)

func init() {
	blockCount = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: monitoring.Stratumn,
			Subsystem: "tmpop",
			Name:      "block_count",
			Help:      "number of blocks created",
		},
	)

	txCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: monitoring.Stratumn,
			Subsystem: "tmpop",
			Name:      "tx_count",
			Help:      "number of transactions received",
		},
		[]string{txStatus},
	)

	txPerBlock = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: monitoring.Stratumn,
			Subsystem: "tmpop",
			Name:      "tx_per_block",
			Help:      "number of transactions per block",
			Buckets:   []float64{1, 5, 10, 50, 100},
		},
	)
}

// exposeMetrics configures metrics and traces exporters and
// exposes them to collectors.
func exposeMetrics(config *monitoring.Config) error {
	if !config.Monitor {
		return nil
	}

	if config.MetricsPort == 0 {
		config.MetricsPort = DefaultMetricsPort
	}

	metricsHandler, err := monitoring.Configure(config, "tmpop")
	if err != nil {
		return err
	}
	if metricsHandler != nil {
		metricsAddr := fmt.Sprintf(":%d", config.MetricsPort)

		monitoring.LogEntry().Infof("Exposing metrics on %s", metricsAddr)
		http.Handle("/metrics", metricsHandler)
		err := http.ListenAndServe(metricsAddr, nil)
		if err != nil {
			panic(err)
		}
	}
	return nil
}
