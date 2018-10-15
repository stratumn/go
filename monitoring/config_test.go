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

package monitoring_test

import (
	"net/http"
	"os"
	"testing"

	"github.com/stratumn/go-core/monitoring"
	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	serviceName := "svc"
	credentialsFile := "./fixtures/credentials_test.json"

	t.Run("with default values", func(t *testing.T) {
		c := &monitoring.Config{
			Monitor:            true,
			MetricsPort:        1,
			TraceSamplingRatio: 1,
			MetricsExporter:    monitoring.PrometheusExporter,
			TracesExporter:     monitoring.JaegerExporter,
			JaegerConfig:       &monitoring.JaegerConfig{},
		}
		handler, err := monitoring.Configure(c, serviceName)
		require.NoError(t, err)
		_, ok := handler.(http.Handler)
		require.True(t, ok)
	})

	t.Run("with stackdriver", func(t *testing.T) {
		os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", credentialsFile)

		c := &monitoring.Config{
			Monitor:            true,
			MetricsPort:        1,
			TraceSamplingRatio: 1,
			MetricsExporter:    monitoring.StackdriverExporter,
			TracesExporter:     monitoring.StackdriverExporter,
			StackdriverConfig: &monitoring.StackdriverConfig{
				ProjectID: "projectID",
			},
		}
		handler, err := monitoring.Configure(c, serviceName)
		require.NoError(t, err)
		require.Nil(t, handler)
	})

	t.Run("invalid traces exporter", func(t *testing.T) {
		c := &monitoring.Config{
			Monitor:            true,
			MetricsPort:        1,
			TraceSamplingRatio: 1,
			MetricsExporter:    monitoring.StackdriverExporter,
			TracesExporter:     "invalid",
			StackdriverConfig: &monitoring.StackdriverConfig{
				ProjectID: "projectID",
			},
		}
		handler, err := monitoring.Configure(c, serviceName)
		require.EqualError(t, err, monitoring.ErrInvalidTracesExporter.Error())
		require.Nil(t, handler)
	})

	t.Run("invalid metrics exporter", func(t *testing.T) {
		c := &monitoring.Config{
			Monitor:            true,
			MetricsPort:        1,
			TraceSamplingRatio: 1,
			MetricsExporter:    "invalid",
			TracesExporter:     monitoring.StackdriverExporter,
			StackdriverConfig: &monitoring.StackdriverConfig{
				ProjectID: "projectID",
			},
		}
		handler, err := monitoring.Configure(c, serviceName)
		require.EqualError(t, err, monitoring.ErrInvalidMetricsExporter.Error())
		require.Nil(t, handler)
	})

	t.Run("missing exporter config", func(t *testing.T) {
		c := &monitoring.Config{
			Monitor:            true,
			MetricsPort:        1,
			TraceSamplingRatio: 1,
			MetricsExporter:    monitoring.PrometheusExporter,
			TracesExporter:     monitoring.JaegerExporter,
			JaegerConfig:       nil,
		}
		handler, err := monitoring.Configure(c, serviceName)
		require.EqualError(t, err, monitoring.ErrMissingExporterConfig.Error())
		require.Nil(t, handler)
	})

	t.Run("missing stackdriver project ID", func(t *testing.T) {
		c := &monitoring.Config{
			Monitor:            true,
			MetricsPort:        1,
			TraceSamplingRatio: 1,
			MetricsExporter:    monitoring.StackdriverExporter,
			TracesExporter:     monitoring.StackdriverExporter,
			StackdriverConfig:  &monitoring.StackdriverConfig{},
		}
		handler, err := monitoring.Configure(c, serviceName)
		require.EqualError(t, err, monitoring.ErrMissingProjectID.Error())
		require.Nil(t, handler)
	})
}
