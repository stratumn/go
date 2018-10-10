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

package validation

import (
	log "github.com/sirupsen/logrus"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
)

var (
	linksCount *stats.Int64Measure
)

func init() {
	linksCount = stats.Int64(
		"stratumn/core/validation/links_count",
		"number of links validated",
		stats.UnitDimensionless,
	)

	err := view.Register(
		&view.View{
			Name:        "stratumn/core/validation/links_count",
			Description: "number of links validated",
			Measure:     linksCount,
			Aggregation: view.Count(),
		},
	)
	if err != nil {
		log.Fatal(err)
	}
}
