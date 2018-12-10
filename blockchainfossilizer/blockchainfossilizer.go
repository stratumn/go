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

// Package blockchainfossilizer implements a fossilizer that fossilize hashes
// on a blockchain.
// A blockchain fossilizer should usually be wrapped inside a batch fossilizer
// to use less blockchain transactions.
package blockchainfossilizer

import (
	"context"
	"fmt"
	"sync"

	"github.com/stratumn/go-core/blockchain"
	"github.com/stratumn/go-core/blockchainfossilizer/evidences"
	"github.com/stratumn/go-core/fossilizer"
	"github.com/stratumn/go-core/monitoring"

	"go.opencensus.io/trace"
)

const (
	// Name is the name set in the fossilizer's information.
	Name = "blockchainfossilizer"

	// Description is the description set in the fossilizer's information.
	Description = "Stratumn's Blockchain Fossilizer"
)

// Config for the blockchain fossilizer.
type Config struct {
	Version     string
	Commit      string
	Timestamper blockchain.HashTimestamper
}

// Info is the info returned by GetInfo.
type Info struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Commit      string `json:"commit"`
	Blockchain  string `json:"blockchain"`
}

// Fossilizer fossilizes data on a blockchain.
type Fossilizer struct {
	config *Config

	eventChansLock sync.RWMutex
	eventChans     []chan *fossilizer.Event
}

// New creates an instance of a blockchain fossilizer.
func New(config *Config) fossilizer.Adapter {
	return &Fossilizer{
		config: config,
	}
}

// GetInfo implements github.com/stratumn/go-core/fossilizer.Adapter.GetInfo.
func (a *Fossilizer) GetInfo(ctx context.Context) (interface{}, error) {
	timestamperInfo := a.config.Timestamper.GetInfo()

	return &Info{
		Name:        Name,
		Description: fmt.Sprintf("%s with %s", Description, timestamperInfo.Description),
		Version:     a.config.Version,
		Commit:      a.config.Commit,
		Blockchain:  timestamperInfo.Network.String(),
	}, nil
}

// AddFossilizerEventChan adds a new listener.
func (a *Fossilizer) AddFossilizerEventChan(fossilizerEventChan chan *fossilizer.Event) {
	a.eventChansLock.Lock()
	defer a.eventChansLock.Unlock()

	a.eventChans = append(a.eventChans, fossilizerEventChan)
}

// Fossilize data to the configured blockchain.
func (a *Fossilizer) Fossilize(ctx context.Context, data []byte, meta []byte) (err error) {
	ctx, span := trace.StartSpan(ctx, "blockchainfossilizer/Fossilize")
	defer func() {
		monitoring.SetSpanStatusAndEnd(span, err)
	}()

	txid, err := a.config.Timestamper.TimestampHash(ctx, data)
	if err != nil {
		return err
	}

	timestamperInfo := a.config.Timestamper.GetInfo()
	proof := evidences.New(data, txid)
	evidence, err := proof.Evidence(timestamperInfo.Network.String())
	if err != nil {
		return err
	}

	a.eventChansLock.RLock()
	defer a.eventChansLock.RUnlock()

	for _, l := range a.eventChans {
		go func(l chan *fossilizer.Event) {
			l <- &fossilizer.Event{
				EventType: fossilizer.DidFossilize,
				Data: &fossilizer.Result{
					Fossil: fossilizer.Fossil{
						Data: data,
						Meta: meta,
					},
					Evidence: *evidence,
				},
			}
		}(l)
	}

	return nil
}
