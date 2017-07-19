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
	"time"

	"github.com/stratumn/sdk/fossilizer"
)

const (
	// Name is the name set in the fossilizer's information.
	Name = "dummy"

	// Description is the description set in the fossilizer's information.
	Description = "Stratumn Dummy Fossilizer"
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
	config      *Config
	resultChans []chan *fossilizer.Result
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

// AddResultChan implements
// github.com/stratumn/sdk/fossilizer.Adapter.AddResultChan.
func (a *DummyFossilizer) AddResultChan(resultChan chan *fossilizer.Result) {
	a.resultChans = append(a.resultChans, resultChan)
}

// Fossilize implements github.com/stratumn/sdk/fossilizer.Adapter.Fossilize.
func (a *DummyFossilizer) Fossilize(data []byte, meta []byte) error {
	r := &fossilizer.Result{
		Evidence: map[string]interface{}{
			"authority": "dummy",
			"timestamp": time.Now().UTC().Format("20060102150405"),
		},
		Data: data,
		Meta: meta,
	}

	for _, c := range a.resultChans {
		c <- r
	}

	return nil
}
