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

package tmpop

import (
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	"go.opencensus.io/exporter/jaeger"
	"go.opencensus.io/exporter/prometheus"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
	"go.opencensus.io/trace"
)

var (
	blockCount *stats.Int64Measure

	txCount    *stats.Int64Measure
	txPerBlock *stats.Int64Measure
	txStatus   tag.Key
)

func init() {
	var err error
	if blockCount, err = stats.Int64(
		"stratumn/indigocore/tmpop/block_count",
		"number of blocks created",
		stats.UnitNone,
	); err != nil {
		log.Fatal(err)
	}

	if txCount, err = stats.Int64(
		"stratumn/indigocore/tmpop/tx_count",
		"number of transactions received",
		stats.UnitNone,
	); err != nil {
		log.Fatal(err)
	}

	if txPerBlock, err = stats.Int64(
		"stratumn/indigocore/tmpop/tx_per_block",
		"number of transactions per block",
		stats.UnitNone,
	); err != nil {
		log.Fatal(err)
	}

	if txStatus, err = tag.NewKey("tx_status"); err != nil {
		log.Fatal(err)
	}

	if err = view.Subscribe(
		&view.View{
			Name:        "block_count",
			Description: "number of blocks created",
			Measure:     blockCount,
			Aggregation: view.Count(),
		},
		&view.View{
			Name:        "tx_count",
			Description: "number of transactions received",
			Measure:     txCount,
			Aggregation: view.Count(),
			TagKeys:     []tag.Key{txStatus},
		},
		&view.View{
			Name:        "tx_per_block",
			Description: "number of transactions per block",
			Measure:     txPerBlock,
			Aggregation: view.Distribution(1, 5, 10, 50, 100),
		}); err != nil {
		log.Fatal(err)
	}
}

// exposeMetrics configures metrics and traces exporters and
// exposes them to collectors.
func exposeMetrics() {
	metricsExporter, err := prometheus.NewExporter(prometheus.Options{})
	if err != nil {
		log.Fatal(err)
	}

	view.RegisterExporter(metricsExporter)
	view.SetReportingPeriod(1 * time.Second)

	traceExporter, err := jaeger.NewExporter(jaeger.Options{
		Endpoint:    "http://jaeger:14268",
		ServiceName: "indigo-tmpop",
	})
	if err != nil {
		log.Fatal(err)
	}

	trace.SetDefaultSampler(trace.AlwaysSample())
	trace.RegisterExporter(traceExporter)

	log.Infof("Exposing metrics on :5001")
	http.Handle("/metrics", metricsExporter)
	http.ListenAndServe(":5001", nil)
}
