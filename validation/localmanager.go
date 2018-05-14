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
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	"github.com/prometheus/common/log"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/validation/validators"
)

// LocalManager manages governance for validation rules management in an indigo network.
type LocalManager struct {
	store *Store

	validationCfg    *Config
	validatorWatcher *fsnotify.Watcher
	validators       map[string]validators.Validators
	current          validators.Validator

	listenersMutex sync.RWMutex
	listeners      []chan validators.Validator
}

// NewLocalManager enhances validator management with some governance concepts.
func NewLocalManager(ctx context.Context, a store.Adapter, validationCfg *Config) (Manager, error) {
	if validationCfg == nil {
		return nil, errors.New("missing configuration")
	}

	var err error
	var govMgr = LocalManager{
		store:         NewStore(a, validationCfg),
		validationCfg: validationCfg,
	}

	if validationCfg.RulesPath != "" {
		if govMgr.validatorWatcher, err = fsnotify.NewWatcher(); err != nil {
			return nil, errors.Wrap(err, "cannot create a new filesystem watcher for validators")
		}
		if err := govMgr.validatorWatcher.Add(validationCfg.RulesPath); err != nil {
			return nil, errors.Wrapf(err, "cannot watch validator configuration file %s", validationCfg.RulesPath)
		}
	}

	_, err = govMgr.GetValidators(ctx)

	if validators, _ := govMgr.store.GetValidators(ctx); len(validators) > 0 {
		govMgr.updateCurrent(validators)
	}

	return &govMgr, err
}

// ListenAndUpdate will update the current validators whenever the provided rule file is updated.
// This method must be run in a goroutine as it will wait for write events on the file.
func (m *LocalManager) ListenAndUpdate(ctx context.Context) error {
	if m.validatorWatcher == nil {
		return ErrNoFileWatcher
	}

	for {
		select {
		case event := <-m.validatorWatcher.Events:
			if event.Op&fsnotify.Write == fsnotify.Write && event.Name != "" {
				if validators, err := m.GetValidators(ctx); err == nil {
					m.updateCurrent(validators)
				}
			}

		case err := <-m.validatorWatcher.Errors:
			log.Warnf("Validator file watcher error caught: %s", err)

		case <-ctx.Done():
			m.listenersMutex.Lock()
			defer m.listenersMutex.Unlock()
			for _, s := range m.listeners {
				close(s)
			}
			return ctx.Err()
		}
	}
}

// Current returns the current validator set
func (m *LocalManager) Current() validators.Validator {
	return m.current
}

// AddListener return a listener that will be notified when the validator changes.
func (m *LocalManager) AddListener() <-chan validators.Validator {
	m.listenersMutex.Lock()
	defer m.listenersMutex.Unlock()

	subscribeChan := make(chan validators.Validator)
	m.listeners = append(m.listeners, subscribeChan)

	// Insert the current validator in the channel if there is one.
	if m.current != nil {
		go func() {
			subscribeChan <- m.current
		}()
	}
	return subscribeChan
}

// RemoveListener removes a listener.
func (m *LocalManager) RemoveListener(c <-chan validators.Validator) {
	m.listenersMutex.Lock()
	defer m.listenersMutex.Unlock()

	index := -1
	for i, l := range m.listeners {
		if l == c {
			index = i
			break
		}
	}

	if index >= 0 {
		close(m.listeners[index])
		m.listeners[index] = m.listeners[len(m.listeners)-1]
		m.listeners = m.listeners[:len(m.listeners)-1]
	}
}

// GetValidators returns the list of validators for each process by parsing a local file.
// The validators are updated in the store according to local changes.
func (m *LocalManager) GetValidators(ctx context.Context) (processesValidators []validators.Validators, err error) {
	if m.validationCfg.RulesPath != "" {
		_, err = LoadConfig(m.validationCfg, func(process string, schema RulesSchema, validators validators.Validators) {
			m.store.UpdateValidator(ctx, process, schema)
			processesValidators = append(processesValidators, validators)
		})
		if err != nil {
			return nil, errors.Wrapf(err, "Cannot load validator rules file %s", m.validationCfg.RulesPath)

		}
	}
	return processesValidators, err
}

func (m *LocalManager) updateCurrent(validatorsList []validators.Validators) {
	m.listenersMutex.RLock()
	defer m.listenersMutex.RUnlock()

	v4ch := make(validators.Validators, 0)
	for _, v := range validatorsList {
		v4ch = append(v4ch, v...)
	}
	m.current = validators.NewMultiValidator(v4ch)
	for _, listener := range m.listeners {
		go func(listener chan validators.Validator) {
			listener <- m.current
		}(listener)
	}
}
