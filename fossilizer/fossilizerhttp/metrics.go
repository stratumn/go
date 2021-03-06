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

package fossilizerhttp

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/stratumn/go-core/monitoring"
)

const (
	// Component name for monitoring.
	Component = "fossilizer"
)

// exposeMetrics configures metrics and traces exporters and
// exposes them to collectors.
func (s *Server) exposeMetrics(config *monitoring.Config) error {
	if !config.Monitor {
		return nil
	}

	metricsHandler, err := monitoring.Configure(config, "fossilizer")
	if err != nil {
		return err
	}

	if metricsHandler != nil {
		s.GetRaw(
			"/metrics",
			func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
				metricsHandler.ServeHTTP(w, r)
			},
		)
	}

	return nil
}
