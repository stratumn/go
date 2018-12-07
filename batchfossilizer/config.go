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

package batchfossilizer

import "time"

const (
	// DefaultInterval is the default interval between batches.
	DefaultInterval = 10 * time.Minute

	// DefaultMaxLeaves if the default maximum number of leaves of a Merkle
	// tree.
	DefaultMaxLeaves = 32 * 1024

	// DefaultMaxSimBatches is the default maximum number of simultaneous
	// batches.
	DefaultMaxSimBatches = 1
)

// Config contains configuration options for the fossilizer.
type Config struct {
	// A version string that will be set in the store's information.
	Version string

	// A git commit hash that will be set in the store's information.
	Commit string

	// Interval between batches.
	Interval time.Duration

	// Maximum number of leaves of a Merkle tree.
	MaxLeaves int

	// Maximum number of simultaneous batches.
	MaxSimBatches int
}

// GetInterval returns the configuration's interval or the default value.
func (c Config) GetInterval() time.Duration {
	if c.Interval > 0 {
		return c.Interval
	}

	return DefaultInterval
}

// GetMaxLeaves returns the configuration's maximum number of leaves of a Merkle
// tree or the default value.
func (c Config) GetMaxLeaves() int {
	if c.MaxLeaves > 0 {
		return c.MaxLeaves
	}

	return DefaultMaxLeaves
}

// GetMaxSimBatches returns the configuration's maximum number of simultaneous
// batches or the default value.
func (c Config) GetMaxSimBatches() int {
	if c.MaxSimBatches > 0 {
		return c.MaxSimBatches
	}

	return DefaultMaxSimBatches
}
