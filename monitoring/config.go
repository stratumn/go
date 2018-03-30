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

package monitoring

import (
	"log"
	"time"

	"go.opencensus.io/exporter/jaeger"
	"go.opencensus.io/exporter/prometheus"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
)

const (
	// DefaultJaegerEndpoint is the default endpoint exposed
	// by the Jaeger collector.
	DefaultJaegerEndpoint = "http://jaeger:14268"
)

// Config contains options for monitoring.
type Config struct {
	// Set to true to monitor Indigo components.
	Monitor bool
	// Port used to expose prometheus metrics.
	MetricsPort int
	// Jaeger collector url.
	JaegerEndpoint string
	// Ratio of traces to record.
	// If set to 1.0, all traces will be recorded.
	// This is what you should do locally or during a beta.
	// For production, you should set this to 0.25 or 0.5,
	// depending on your load.
	TraceSamplingRatio float64
}

// Configure configures metrics and trace monitoring.
// It returns the metrics exporter that you should expose
// on a /metrics http route.
func Configure(config *Config, serviceName string) *prometheus.Exporter {
	if !config.Monitor {
		return nil
	}

	metricsExporter, err := prometheus.NewExporter(prometheus.Options{})
	if err != nil {
		log.Fatal(err)
	}

	view.RegisterExporter(metricsExporter)
	view.SetReportingPeriod(1 * time.Second)

	if len(config.JaegerEndpoint) == 0 {
		config.JaegerEndpoint = DefaultJaegerEndpoint
	}

	traceExporter, err := jaeger.NewExporter(jaeger.Options{
		Endpoint:    config.JaegerEndpoint,
		ServiceName: serviceName,
	})
	if err != nil {
		log.Fatal(err)
	}

	trace.ApplyConfig(trace.Config{DefaultSampler: trace.ProbabilitySampler(config.TraceSamplingRatio)})
	trace.RegisterExporter(traceExporter)

	return metricsExporter
}
