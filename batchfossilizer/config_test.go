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

package batchfossilizer_test

import (
	"testing"
	"time"

	"github.com/stratumn/go-core/batchfossilizer"
	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	t.Run("Interval", func(t *testing.T) {
		t.Run("default", func(t *testing.T) {
			i := batchfossilizer.Config{}.GetInterval()
			assert.Equal(t, batchfossilizer.DefaultInterval, i)
		})

		t.Run("configured", func(t *testing.T) {
			i := batchfossilizer.Config{
				Interval: time.Second,
			}.GetInterval()
			assert.Equal(t, time.Second, i)
		})
	})

	t.Run("MaxLeaves", func(t *testing.T) {
		t.Run("default", func(t *testing.T) {
			v := batchfossilizer.Config{}.GetMaxLeaves()
			assert.Equal(t, batchfossilizer.DefaultMaxLeaves, v)
		})

		t.Run("configured", func(t *testing.T) {
			v := batchfossilizer.Config{
				MaxLeaves: 42,
			}.GetMaxLeaves()
			assert.Equal(t, 42, v)
		})
	})

	t.Run("MaxSimBatches", func(t *testing.T) {
		t.Run("default", func(t *testing.T) {
			v := batchfossilizer.Config{}.GetMaxSimBatches()
			assert.Equal(t, batchfossilizer.DefaultMaxSimBatches, v)
		})

		t.Run("configured", func(t *testing.T) {
			v := batchfossilizer.Config{
				MaxSimBatches: 42,
			}.GetMaxSimBatches()
			assert.Equal(t, 42, v)
		})
	})
}
