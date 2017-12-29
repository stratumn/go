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

// Package dummyfossilizer implements a fossilizer that can be used for testing.
//
// It doesn't do much -- it just adds a timestamp.
package dummyfossilizer

import (
	"encoding/json"
	"time"

	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/fossilizer"
)

const (
	// Name is the name set in the fossilizer's information.
	Name = "dummy"

	// Description is the description set in the fossilizer's information.
	Description = "Indigo's Dummy Fossilizer"
)

// Config contains configuration options for the store.
type Config struct {
	// A version string that will be set in the store's information.
	Version string

	// A git commit hash that will be set in the store's information.
	Commit string
}

// Info is the info returned by GetInfo.
type Info struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Commit      string `json:"commit"`
}

// DummyFossilizer is the type that implements
// github.com/stratumn/sdk/fossilizer.Adapter.
type DummyFossilizer struct {
	config               *Config
	fossilizerEventChans []chan *fossilizer.Event
}

// DummyProof implements the cs.Proof interface
type DummyProof struct {
	Timestamp uint64 `json:"timestamp"`
}

// Time returns the timestamp from the block header
func (p *DummyProof) Time() uint64 {
	return p.Timestamp
}

// FullProof returns a JSON formatted proof
func (p *DummyProof) FullProof() []byte {
	bytes, err := json.MarshalIndent(p, "", "   ")
	if err != nil {
		return nil
	}
	return bytes
}

// Verify returns true if the proof of a given linkHash is correct
func (p *DummyProof) Verify(interface{}) bool {
	return true
}

// init needs to define a way to deserialize a DummyProof
func init() {
	cs.DeserializeMethods[Name] = func(rawProof json.RawMessage) (cs.Proof, error) {
		p := DummyProof{}
		if err := json.Unmarshal(rawProof, &p); err != nil {
			return nil, err
		}
		return &p, nil
	}
}

// New creates an instance of a DummyFossilizer.
func New(config *Config) *DummyFossilizer {
	return &DummyFossilizer{config, nil}
}

// GetInfo implements github.com/stratumn/sdk/fossilizer.Adapter.GetInfo.
func (a *DummyFossilizer) GetInfo() (interface{}, error) {
	return &Info{
		Name:        Name,
		Description: Description,
		Version:     a.config.Version,
		Commit:      a.config.Commit,
	}, nil
}

// AddFossilizerEventChan implements
// github.com/stratumn/sdk/fossilizer.Adapter.AddFossilizerEventChan.
func (a *DummyFossilizer) AddFossilizerEventChan(fossilizerEventChan chan *fossilizer.Event) {
	a.fossilizerEventChans = append(a.fossilizerEventChans, fossilizerEventChan)
}

// Fossilize implements github.com/stratumn/sdk/fossilizer.Adapter.Fossilize.
func (a *DummyFossilizer) Fossilize(data []byte, meta []byte) error {

	r := &fossilizer.Result{
		Evidence: cs.Evidence{
			Backend:  Name,
			Provider: Name,
			Proof: &DummyProof{
				Timestamp: uint64(time.Now().Unix()),
			},
		},
		Data: data,
		Meta: meta,
	}
	event := &fossilizer.Event{
		EventType: fossilizer.DidFossilizeLink,
		Data:      r,
	}

	for _, c := range a.fossilizerEventChans {
		c <- event
	}

	return nil
}
