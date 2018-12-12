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

import (
	"flag"
	"time"
)

// Flags variables.
var (
	interval      time.Duration
	maxLeaves     int
	maxSimBatches int
)

// RegisterFlags registers the flags used by batch fossilizers.
func RegisterFlags() {
	flag.DurationVar(&interval, "interval", DefaultInterval, "batch interval")
	flag.IntVar(&maxLeaves, "maxleaves", DefaultMaxLeaves, "maximum number of leaves in a Merkle tree")
	flag.IntVar(&maxSimBatches, "maxsimbatches", DefaultMaxSimBatches, "maximum number of simultaneous batches")
}

// ConfigFromFlags builds the configuration from command-line flags.
func ConfigFromFlags(version string, commit string) *Config {
	return &Config{
		Version:       version,
		Commit:        commit,
		Interval:      interval,
		MaxLeaves:     maxLeaves,
		MaxSimBatches: maxSimBatches,
	}
}
