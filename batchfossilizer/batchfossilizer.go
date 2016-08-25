// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// BUG(stephan): Optimize memory allocation.

// Package batchfossilizer implements a fossilizer that fossilize batches of data using a Merkle tree.
// The evidence will contain the Merkle root, the Merkle path, and a timestamp.
package batchfossilizer

import (
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/stratumn/go/fossilizer"
	"github.com/stratumn/goprivate/merkle"
	"github.com/stratumn/goprivate/types"
)

const (
	// Name is the name set in the fossilizer's information.
	Name = "batch"

	// Description is the description set in the fossilizer's information.
	Description = "Stratumn Batch Fossilizer"

	// DefaultInterval is the default interval between batches.
	DefaultInterval = time.Minute

	// DefaultMaxLeaves if the default maximum number of leaves of a Merkle tree.
	DefaultMaxLeaves = 32 * 1024

	// DefaultArchive is whether to archive completed batches by default.
	DefaultArchive = true

	// DefaultStopBatch is whether to do a batch on stop by default.
	DefaultStopBatch = true

	// DefaultFSync is whether to fsync after saving a hash to disk by default.
	DefaultFSync = false

	// PendingExt is the pending hashes filename extension.
	PendingExt = "pending"

	// FilePerm is the files's permissions.
	FilePerm = 0644
)

// Config contains configuration options for the fossilizer.
type Config struct {
	// A version string that will set in the store's information.
	Version string

	// A git commit sha that will set in the store's information.
	Commit string

	// Interval between batches.
	Interval time.Duration

	// Maximum number of leaves of a Merkle tree.
	MaxLeaves int

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

// Evidence is the evidence sent to the result channel.
type Evidence struct {
	Time int64         `json:"time"`
	Root types.Bytes32 `json:"merkleRoot"`
	Path merkle.Path   `json:"merklePath"`
}

// EvidenceWrapper wraps evidence with a namespace.
type EvidenceWrapper struct {
	Evidence *Evidence `json:"batch"`
}

type batch struct {
	leaves []types.Bytes32
	meta   [][]byte
	path   string
}

type chunk struct {
	Data []byte
	Meta []byte
}

// Fossilizer is the type that implements github.com/stratumn/go/fossilizer.Adapter.
type Fossilizer struct {
	config      *Config
	resultChans []chan *fossilizer.Result
	leaves      []types.Bytes32
	meta        [][]byte
	file        *os.File
	encoder     *gob.Encoder
	mutex       sync.Mutex
	waitGroup   sync.WaitGroup
	closeChan   chan error
}

// New creates an instance of a Fossilizer.
func New(config *Config) (*Fossilizer, error) {
	maxLeaves := config.MaxLeaves
	if maxLeaves == 0 {
		maxLeaves = DefaultMaxLeaves
	}

	a := &Fossilizer{
		config:    config,
		leaves:    make([]types.Bytes32, 0, maxLeaves),
		meta:      make([][]byte, 0, maxLeaves),
		closeChan: make(chan error),
	}

	if a.config.Path != "" {
		if err := a.recover(); err != nil {
			return nil, err
		}
	}

	return a, nil
}

// Start starts the fossilizer.
func (a *Fossilizer) Start() error {
	interval := a.config.Interval
	if interval == 0 {
		interval = DefaultInterval
	}

	for {
		select {
		case <-time.After(interval):
			a.mutex.Lock()
			if len(a.leaves) > 0 {
				if err := a.makeBatch(); err != nil {
					return err
				}
			}
			a.mutex.Unlock()
		case err := <-a.closeChan:
			return err
		}
	}
}

// Stop stops the fossilizer.
func (a *Fossilizer) Stop() error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	close(a.closeChan)
	a.closeChan = nil

	if a.config.StopBatch && len(a.leaves) > 0 {
		if err := a.makeBatch(); err != nil {
			return err
		}
	}

	a.waitGroup.Wait()

	if a.file != nil {
		return a.file.Close()
	}

	return nil
}

// GetInfo implements github.com/stratumn/go/fossilizer.Adapter.GetInfo.
func (a *Fossilizer) GetInfo() (interface{}, error) {
	return map[string]interface{}{
		"name":        Name,
		"description": Description,
		"version":     a.config.Version,
		"commit":      a.config.Commit,
	}, nil
}

