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

// Package batchfossilizer implements a fossilizer that fossilize batches of
// data using a Merkle tree. The evidence will contain the Merkle root, the
// Merkle path, and a timestamp.
package batchfossilizer

import (
	"context"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stratumn/go-indigocore/batchfossilizer/evidences"
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/fossilizer"
	"github.com/stratumn/go-indigocore/monitoring"
	"github.com/stratumn/go-indigocore/types"
	"github.com/stratumn/merkle"

	"go.opencensus.io/stats"
	"go.opencensus.io/trace"
)

const (
	// Name is the name set in the fossilizer's information.
	Name = "batch"

	// Description is the description set in the fossilizer's information.
	Description = "Stratumn Batch Fossilizer"

	// DefaultInterval is the default interval between batches.
	DefaultInterval = 10 * time.Minute

	// DefaultMaxLeaves if the default maximum number of leaves of a Merkle
	// tree.
	DefaultMaxLeaves = 32 * 1024

	// DefaultMaxSimBatches is the default maximum number of simultaneous
	// batches.
	DefaultMaxSimBatches = 1

	// DefaultArchive is whether to archive completed batches by default.
	DefaultArchive = true

	// DefaultStopBatch is whether to do a batch on stop by default.
	DefaultStopBatch = true

	// DefaultFSync is whether to fsync after saving a hash to disk by
	// default.
	DefaultFSync = false

	// PendingExt is the pending hashes filename extension.
	PendingExt = "pending"

	// DirPerm is the directory's permissions.
	DirPerm = 0600

	// FilePerm is the files's permissions.
	FilePerm = 0600
)

// Config contains configuration options for the fossilizer.
type Config struct {
	// A version string that will be set in the store's information.
	Version string

	// A git commit sha that will be set in the store's information.
	Commit string

	// Interval between batches.
	Interval time.Duration

	// Maximum number of leaves of a Merkle tree.
	MaxLeaves int

	// Maximum number of simultaneous batches.
	MaxSimBatches int

	// Where to store pending hashes.
	// If empty, pending hashes are not saved and will be lost if stopped
	// abruptly.
	Path string

	// Whether to archive completed batches.
	Archive bool

	// Whether to do a batch on stop.
	StopBatch bool

	// Whether to fsync after saving a hash to disk.
	FSync bool
}

// GetInterval returns the configuration's interval or the default value.
func (c *Config) GetInterval() time.Duration {
	if c.Interval > 0 {
		return c.Interval
	}
	return DefaultInterval
}

// GetMaxLeaves returns the configuration's maximum number of leaves of a Merkle
// tree or the default value.
func (c *Config) GetMaxLeaves() int {
	if c.MaxLeaves > 0 {
		return c.MaxLeaves
	}
	return DefaultMaxLeaves
}

// GetMaxSimBatches returns the configuration's maximum number of simultaneous
// batches or the default value.
func (c *Config) GetMaxSimBatches() int {
	if c.MaxSimBatches > 0 {
		return c.MaxSimBatches
	}
	return DefaultMaxSimBatches
}

// Info is the info returned by GetInfo.
type Info struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Commit      string `json:"commit"`
}

// Fossilizer is the type that
// implements github.com/stratumn/go-indigocore/fossilizer.Adapter.
type Fossilizer struct {
	config               *Config
	startedChan          chan chan struct{}
	fossilChan           chan *fossil
	resultChan           chan error
	batchChan            chan *batch
	stopChan             chan error
	semChan              chan struct{} // used to control number of simultaneous batches
	fossilizerEventMutex sync.RWMutex
	fossilizerEventChans []chan *fossilizer.Event
	waitGroup            sync.WaitGroup
	transformer          Transformer
	pending              *batch
	stopping             bool
}

// Transformer is the type of a function to transform results.
type Transformer func(evidence *cs.Evidence, data, meta []byte) (*fossilizer.Result, error)

// New creates an instance of a Fossilizer.
func New(config *Config) (*Fossilizer, error) {
	a := &Fossilizer{
		config:      config,
		startedChan: make(chan chan struct{}),
		fossilChan:  make(chan *fossil),
		resultChan:  make(chan error),
		batchChan:   make(chan *batch, 1),
		stopChan:    make(chan error, 1),
		semChan:     make(chan struct{}, config.GetMaxSimBatches()),
		pending:     newBatch(config.GetMaxLeaves()),
		stopping:    false,
	}

	a.SetTransformer(nil)

	if a.config.Path != "" {
		if err := a.ensurePath(); err != nil {
			return nil, err
		}
		if err := a.recover(); err != nil {
			return nil, err
		}
	}

	return a, nil
}

