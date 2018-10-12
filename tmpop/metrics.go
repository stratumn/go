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

	log "github.com/sirupsen/logrus"
	"github.com/stratumn/go-core/monitoring"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

const (
	// DefaultMetricsPort is the default port used to expose metrics.
	DefaultMetricsPort = 5090
)

var (
	blockCount *stats.Int64Measure

	txCount    *stats.Int64Measure
	txPerBlock *stats.Int64Measure
	txStatus   tag.Key
)

func init() {
	blockCount = stats.Int64(
		"stratumn/core/tmpop/block_count",
		"number of blocks created",
		stats.UnitDimensionless,
	)

	txCount = stats.Int64(
		"stratumn/core/tmpop/tx_count",
		"number of transactions received",
		stats.UnitDimensionless,
	)

	txPerBlock = stats.Int64(
		"stratumn/core/tmpop/tx_per_block",
		"number of transactions per block",
		stats.UnitDimensionless,
	)

	var err error
	if txStatus, err = tag.NewKey("stratumn/core/tmpop/tx_status"); err != nil {
		log.Fatal(err)
	}

	if err = view.Register(
		&view.View{
			Name:        "stratumn/core/tmpop/block_count",
			Description: "number of blocks created",
			Measure:     blockCount,
			Aggregation: view.Count(),
		},
		&view.View{
			Name:        "stratumn/core/tmpop/tx_count",
			Description: "number of transactions received",
			Measure:     txCount,
			Aggregation: view.Count(),
			TagKeys:     []tag.Key{txStatus},
		},
		&view.View{
			Name:        "stratumn/core/tmpop/tx_per_block",
			Description: "number of transactions per block",
			Measure:     txPerBlock,
			Aggregation: view.Distribution(1, 5, 10, 50, 100),
		}); err != nil {
		log.Fatal(err)
	}
}

// exposeMetrics configures metrics and traces exporters and
// exposes them to collectors.
func exposeMetrics(config *monitoring.Config) {
	if !config.Monitor {
		return
	}

	if config.MetricsPort == 0 {
		config.MetricsPort = DefaultMetricsPort
	}

	metricsHandler := monitoring.Configure(config, "tmpop")
	if metricsHandler != nil {
		metricsAddr := fmt.Sprintf(":%d", config.MetricsPort)

		log.Infof("Exposing metrics on %s", metricsAddr)
		http.Handle("/metrics", metricsHandler)
		err := http.ListenAndServe(metricsAddr, nil)
		if err != nil {
			panic(err)
		}
	}
}
