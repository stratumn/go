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

package bufferedbatch

import (
	"log"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

var (
	batchCount    *stats.Int64Measure
	linksPerBatch *stats.Int64Measure
	writeCount    *stats.Int64Measure
	writeStatus   tag.Key
)

func init() {
	batchCount = stats.Int64(
		"stratumn/indigocore/bufferedbatch/batch_count",
		"number of batches created",
		stats.UnitNone,
	)

	linksPerBatch = stats.Int64(
		"stratumn/indigocore/bufferedbatch/links_per_batch",
		"number of links per batch",
		stats.UnitNone,
	)

	writeCount = stats.Int64(
		"stratumn/indigocore/bufferedbatch/write_count",
		"number of batch writes",
		stats.UnitNone,
	)

	var err error
	if writeStatus, err = tag.NewKey("batch_write_status"); err != nil {
		log.Fatal(err)
	}

	if err = view.Register(
		&view.View{
			Name:        "stratumn_indigocore_bufferedbatch_batch_count",
			Description: "number of batches created",
			Measure:     batchCount,
			Aggregation: view.Count(),
		},
		&view.View{
			Name:        "stratumn_indigocore_bufferedbatch_write_count",
			Description: "number of batch writes",
			Measure:     writeCount,
			Aggregation: view.Count(),
			TagKeys:     []tag.Key{writeStatus},
		},
		&view.View{
			Name:        "stratumn_indigocore_bufferedbatch_links_per_batch",
			Description: "number of links per batch",
			Measure:     linksPerBatch,
			Aggregation: view.Distribution(1, 5, 10, 50, 100),
		}); err != nil {
		log.Fatal(err)
	}
}
