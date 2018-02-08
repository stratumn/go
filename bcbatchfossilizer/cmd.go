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

package bcbatchfossilizer

import (
	"context"
	"flag"
	"time"

	"github.com/stratumn/go-indigocore/batchfossilizer"
	"github.com/stratumn/go-indigocore/blockchain"

	log "github.com/sirupsen/logrus"
)

var (
	interval        time.Duration
	maxLeaves       int
	path            string
	archive         bool
	exitBatch       bool
	fsync           bool
	key             string
	fee             int64
	bcyAPIKey       string
	limiterInterval time.Duration
	limiterSize     int
)

// RegisterFlags registers the flags used by RunWithFlags.
func RegisterFlags() {
	flag.DurationVar(&interval, "interval", batchfossilizer.DefaultInterval, "batch interval")
	flag.IntVar(&maxLeaves, "maxleaves", batchfossilizer.DefaultMaxLeaves, "maximum number of leaves in a Merkle tree")
	flag.StringVar(&path, "path", "", "an optional path to store files")
	flag.BoolVar(&archive, "archive", batchfossilizer.DefaultArchive, "whether to archive completed batches (requires path)")
	flag.BoolVar(&exitBatch, "exitbatch", batchfossilizer.DefaultStopBatch, "whether to do a batch on exit")
	flag.BoolVar(&fsync, "fsync", batchfossilizer.DefaultFSync, "whether to fsync after saving a pending hash (requires path)")
}

// RunWithFlags should be called after RegisterFlags and flag.Parse to initialize
// a bcbatchfossilizer using flag values.
func RunWithFlags(ctx context.Context, version, commit string, hashTS blockchain.HashTimestamper) *Fossilizer {
	log.Infof("%s v%s@%s", Description, version, commit[:7])

	a, err := New(&Config{
		HashTimestamper: hashTS,
	}, &batchfossilizer.Config{
		Version:   version,
		Commit:    commit,
		Interval:  interval,
		MaxLeaves: maxLeaves,
		Path:      path,
		Archive:   archive,
		StopBatch: exitBatch,
		FSync:     fsync,
	})
	if err != nil {
		log.WithField("error", err).Fatal("Failed to create blockchain batch fossilizer")
	}

	go func() {
		if err := a.Start(ctx); err != nil {
			log.WithField("error", err)
		}
	}()

	return a
}
