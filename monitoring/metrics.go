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
	log "github.com/sirupsen/logrus"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

// Default distributions used by views.
var (
	DefaultLatencyDistribution = view.Distribution(0, 1, 2, 3, 4, 5, 6, 8, 10, 13, 16, 20, 25, 30, 40, 50, 65, 80, 100, 130, 160, 200, 250, 300, 400, 500, 650, 800, 1000, 2000, 5000, 10000, 20000, 50000, 100000)
)

var (
	storeRequestType    tag.Key
	storeRequestCount   *stats.Int64Measure
	storeRequestErr     *stats.Int64Measure
	storeRequestLatency *stats.Float64Measure
)

func init() {
	storeRequestCount = stats.Int64(
		"stratumn/core/store/request_count",
		"number of requests to the store",
		stats.UnitDimensionless,
	)

	storeRequestErr = stats.Int64(
		"stratumn/core/store/request_error",
		"number of request errors",
		stats.UnitDimensionless,
	)

	storeRequestLatency = stats.Float64(
		"stratumn/core/store/request_latency",
		"latency of store requests",
		stats.UnitMilliseconds,
	)

	var err error
	if storeRequestType, err = tag.NewKey("stratumn/core/store/request_type"); err != nil {
		log.Fatal(err)
	}

	err = view.Register(
		&view.View{
			Name:        "stratumn/core/store/request_count",
			Description: "number of requests to the store",
			Measure:     storeRequestCount,
			Aggregation: view.Count(),
			TagKeys:     []tag.Key{storeRequestType},
		},
		&view.View{
			Name:        "stratumn/core/store/request_error",
			Description: "number of request errors",
			Measure:     storeRequestErr,
			Aggregation: view.Count(),
			TagKeys:     []tag.Key{storeRequestType},
		},
		&view.View{
			Name:        "stratumn/core/store/request_latency",
			Description: "latency of store requests",
			Measure:     storeRequestLatency,
			Aggregation: DefaultLatencyDistribution,
			TagKeys:     []tag.Key{storeRequestType},
		})
	if err != nil {
		log.Fatal(err)
	}
}
