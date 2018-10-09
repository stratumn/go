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

package validators

import (
	log "github.com/sirupsen/logrus"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

var (
	linksErr *stats.Int64Measure
	linkErr  tag.Key
)

func init() {
	linksErr = stats.Int64(
		"stratumn/core/validation/links_error",
		"number of invalid links",
		stats.UnitDimensionless,
	)

	var err error
	if linkErr, err = tag.NewKey("link_error"); err != nil {
		log.Fatal(err)
	}

	err = view.Register(
		&view.View{
			Name:        "stratumn_core_validation_links_error",
			Description: "number of invalid links",
			Measure:     linksErr,
			Aggregation: view.Count(),
			TagKeys:     []tag.Key{linkErr},
		},
	)
	if err != nil {
		log.Fatal(err)
	}
}
