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

package validation

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/pkg/errors"
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/validation/validators"
)

// NetworkManager manages governance for validation rules management in an indigo network.
type NetworkManager struct {
	store *Store

	validationCfg *Config
	current       validators.Validator

	networkListener <-chan *cs.Link
	listenersMutex  sync.RWMutex
	listeners       []chan validators.Validator
}

// NewNetworkManager returns a new NetworManager able to listen to the network and update governance rules.
func NewNetworkManager(ctx context.Context, a store.Adapter, networkListener <-chan *cs.Link, validationCfg *Config) (Manager, error) {
	var err error
	var govMgr = NetworkManager{
		store:           NewStore(a, validationCfg),
		networkListener: networkListener,
		validationCfg:   validationCfg,
	}

	currentValidators, err := govMgr.store.GetValidators(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "could not initialize network governor")
	}
	if len(currentValidators) > 0 {
		govMgr.updateCurrent(currentValidators)
	}

	return &govMgr, nil
}

func isGovernanceLink(link *cs.Link) bool {
	return true
}

// ListenAndUpdate will update the current validators whenever the provided rule file is updated.
// This method must be run in a goroutine as it will wait for write events on the file.
func (g *NetworkManager) ListenAndUpdate(ctx context.Context) error {
	for {
		select {
		case link := <-g.networkListener:
			if isGovernanceLink(link) {
				if validators, err := g.GetValidators(ctx, link); err == nil {
					g.updateCurrent(validators)
				}
			}

		case <-ctx.Done():
			g.listenersMutex.Lock()
			defer g.listenersMutex.Unlock()
			for _, s := range g.listeners {
				close(s)
			}
			return ctx.Err()
		}
	}
}

// AddListener return a listener that will be notified when the validator changes.
func (g *NetworkManager) AddListener() <-chan validators.Validator {
	g.listenersMutex.Lock()
	defer g.listenersMutex.Unlock()

	subscribeChan := make(chan validators.Validator)
	g.listeners = append(g.listeners, subscribeChan)

	// Insert the current validator in the channel if there is one.
	if g.current != nil {
		go func() {
			subscribeChan <- g.current
		}()
	}
	return subscribeChan
}

// GetValidators extract the config from a link, parses it and returns as list of validators.
func (g *NetworkManager) GetValidators(ctx context.Context, link *cs.Link) (processesValidators validators.ProcessesValidators, err error) {
	jsonRules, err := json.Marshal(link.State)
	if err != nil {
		return nil, err
	}

	processRulesUpdate := func(process string, schema RulesSchema, validators validators.Validators) {
		g.store.UpdateValidator(ctx, process, schema)
		processesValidators[process] = validators
	}
	if _, err = LoadConfigContent(jsonRules, g.validationCfg.PluginsPath, processRulesUpdate); err != nil {
		return nil, err
	}

	return processesValidators, nil
}

// RemoveListener removes a listener.
func (g *NetworkManager) RemoveListener(c <-chan validators.Validator) {
	g.listenersMutex.Lock()
	defer g.listenersMutex.Unlock()

	index := -1
	for i, l := range g.listeners {
		if l == c {
			index = i
			break
		}
	}

	if index >= 0 {
		close(g.listeners[index])
		g.listeners[index] = g.listeners[len(g.listeners)-1]
		g.listeners = g.listeners[:len(g.listeners)-1]
	}
}

// Current returns the current validator set
func (g *NetworkManager) Current() validators.Validator {
	return g.current
}

func (g *NetworkManager) updateCurrent(validatorsMap validators.ProcessesValidators) {
	g.listenersMutex.RLock()
	defer g.listenersMutex.RUnlock()

	g.current = validators.NewMultiValidator(validatorsMap)

	for _, listener := range g.listeners {
		go func(listener chan validators.Validator) {
			listener <- g.current
		}(listener)
	}
}
