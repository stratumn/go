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
	"testing"

	"github.com/stratumn/go-core/monitoring"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	t.Run("monitoring turned off", func(t *testing.T) {
		c := &monitoring.Config{Monitor: false}
		handler, err := monitoring.Configure(c, "test")

		require.NoError(t, err)
		require.Nil(t, handler)
	})

	t.Run("invalid exporter", func(t *testing.T) {
		c := &monitoring.Config{
			Monitor:     true,
			MetricsPort: 1,
			Exporter:    "stackdriver",
		}

		handler, err := monitoring.Configure(c, "test")
		require.EqualError(t, err, monitoring.ErrInvalidExporter.Error())
		require.Nil(t, handler)
	})

	t.Run("prometheus exporter", func(t *testing.T) {
		c := &monitoring.Config{
			Monitor:  true,
			Exporter: monitoring.PrometheusExporter,
		}

		handler, err := monitoring.Configure(c, "test")
		require.NoError(t, err)
		require.NotNil(t, handler)

		_, ok := handler.(http.Handler)
		assert.True(t, ok)
	})

	t.Run("elastic exporter", func(t *testing.T) {
		c := &monitoring.Config{
			Monitor:  true,
			Exporter: monitoring.ElasticExporter,
		}

		handler, err := monitoring.Configure(c, "test")
		require.NoError(t, err)
		require.Nil(t, handler)
	})
}
