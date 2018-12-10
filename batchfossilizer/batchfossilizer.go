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

// Package batchfossilizer implements a fossilizer that fossilizes batches of
// data using a merkle tree.
// The evidence will contain the merkle root, the merkle path, and a timestamp.
package batchfossilizer

import (
	"context"
	"encoding/hex"
	"sync"
	"time"

	"github.com/stratumn/go-core/batchfossilizer/evidences"
	"github.com/stratumn/go-core/fossilizer"
	"github.com/stratumn/go-core/monitoring"
	"github.com/stratumn/merkle"

	"go.opencensus.io/trace"
)

const (
	// Name is the name set in the fossilizer's information.
	Name = "batchfossilizer"

	// Description is the description set in the fossilizer's information.
	Description = "Stratumn Batch Fossilizer"
)

// Info is the info returned by GetInfo.
type Info struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Commit      string `json:"commit"`
}

// PendingProof is used to store an incomplete proof (waiting for the proof of
// the wrapped fossilizer).
type PendingProof struct {
	fossil *fossilizer.Fossil
	proof  *evidences.BatchProof
}

// Fossilizer is the type that
// implements github.com/stratumn/go-core/fossilizer.Adapter.
type Fossilizer struct {
	config *Config
	ticker *time.Ticker

	foss  fossilizer.Adapter
	queue fossilizer.FossilsQueue

	eventChansLock sync.RWMutex
	eventChans     []chan *fossilizer.Event

	pendingProofsLock sync.Mutex
	pendingProofs     map[string][]*PendingProof
}

// New creates an instance of a batch Fossilizer by wrapping an existing
// fossilizer.
// You should cancel the input context to properly free internal go routines
// when you don't need the fossilizer.
func New(ctx context.Context, config *Config, f fossilizer.Adapter, q fossilizer.FossilsQueue) fossilizer.Adapter {
	t := time.NewTicker(config.GetInterval())
	a := &Fossilizer{
		config:        config,
		ticker:        t,
		foss:          f,
		queue:         q,
		pendingProofs: make(map[string][]*PendingProof),
	}

	// The wrapped fossilizer will be the one doing the actual fossilization
	// and producing a blockchain proof.
	// So we need to listen to its events and re-trigger our own events on top.
	fChan := make(chan *fossilizer.Event)
	f.AddFossilizerEventChan(fChan)
	go a.eventLoop(ctx, fChan)

	go a.fossilizeLoop(ctx)

	return a
}

// GetInfo returns the fossilizer info.
func (a *Fossilizer) GetInfo(ctx context.Context) (interface{}, error) {
	return &Info{
		Name:        Name,
		Description: Description,
		Version:     a.config.Version,
		Commit:      a.config.Commit,
	}, nil
}

// AddFossilizerEventChan forwards to the underlying fossilizer.
func (a *Fossilizer) AddFossilizerEventChan(fossilizerEventChan chan *fossilizer.Event) {
	a.eventChansLock.Lock()
	defer a.eventChansLock.Unlock()

	a.eventChans = append(a.eventChans, fossilizerEventChan)
}

// Fossilize adds to the fossilizing queue.
func (a *Fossilizer) Fossilize(ctx context.Context, data []byte, meta []byte) error {
	return a.queue.Push(ctx, &fossilizer.Fossil{
		Data: data,
		Meta: meta,
	})
}

func (a *Fossilizer) fossilizeLoop(ctx context.Context) {
	for {
		select {
		case <-a.ticker.C:
			a.fossilizeBatch(context.Background())
		case <-ctx.Done():
			a.ticker.Stop()
			return
		}
	}
}

func (a *Fossilizer) fossilizeBatch(ctx context.Context) {
	ctx, span := trace.StartSpan(ctx, "batchfossilizer/fossilizeBatch")
	defer span.End()

	batchSize := a.config.GetMaxLeaves()

	for i := 0; i < a.config.GetMaxSimBatches(); i++ {
		fossils, err := a.queue.Pop(ctx, batchSize)
		if err != nil {
			monitoring.SetSpanStatus(span, err)
			return
		}

		if len(fossils) == 0 {
			return
		}

		leaves := make([][]byte, len(fossils))
		for i := 0; i < len(fossils); i++ {
			leaves[i] = fossils[i].Data
		}

		tree, err := merkle.NewStaticTree(leaves)
		if err != nil {
			monitoring.SetSpanStatus(span, err)
			return
		}

		root := tree.Root()
		a.addPendingProofs(fossils, root, tree)

		err = a.foss.Fossilize(ctx, root, nil)
		if err != nil {
			monitoring.SetSpanStatus(span, err)
			return
		}

		// If the queue is empty, early return instead of polling the queue.
		if len(fossils) < batchSize {
			return
		}
	}
}

// addPendingProofs pre-fills proofs for all the fossils that are pending.
// Once the wrapped fossilizer confirms fossilization of the merkle tree, we
// can create fossilization events for all those pending proofs.
func (a *Fossilizer) addPendingProofs(fossils []*fossilizer.Fossil, root []byte, tree *merkle.StaticTree) {
	a.pendingProofsLock.Lock()
	defer a.pendingProofsLock.Unlock()

	pendingProofs := make([]*PendingProof, len(fossils))
	for i := 0; i < len(fossils); i++ {
		pendingProofs[i] = &PendingProof{
			fossil: fossils[i],
			proof: &evidences.BatchProof{
				Root: root,
				Path: tree.Path(i),
			},
		}
	}

	key := hex.EncodeToString(root)
	a.pendingProofs[key] = pendingProofs
}

func (a *Fossilizer) eventLoop(ctx context.Context, fChan chan *fossilizer.Event) {
	for {
		select {
		case e := <-fChan:
			a.eventBatch(context.Background(), e)
		case <-ctx.Done():
			close(fChan)
			return
		}
	}
}

// eventBatch transforms a single fossilization event of a merkle root into
// individual fossilization events for each fossil included in the merkle tree.
// It then sends these events to all registered listeners.
func (a *Fossilizer) eventBatch(ctx context.Context, e *fossilizer.Event) {
	ctx, span := trace.StartSpan(ctx, "batchfossilizer/eventBatch")
	defer span.End()

	if e.EventType != fossilizer.DidFossilize {
		return
	}

	r := e.Data.(*fossilizer.Result)
	key := hex.EncodeToString(r.Data)

	a.pendingProofsLock.Lock()
	defer a.pendingProofsLock.Unlock()

	pendingProofs, ok := a.pendingProofs[key]
	if !ok {
		span.Annotatef(nil, "pending proofs not found for root %s", key)
		return
	}

	a.eventChansLock.RLock()
	defer a.eventChansLock.RUnlock()

	for _, p := range pendingProofs {
		// Wrap the underlying (blockchain) proof.
		p.proof.Timestamp = time.Now().Unix()
		p.proof.Proof = r.Evidence.Proof
		ev, err := p.proof.Evidence(Name)
		if err != nil {
			span.Annotate(
				[]trace.Attribute{trace.StringAttribute("error", err.Error())},
				"could not create evidence",
			)
			continue
		}

		for _, l := range a.eventChans {
			go func(l chan *fossilizer.Event) {
				l <- &fossilizer.Event{
					EventType: fossilizer.DidFossilize,
					Data: &fossilizer.Result{
						Fossil:   *p.fossil,
						Evidence: *ev,
					},
				}
			}(l)
		}
	}

	delete(a.pendingProofs, key)
}
