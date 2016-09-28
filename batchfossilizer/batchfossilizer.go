// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package batchfossilizer implements a fossilizer that fossilize batches of data using a Merkle tree.
// The evidence will contain the Merkle root, the Merkle path, and a timestamp.
package batchfossilizer

import (
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/stratumn/go/fossilizer"
	"github.com/stratumn/go/types"

	"github.com/stratumn/goprivate/merkle"
)

const (
	// Name is the name set in the fossilizer's information.
	Name = "batch"

	// Description is the description set in the fossilizer's information.
	Description = "Stratumn Batch Fossilizer"

	// DefaultInterval is the default interval between batches.
	DefaultInterval = 10 * time.Minute

	// DefaultMaxLeaves if the default maximum number of leaves of a Merkle tree.
	DefaultMaxLeaves = 32 * 1024

	// DefaultMaxSimBatches is the default maximum number of simultaneous batches.
	DefaultMaxSimBatches = 1

	// DefaultArchive is whether to archive completed batches by default.
	DefaultArchive = true

	// DefaultStopBatch is whether to do a batch on stop by default.
	DefaultStopBatch = true

	// DefaultFSync is whether to fsync after saving a hash to disk by default.
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
	// If empty, pending hashes are not saved and will be lost if stopped abruptly.
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

// GetMaxLeaves returns the configuration's maximum number of leaves of a Merkle tree or the default value.
func (c *Config) GetMaxLeaves() int {
	if c.MaxLeaves > 0 {
		return c.MaxLeaves
	}
	return DefaultMaxLeaves
}

// GetMaxSimBatches returns the configuration's maximum number of simultaneous batches or the default value.
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

// Evidence is the evidence sent to the result channel.
type Evidence struct {
	Time int64          `json:"time"`
	Root *types.Bytes32 `json:"merkleRoot"`
	Path merkle.Path    `json:"merklePath"`
}

// EvidenceWrapper wraps evidence with a namespace.
type EvidenceWrapper struct {
	Evidence *Evidence `json:"batch"`
}

// Fossilizer is the type that implements github.com/stratumn/go/fossilizer.Adapter.
type Fossilizer struct {
	config      *Config
	startedChan chan chan struct{}
	fossilChan  chan *fossil
	resultChan  chan error
	batchChan   chan *batch
	stopChan    chan error
	semChan     chan struct{}
	resultChans []chan *fossilizer.Result
	waitGroup   sync.WaitGroup
	transformer Transformer
	pending     *batch
}

// Transformer is the type of a function to transform results.
type Transformer func(evidence *Evidence, data, meta []byte) (*fossilizer.Result, error)

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
	}

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

// GetInfo implements github.com/stratumn/go/fossilizer.Adapter.GetInfo.
func (a *Fossilizer) GetInfo() (interface{}, error) {
	return &Info{
		Name:        Name,
		Description: Description,
		Version:     a.config.Version,
		Commit:      a.config.Commit,
	}, nil
}

// AddResultChan implements github.com/stratumn/go/fossilizer.Adapter.AddResultChan.
func (a *Fossilizer) AddResultChan(resultChan chan *fossilizer.Result) {
	a.resultChans = append(a.resultChans, resultChan)
}

// Fossilize implements github.com/stratumn/go/fossilizer.Adapter.Fossilize.
func (a *Fossilizer) Fossilize(data []byte, meta []byte) error {
	f := fossil{Meta: meta}
	copy(f.Data[:], data)
	a.fossilChan <- &f
	return <-a.resultChan
}

// Start starts the fossilizer.
func (a *Fossilizer) Start() error {
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
				log.WithField("interval", interval).Info("Requested new batch because the %s interval was reached")
			} else {
				log.WithField("interval", interval).Info("No batch is needed after the %s interval because there are no pending hashes")
			}
		case err := <-a.stopChan:
			e := a.stop(err)
			a.stopChan <- e
			return e
		}
	}
}

// Stop stops the fossilizer.
func (a *Fossilizer) Stop() {
	a.stopChan <- nil
	<-a.stopChan
}

// Started return a channel that will receive once the fossilizer has started.
func (a *Fossilizer) Started() <-chan struct{} {
	c := make(chan struct{}, 1)
	a.startedChan <- c
	return c
}

// SetTransformer sets a transformer.
func (a *Fossilizer) SetTransformer(t Transformer) {
	a.transformer = t
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

	a.waitGroup.Add(1)

	go func() {
		defer func() {
			a.waitGroup.Done()
			<-a.semChan
		}()

		a.semChan <- struct{}{}

		tree, err := merkle.NewStaticTree(b.data)
		if err != nil {
			a.stop(err)
			return
		}

		root := tree.Root()
		log.WithField("root", root).Info("Created tree with Merkle root")

		a.sendEvidence(tree, b.meta)
		log.WithField("root", root).Info("Sent evidence for batch with Merkle root")

		if b.file != nil {
			path := b.file.Name()

			if err := b.close(); err != nil {
				log.WithField("error", err).Warn("Failed to close batch file")
			}

			if a.config.Archive {
				archivePath := filepath.Join(a.config.Path, root.String())
				if err := os.Rename(path, archivePath); err == nil {
					log.WithFields(log.Fields{
						"old": filepath.Base(path),
						"new": filepath.Base(archivePath),
					}).Info("Renamed batch file")
				} else {
					log.WithFields(log.Fields{
						"old":   filepath.Base(path),
						"new":   filepath.Base(archivePath),
						"error": err,
					}).Warn("Failed to rename batch file")
				}
			} else {
				if err := os.Remove(path); err == nil {
					log.WithField("file", filepath.Base(path)).Info("Removed pending hashes file")
				} else {
					log.WithFields(log.Fields{
						"file":  filepath.Base(path),
						"error": err,
					}).Warn("Failed to remove batch file")
				}
			}
		}

		log.WithField("root", root).Info("Finished batch")
	}()
}

func (a *Fossilizer) sendEvidence(tree *merkle.StaticTree, meta [][]byte) {
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

		evidence := Evidence{
			Time: ts,
			Root: root,
			Path: tree.Path(i),
		}

		if a.transformer != nil {
			r, err = a.transformer(&evidence, d, m)
		} else {
			r = &fossilizer.Result{
				Evidence: &EvidenceWrapper{
					&evidence,
				},
				Data: d,
				Meta: m,
			}
		}

		if err == nil {
			for _, c := range a.resultChans {
				c <- r
			}
		} else {
			log.WithField("error", err).Error("Failed to transform evidence")
		}
	}
}

func (a *Fossilizer) stop(err error) error {
	if a.config.StopBatch {
		if len(a.pending.data) > 0 {
			a.batch(a.pending)
			log.Info("Requested final batch for pending hashes")
		} else {
			log.Info("No final batch is needed because there are no pending hashes")
		}
	}

	a.waitGroup.Wait()
	a.transformer = nil

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
