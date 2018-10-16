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

import "flag"

var (
	monitor            bool
	traceSamplingRatio float64
	metricsPort        int
	reportingPeriod    int

	metricsExporter string
	tracesExporter  string

	jaegerEndpoint       string
	stackdriverProjectID string
)

// RegisterFlags registers the command-line monitoring flags.
func RegisterFlags() {
	flag.BoolVar(&monitor, "monitoring.active", true, "Set to true to activate monitoring")
	flag.IntVar(&metricsPort, "monitoring.metrics.port", 0, "Port to use to expose metrics, for example 5001")
	flag.IntVar(&reportingPeriod, "monitoring.metrics.reporting_period", DefaultReportingPeriod, "Interval between reporting aggregated views (in seconds)")
	flag.Float64Var(&traceSamplingRatio, "monitoring.trace_sampling_ratio", 1.0, "Set an appropriate sampling ratio depending on your load")

	flag.StringVar(&metricsExporter, "monitoring.exporter.metrics", PrometheusExporter, "Exporter for metrics (either Prometheus or Stackdriver)")
	flag.StringVar(&tracesExporter, "monitoring.exporter.traces", JaegerExporter, "Exporter for traces (either Jaeger or Stackdriver)")

	flag.StringVar(&jaegerEndpoint, "monitoring.jaeger.endpoint", DefaultJaegerEndpoint, "Endpoint where a Jaeger agent is running")
	flag.StringVar(&stackdriverProjectID, "monitoring.stackdriver.projectID", "", "ID of the project for which we want to export traces or metrics")
}

// ConfigurationFromFlags builds configuration from user-provided
// command-line flags.
func ConfigurationFromFlags() *Config {
	config := &Config{
		Monitor: monitor,

		TraceSamplingRatio:     traceSamplingRatio,
		MetricsReportingPeriod: reportingPeriod,

		MetricsExporter: metricsExporter,
		TracesExporter:  tracesExporter,
	}

	switch metricsExporter {
	case PrometheusExporter:
		break
	case StackdriverExporter:
		config.StackdriverConfig = &StackdriverConfig{
			ProjectID: stackdriverProjectID,
		}
	}

	switch tracesExporter {
	case JaegerExporter:
		config.JaegerConfig = &JaegerConfig{
			Endpoint: jaegerEndpoint,
		}
	case StackdriverExporter:
		config.StackdriverConfig = &StackdriverConfig{
			ProjectID: stackdriverProjectID,
		}
	}

	return config
}
