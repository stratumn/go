// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package batchfossilizer implements a fossilizer that fossilize batches of data using a Merkle tree.
// The evidence will contain the Merkle root, the Merkle path, and a timestamp.
package batchfossilizer

// TODO: save pending leaves to file and recover them on start
// TODO: optimize memory allocation

import (
	"log"
	"sync"
	"time"

	"github.com/stratumn/go/fossilizer"
	"github.com/stratumn/goprivate/merkle"
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
)

// Config contains configuration options for the fossilizer.
type Config struct {
	// A version string that will set in the store's information.
	Version string

	// Interval between batches.
	Interval time.Duration

	// Maximum number of leaves of a Merkle tree.
	MaxLeaves int
}

type batch struct {
	leaves []merkle.Hash
	meta   [][]byte
}

// BatchFossilizer is the type that implements github.com/stratumn/go/fossilizer.Adapter.
type BatchFossilizer struct {
	config      *Config
	resultChans []chan *fossilizer.Result
	leaves      []merkle.Hash
	meta        [][]byte
	mutex       sync.Mutex
	closeChan   chan struct{}
}

// New creates an instance of a BatchFossilizer.
func New(config *Config) *BatchFossilizer {
	maxLeaves := config.MaxLeaves
	if maxLeaves == 0 {
		maxLeaves = DefaultMaxLeaves
	}

	a := &BatchFossilizer{
		config:    config,
		leaves:    make([]merkle.Hash, 0, maxLeaves),
		meta:      make([][]byte, 0, maxLeaves),
		closeChan: make(chan struct{}),
	}

	return a
}

// Start starts the fossilizer.
func (a *BatchFossilizer) Start() {
	interval := a.config.Interval
	if interval == 0 {
		interval = DefaultInterval
	}

	for {
		select {
		case <-time.After(interval):
			a.mutex.Lock()
			if len(a.leaves) > 0 {
				maxLeaves := a.config.MaxLeaves
				if maxLeaves == 0 {
					maxLeaves = DefaultMaxLeaves
				}
				go a.batch(batch{a.leaves, a.meta})
				a.leaves, a.meta = make([]merkle.Hash, 0, maxLeaves), make([][]byte, 0, maxLeaves)
			}
			a.mutex.Unlock()
		case <-a.closeChan:
			return
		}
	}
}

// Stop stops the fossilizer.
func (a *BatchFossilizer) Stop() {
	a.closeChan <- struct{}{}
}

// GetInfo implements github.com/stratumn/go/fossilizer.Adapter.GetInfo.
func (a *BatchFossilizer) GetInfo() (interface{}, error) {
	return map[string]interface{}{
		"name":        Name,
		"description": Description,
		"version":     a.config.Version,
	}, nil
}

// AddResultChan implements github.com/stratumn/go/fossilizer.Adapter.AddResultChan.
func (a *BatchFossilizer) AddResultChan(resultChan chan *fossilizer.Result) {
	a.resultChans = append(a.resultChans, resultChan)
}

// Fossilize implements github.com/stratumn/go/fossilizer.Adapter.Fossilize.
func (a *BatchFossilizer) Fossilize(data []byte, meta []byte) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	var leaf merkle.Hash
	copy(leaf[:], data)
	a.leaves = append(a.leaves, leaf)
	a.meta = append(a.meta, meta)

	maxLeaves := a.config.MaxLeaves
	if maxLeaves == 0 {
		maxLeaves = DefaultMaxLeaves
	}
	if len(a.leaves) >= maxLeaves {
		go a.batch(batch{a.leaves, a.meta})
		a.leaves, a.meta = make([]merkle.Hash, 0, maxLeaves), make([][]byte, 0, maxLeaves)
	}

	return nil
}

type evidence struct {
	Time int64       `json:"time"`
	Root merkle.Hash `json:"merkleRoot"`
	Path merkle.Path `json:"merklePath"`
}

type evidenceWrapper struct {
	Evidence evidence `json:"batch"`
}

func (a *BatchFossilizer) batch(b batch) {
	tree, err := merkle.NewStaticTree(b.leaves)

	// TODO: handle error properly
	if err != nil {
		log.Println(err)
		return
	}

	var (
		meta = b.meta
		ts   = time.Now().UTC().Unix()
		root = tree.Root()
	)

	for i := 0; i < tree.NumLeaves(); i++ {
		leaf := tree.Leaf(i)
		r := &fossilizer.Result{
			Evidence: evidenceWrapper{
				evidence{
					Time: ts,
					Root: root,
					Path: tree.Path(i),
				},
			},
			Data: leaf[:],
			Meta: meta[i],
		}

		go func(chans []chan *fossilizer.Result) {
			for _, c := range chans {
				c <- r
			}
		}(a.resultChans)
	}
}
