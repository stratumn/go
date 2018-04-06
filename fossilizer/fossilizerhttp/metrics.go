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

package fossilizerhttp

import (
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/stratumn/go-indigocore/monitoring"

	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/stats/view"
)

func init() {
	if err := view.Register(ochttp.DefaultServerViews...); err != nil {
		log.Fatal(err)
	}
}

// exposeMetrics configures metrics and traces exporters and
// exposes them to collectors.
func (s *Server) exposeMetrics(config *monitoring.Config) {
	if !config.Monitor {
		return
	}

	metricsExporter := monitoring.Configure(config, "indigo-fossilizer")
	s.GetRaw(
		"/metrics",
		func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
			metricsExporter.ServeHTTP(w, r)
		},
	)
}