// GetInfo implements github.com/stratumn/go-indigocore/fossilizer.Adapter.GetInfo.
func (a *Fossilizer) GetInfo(ctx context.Context) (interface{}, error) {
	return &Info{
		Name:        Name,
		Description: Description,
		Version:     a.config.Version,
		Commit:      a.config.Commit,
	}, nil
}

// AddFossilizerEventChan implements
// github.com/stratumn/go-indigocore/fossilizer.Adapter.AddFossilizerEventChan.
func (a *Fossilizer) AddFossilizerEventChan(fossilizerEventChan chan *fossilizer.Event) {
	a.fossilizerEventMutex.Lock()
	defer a.fossilizerEventMutex.Unlock()
	a.fossilizerEventChans = append(a.fossilizerEventChans, fossilizerEventChan)
}

// Fossilize implements github.com/stratumn/go-indigocore/fossilizer.Adapter.Fossilize.
func (a *Fossilizer) Fossilize(ctx context.Context, data []byte, meta []byte) error {
	f := fossil{Meta: meta}
	f.Data = data
	a.fossilChan <- &f
	return <-a.resultChan
}

// Start starts the fossilizer.
func (a *Fossilizer) Start(ctx context.Context) error {
	var (
		interval = a.config.GetInterval()
		timer    = time.NewTimer(interval)
	)

	for {
		select {
		case c := <-a.startedChan:
			c <- struct{}{}
		case f := <-a.fossilChan:
			a.resultChan <- a.fossilize(f)
		case b := <-a.batchChan:
			if !timer.Stop() {
				<-timer.C
			}
			timer.Reset(interval)
			a.batch(b)
		case <-timer.C:
			timer.Stop()
			timer.Reset(interval)
			if len(a.pending.data) > 0 {
				a.sendBatch()
				log.WithField("interval", interval).Info("Requested new batch because the timer interval was reached")
			} else {
				log.WithField("interval", interval).Info("No batch is needed after the timer interval because there are no pending hashes")
			}
		case <-ctx.Done():
			e := a.stop(ctx.Err())
			return e
		}
	}
}

// Started return a channel that will receive once the fossilizer has started.
func (a *Fossilizer) Started() <-chan struct{} {
	c := make(chan struct{}, 1)
	a.startedChan <- c
	return c
}

// SetTransformer sets a transformer.
func (a *Fossilizer) SetTransformer(t Transformer) {
	if t != nil {
		a.transformer = t
	} else {
		a.transformer = func(evidence *cs.Evidence, data, meta []byte) (*fossilizer.Result, error) {
			return &fossilizer.Result{
				Evidence: *evidence,
				Data:     data,
				Meta:     meta,
			}, nil
		}
	}
}

func (a *Fossilizer) fossilize(f *fossil) error {
	if a.config.Path != "" {
		if a.pending.file == nil {
			if err := a.pending.open(a.pendingPath()); err != nil {
				return err
			}
		}
		if err := f.write(a.pending.encoder); err != nil {
			return err
		}
		if a.config.FSync {
			if err := a.pending.file.Sync(); err != nil {
				return err
			}
		}
	}

	a.pending.append(f)

	if numLeaves, maxLeaves := len(a.pending.data), a.config.GetMaxLeaves(); numLeaves >= maxLeaves {
		a.sendBatch()
		log.WithFields(log.Fields{
			"leaves": numLeaves,
			"max":    maxLeaves,
		}).Info("Requested new batch because the maximum number of leaves was reached")
	}

	return nil
}

func (a *Fossilizer) sendBatch() {
	b := a.pending
	a.pending = newBatch(a.config.GetMaxLeaves())
	a.batchChan <- b
}

