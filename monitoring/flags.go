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

import "flag"

var (
	monitor            bool
	metricsPort        int
	jaegerEndpoint     string
	traceSamplingRatio float64
)

// RegisterFlags registers the command-line monitoring flags.
func RegisterFlags() {
	flag.BoolVar(&monitor, "monitoring.active", true, "Set to true to activate monitoring")
	flag.IntVar(&metricsPort, "monitoring.port", 0, "Port to use to expose metrics, for example 5001")
	flag.StringVar(&jaegerEndpoint, "monitoring.jaeger_endpoint", DefaultJaegerEndpoint, "Endpoint where a Jaeger agent is running")
	flag.Float64Var(&traceSamplingRatio, "monitoring.trace_sampling_ratio", 1.0, "Set an appropriate sampling ratio depending on your load")
}

// ConfigurationFromFlags builds configuration from user-provided
// command-line flags.
func ConfigurationFromFlags() *Config {
	return &Config{
		Monitor:            monitor,
		MetricsPort:        metricsPort,
		JaegerEndpoint:     jaegerEndpoint,
		TraceSamplingRatio: traceSamplingRatio,
	}
}