// AddResultChan implements github.com/stratumn/go/fossilizer.Adapter.AddResultChan.
func (a *Fossilizer) AddResultChan(resultChan chan *fossilizer.Result) {
	a.resultChans = append(a.resultChans, resultChan)
}

// Fossilize implements github.com/stratumn/go/fossilizer.Adapter.Fossilize.
func (a *Fossilizer) Fossilize(data []byte, meta []byte) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if a.closeChan == nil {
		return errors.New("fossilizer is stopped")
	}

	if a.config.Path != "" {
		if a.file == nil {
			if err := a.open(); err != nil {
				return err
			}
		}

		if err := a.write(data, meta); err != nil {
			return err
		}
	}

	var leaf types.Bytes32
	copy(leaf[:], data)
	a.leaves = append(a.leaves, leaf)
	a.meta = append(a.meta, meta)

	maxLeaves := a.config.MaxLeaves
	if maxLeaves == 0 {
		maxLeaves = DefaultMaxLeaves
	}
	if len(a.leaves) >= maxLeaves {
		if err := a.makeBatch(); err != nil {
			a.closeChan <- err
			return err
		}
	}

	return nil
}

func (a *Fossilizer) batch(b batch) {
	defer a.waitGroup.Done()

	tree, err := merkle.NewStaticTree(b.leaves)

	if err != nil {
		a.closeChan <- err
		return
	}

	var (
		meta = b.meta
		ts   = time.Now().UTC().Unix()
		root = tree.Root()
	)

	for i := 0; i < tree.LeavesLen(); i++ {
		leaf := tree.Leaf(i)
		r := &fossilizer.Result{
			Evidence: &EvidenceWrapper{
				&Evidence{
					Time: ts,
					Root: root,
					Path: tree.Path(i),
				},
			},
			Data: leaf[:],
			Meta: meta[i],
		}

		for _, c := range a.resultChans {
			c <- r
		}
	}

	if b.path != "" {
		if a.config.Archive {
			path := filepath.Join(a.config.Path, root.String())
			if err := os.Rename(b.path, path); err != nil {
				log.Println(err)
			}
		} else {
			if err := os.Remove(b.path); err != nil {
				log.Println(err)
			}
		}
	}
}

func (a *Fossilizer) makeBatch() error {
	var path string

	if a.file != nil {
		path = a.file.Name()
		if err := a.file.Close(); err != nil {
			return err
		}
		a.file = nil
	}

	a.waitGroup.Add(1)
	go a.batch(batch{a.leaves, a.meta, path})

	maxLeaves := a.config.MaxLeaves
	if maxLeaves == 0 {
		maxLeaves = DefaultMaxLeaves
	}

	a.leaves, a.meta = make([]types.Bytes32, 0, maxLeaves), make([][]byte, 0, maxLeaves)

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

		dec := gob.NewDecoder(file)

		for {
			var c chunk

			if err := dec.Decode(&c); err != nil {
				if err == io.EOF {
					break
				}
				file.Close()
				return err
			}

			if err := a.Fossilize(c.Data, c.Meta); err != nil {
				file.Close()
				return err
			}
		}

		a.waitGroup.Wait()
		file.Close()

		if err := os.Remove(path); err != nil {
			return err
		}

		log.Printf("recovered %s\n", filepath.Base(path))
	}

	return nil
}

func (a *Fossilizer) open() error {
	filename := fmt.Sprintf("%d.%s", time.Now().UTC().UnixNano(), PendingExt)
	path := filepath.Join(a.config.Path, filename)
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_EXCL|os.O_CREATE, FilePerm)
	if err != nil {
		return err
	}

	a.file, a.encoder = file, gob.NewEncoder(file)

	return nil
}

func (a *Fossilizer) write(data []byte, meta []byte) error {
	if err := a.encoder.Encode(chunk{data, meta}); err != nil {
		return err
	}

	if a.config.FSync {
		if err := a.file.Sync(); err != nil {
			return err
		}
	}

	return nil
}
