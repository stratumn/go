package bcbatchfossilizer

import (
	"context"
	"flag"
	"time"

	"github.com/stratumn/sdk/batchfossilizer"
	"github.com/stratumn/sdk/blockchain"

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
