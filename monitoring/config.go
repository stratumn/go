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
	"time"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"go.opencensus.io/exporter/jaeger"
	"go.opencensus.io/exporter/prometheus"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
)

// Available exporters.
const (
	PrometheusExporter  = "prometheus"
	JaegerExporter      = "jaeger"
	StackdriverExporter = "stackdriver"
)

const (
	// DefaultJaegerEndpoint is the default endpoint exposed
	// by the Jaeger collector.
	DefaultJaegerEndpoint = "http://jaeger:14268"
)

// Errors used by the configuration module.
var (
	ErrInvalidMetricsExporter = errors.New("metrics exporter should be 'prometheus' or 'stackdriver'")
	ErrInvalidTracesExporter  = errors.New("metrics exporter should be 'jaeger' or 'stackdriver'")
	ErrMissingExporterConfig  = errors.New("missing exporter configuration section")
	ErrMissingProjectID       = errors.New("missing stackdriver project id")
)

// Config contains options for monitoring.
type Config struct {
	// Set to true to monitor Stratumn components.
	Monitor bool

	// Port used to expose metrics.
	MetricsPort int

	// Ratio of traces to record.
	// If set to 1.0, all traces will be recorded.
	// This is what you should do locally or during a beta.
	// For production, you should set this to 0.25 or 0.5,
	// depending on your load.
	TraceSamplingRatio float64

	// MetricsExporter is the name of the metrics exporter.
	MetricsExporter string

	// TraceExporter is the name of the trace exporter.
	TracesExporter string

	// JaegerConfig options (if enabled).
	JaegerConfig *JaegerConfig

	// StackdriverConfig options (if enabled).
	StackdriverConfig *StackdriverConfig
}

// StackdriverConfig contains configuration options for Stackdriver (metrics and tracing).
type StackdriverConfig struct {
	// ProjectID is the identifier of the Stackdriver project
	ProjectID string
}

// JaegerConfig contains configuration options for Jaeger (tracing).
type JaegerConfig struct {
	// Endpoint is the address of the Jaeger agent to collect traces.
	Endpoint string
}

// Validate the stackdriver configuration section.
func (c *StackdriverConfig) Validate() error {
	if c == nil {
		return ErrMissingExporterConfig
	}

	if c.ProjectID == "" {
		return ErrMissingProjectID
	}

	return nil
}

// Validate the jaeger configuration section.
func (c *JaegerConfig) Validate() error {
	if c == nil {
		return ErrMissingExporterConfig
	}

	return nil
}

func configureMetricsExporter(config *Config) (exporter view.Exporter, err error) {
	switch config.MetricsExporter {
	case PrometheusExporter:
		exporter, err = prometheus.NewExporter(prometheus.Options{})
		if err != nil {
			return nil, err
		}
	case StackdriverExporter:
		if err := config.StackdriverConfig.Validate(); err != nil {
			return nil, err
		}
		exporter, err = stackdriver.NewExporter(stackdriver.Options{
			ProjectID: config.StackdriverConfig.ProjectID,
		})
		if err != nil {
			return nil, err
		}

	default:
		return nil, ErrInvalidMetricsExporter
	}

	view.RegisterExporter(exporter)
	view.SetReportingPeriod(1 * time.Second)

	return exporter, nil
}

func configureTracesExporter(config *Config, serviceName string) (exporter trace.Exporter, err error) {
	switch config.TracesExporter {
	case JaegerExporter:
		if err := config.JaegerConfig.Validate(); err != nil {
			return nil, err
		}
		if len(config.JaegerConfig.Endpoint) == 0 {
			config.JaegerConfig.Endpoint = DefaultJaegerEndpoint
		}
		exporter, err = jaeger.NewExporter(jaeger.Options{
			Endpoint:    config.JaegerConfig.Endpoint,
			ServiceName: serviceName,
		})
		if err != nil {
			return nil, err
		}
	case StackdriverExporter:
		if err := config.StackdriverConfig.Validate(); err != nil {
			return nil, err
		}
		exporter, err = stackdriver.NewExporter(stackdriver.Options{
			ProjectID: config.StackdriverConfig.ProjectID,
		})
		if err != nil {
			return nil, err
		}
	default:
		return nil, ErrInvalidTracesExporter
	}

	trace.ApplyConfig(trace.Config{DefaultSampler: trace.ProbabilitySampler(config.TraceSamplingRatio)})
	trace.RegisterExporter(exporter)

	return exporter, nil
}

// Configure configures metrics and trace monitoring.
// If metrics need to be exposed on an http route ('/metrics'),
// this function returns an http.Handler. It returns nil otherwise.
func Configure(config *Config, serviceName string) (http.Handler, error) {
	if !config.Monitor {
		return nil, nil
	}
	_, err := configureTracesExporter(config, serviceName)
	if err != nil {
		return nil, err
	}

	metricsExporter, err := configureMetricsExporter(config)
	if err != nil {
		return nil, err
	}
	if handler, ok := metricsExporter.(http.Handler); ok {
		return handler, nil
	}

	return nil, nil
}
