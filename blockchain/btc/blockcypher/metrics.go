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
	"log"

	"github.com/stratumn/go-core/monitoring"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

// Private metrics used only inside this package.
var (
	requestType    tag.Key
	requestCount   *stats.Int64Measure
	requestErr     *stats.Int64Measure
	requestLatency *stats.Float64Measure

	accountAddress tag.Key
	accountBalance *stats.Int64Measure
)

func init() {
	requestCount = stats.Int64(
		"stratumn/core/blockchain/btc/request_count",
		"number of requests to the bitcoin blockchain",
		stats.UnitDimensionless,
	)

	requestErr = stats.Int64(
		"stratumn/core/blockchain/btc/request_error",
		"number of request errors",
		stats.UnitDimensionless,
	)

	requestLatency = stats.Float64(
		"stratumn/core/blockchain/btc/request_latency",
		"latency of requests to the bitcoin blockchain",
		stats.UnitMilliseconds,
	)

	accountBalance = stats.Int64(
		"stratumn/core/blockchain/btc/account_balance",
		"balance of the bitcoin addresses used",
		stats.UnitDimensionless,
	)

	var err error
	if requestType, err = tag.NewKey("stratumn/core/blockchain/btc/request_type"); err != nil {
		log.Fatal(err)
	}
	if accountAddress, err = tag.NewKey("stratumn/core/blockchain/btc/address"); err != nil {
		log.Fatal(err)
	}

	err = view.Register(
		&view.View{
			Name:        "stratumn/core/blockchain/btc/request_count",
			Description: "number of requests to the bitcoin blockchain",
			Measure:     requestCount,
			Aggregation: view.Count(),
			TagKeys:     []tag.Key{requestType},
		},
		&view.View{
			Name:        "stratumn/core/blockchain/btc/request_error",
			Description: "number of request errors",
			Measure:     requestErr,
			Aggregation: view.Count(),
			TagKeys:     []tag.Key{requestType},
		},
		&view.View{
			Name:        "stratumn/core/blockchain/btc/request_latency",
			Description: "latency of requests to the bitcoin blockchain",
			Measure:     requestLatency,
			Aggregation: monitoring.DefaultLatencyDistribution,
			TagKeys:     []tag.Key{requestType},
		},
		&view.View{
			Name:        "stratumn/core/blockchain/btc/account_balance",
			Description: "balance of the bitcoin addresses used",
			Measure:     accountBalance,
			Aggregation: view.LastValue(),
			TagKeys:     []tag.Key{accountAddress},
		})
	if err != nil {
		log.Fatal(err)
	}
}
