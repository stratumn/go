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

package batchfossilizer

import (
	"log"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
)

var (
	batchCount           *stats.Int64Measure
	fossilizedLinksCount *stats.Int64Measure
)

func init() {
	batchCount = stats.Int64(
		"stratumn/indigocore/batchfossilizer/batch_count",
		"number of batches sent",
		stats.UnitNone,
	)

	fossilizedLinksCount = stats.Int64(
		"stratumn/indigocore/batchfossilizer/fossilized_links_count",
		"number of links fossilized",
		stats.UnitNone,
	)

	if err := view.Subscribe(
		&view.View{
			Name:        "batch_count",
			Description: "number of batches sent",
			Measure:     batchCount,
			Aggregation: view.Count(),
		},
		&view.View{
			Name:        "fossilized_links_count",
			Description: "number of links fossilized",
			Measure:     fossilizedLinksCount,
			Aggregation: view.Count(),
		}); err != nil {
		log.Fatal(err)
	}
}
