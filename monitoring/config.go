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
	"errors"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"

	"go.elastic.co/apm"
	"go.elastic.co/apm/module/apmprometheus"
)

// Available exporters.
const (
	PrometheusExporter = "prometheus"
	ElasticExporter    = "elastic"
)

// Errors used by the configuration module.
var (
	ErrInvalidExporter = errors.New("exporter should be 'prometheus' or 'elastic'")
)

// Config contains options for monitoring.
type Config struct {
	// Set to true to monitor Stratumn components.
	Monitor bool

	// Port used to expose metrics.
	MetricsPort int

	// Exporter is the name of the metrics and traces exporter.
	Exporter string
}

// Configure configures metrics and trace monitoring.
// If metrics need to be exposed on an http route ('/metrics'),
// this function returns an http.Handler. It returns nil otherwise.
func Configure(config *Config, serviceName string) (http.Handler, error) {
	if !config.Monitor {
		return nil, nil
	}

	switch config.Exporter {
	case PrometheusExporter:
		handler := promhttp.Handler()
		log.Info("Prometheus handler registered: metrics are available for pulling")
		return handler, nil
	case ElasticExporter:
		// Plug the default prometheus gatherer to APM.
		// This will export custom metrics regularly.
		// It doesn't need to expose a handler, it uses a push model.
		apm.DefaultTracer.RegisterMetricsGatherer(
			apmprometheus.Wrap(prometheus.DefaultGatherer),
		)

		log.Info("Elastic APM registered: metrics and traces will be pushed regularly")
		return nil, nil
	default:
		return nil, ErrInvalidExporter
	}
}