func (a *Fossilizer) batch(b *batch) {
	log.Info("Starting batch...")

	stats.Record(context.Background(), batchCount.M(1))
	a.waitGroup.Add(1)

	go func() {
		ctx, span := trace.StartSpan(context.Background(), "batchfossilizer/batch")
		defer func() {
			a.waitGroup.Done()
			<-a.semChan
			span.End()
		}()

		a.semChan <- struct{}{}

		tree, err := merkle.NewStaticTree(b.data)
		if err != nil {
			span.SetStatus(trace.Status{Code: monitoring.Internal, Message: err.Error()})
			if !a.stopping {
				a.stop(err)
			}
			return
		}

		root := tree.Root()
		log.WithField("root", root).Info("Created tree with Merkle root")

		a.sendEvidence(ctx, tree, b.meta)
		log.WithField("root", root).Info("Sent evidence for batch with Merkle root")

		if b.file != nil {
			path := b.file.Name()

			if err := b.close(); err != nil {
				log.WithField("error", err).Warn("Failed to close batch file")
				span.SetStatus(trace.Status{Code: monitoring.Unknown, Message: err.Error()})
			}

			if a.config.Archive {
				archivePath := filepath.Join(a.config.Path, hex.EncodeToString(root))
				if err := os.Rename(path, archivePath); err == nil {
					log.WithFields(log.Fields{
						"old": filepath.Base(path),
						"new": filepath.Base(archivePath),
					}).Info("Renamed batch file")
					span.Annotate(nil, "Renamed batch file")
				} else {
					log.WithFields(log.Fields{
						"old":   filepath.Base(path),
						"new":   filepath.Base(archivePath),
						"error": err,
					}).Warn("Failed to rename batch file")
					span.SetStatus(trace.Status{Code: monitoring.Unknown, Message: err.Error()})
				}
			} else {
				if err := os.Remove(path); err == nil {
					log.WithField("file", filepath.Base(path)).Info("Removed pending hashes file")
					span.Annotate(nil, "Removed pending hashes file")
				} else {
					log.WithFields(log.Fields{
						"file":  filepath.Base(path),
						"error": err,
					}).Warn("Failed to remove batch file")
					span.Annotatef(nil, "Failed to remove batch file: %s", err.Error())
				}
			}
		}

		log.WithField("root", root).Info("Finished batch")
		span.Annotate(nil, "Finished batch")
	}()
}

func (a *Fossilizer) sendEvidence(ctx context.Context, tree *merkle.StaticTree, meta [][]byte) {
	ctx, span := trace.StartSpan(ctx, "batchfossilizer/sendEvidence")
	defer span.End()

	for i := 0; i < tree.LeavesLen(); i++ {
		var (
			err  error
			ts   = time.Now().UTC().Unix()
			root = tree.Root()
			leaf = tree.Leaf(i)
			d    = leaf[:]
			m    = meta[i]
			r    *fossilizer.Result
		)

		evidence := cs.Evidence{
			Backend:  Name,
			Provider: Name,
			Proof: &evidences.BatchProof{
				Timestamp: ts,
				Root:      types.NewBytes32FromBytes(root),
				Path:      tree.Path(i),
			},
		}

		if r, err = a.transformer(&evidence, d, m); err != nil {
			log.WithField("error", err).Error("Failed to transform evidence")
			span.SetStatus(trace.Status{Code: monitoring.InvalidArgument, Message: err.Error()})
		} else {
			event := &fossilizer.Event{
				EventType: fossilizer.DidFossilizeLink,
				Data:      r,
			}

			a.fossilizerEventMutex.RLock()
			defer a.fossilizerEventMutex.RUnlock()

			for _, c := range a.fossilizerEventChans {
				c <- event
			}

			stats.Record(ctx, fossilizedLinksCount.M(1))
		}
	}
}

func (a *Fossilizer) stop(err error) error {
	a.stopping = true
	if a.config.StopBatch {
		if len(a.pending.data) > 0 {
			a.batch(a.pending)
			log.Info("Requested final batch for pending hashes")
		} else {
			log.Info("No final batch is needed because there are no pending hashes")
		}
	}

	a.waitGroup.Wait()

	if a.pending.file != nil {
		if e := a.pending.file.Close(); e != nil {
			if err == nil {
				err = e
			} else {
				log.WithField("error", err).Error("Failed to close pending batch file")
			}
		}
	}

	return err
}

func (a *Fossilizer) ensurePath() error {
	if err := os.MkdirAll(a.config.Path, DirPerm); err != nil && !os.IsExist(err) {
		return err
	}
	return nil
}

func (a *Fossilizer) recover() error {
	matches, err := filepath.Glob(filepath.Join(a.config.Path, "*."+PendingExt))
	if err != nil {
		return err
	}

	for _, path := range matches {
		file, err := os.OpenFile(path, os.O_RDONLY|os.O_EXCL, FilePerm)
		if err != nil {
			return err
		}
		defer file.Close()

		dec := gob.NewDecoder(file)

		for {
			f, err := newFossilFromDecoder(dec)
			if err == io.EOF {
				break
			}
			if err = a.fossilize(f); err != nil {
				return err
			}
		}

		a.waitGroup.Wait()

		if err := os.Remove(path); err != nil {
			return err
		}

		log.WithField("file", filepath.Base(path)).Info("Recovered pending hashes file")
	}

	return nil
}

func (a *Fossilizer) pendingPath() string {
	filename := fmt.Sprintf("%d.%s", time.Now().UTC().UnixNano(), PendingExt)
	return filepath.Join(a.config.Path, filename)
}
