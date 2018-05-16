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

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	"github.com/prometheus/common/log"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/validation/validators"
)

// LocalManager manages governance for validation rules management in an indigo network.
type LocalManager struct {
	*UpdateBroadcaster
	store *Store

	validationCfg    *Config
	validatorWatcher *fsnotify.Watcher

	current validators.Validator
}

// NewLocalManager enhances validator management with some governance concepts.
func NewLocalManager(ctx context.Context, a store.Adapter, validationCfg *Config) (Manager, error) {
	if validationCfg == nil {
		return nil, errors.New("missing configuration")
	}

	var err error
	var govMgr = LocalManager{
		UpdateBroadcaster: NewUpdateBroadcaster(),
		store:             NewStore(a, validationCfg),
		validationCfg:     validationCfg,
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
			m.Close()
			return ctx.Err()
		}
	}
}

// Current returns the current validator set
func (m *LocalManager) Current() validators.Validator {
	return m.current
}

// GetValidators returns the list of validators for each process by parsing a local file.
// The validators are updated in the store according to local changes.
func (m *LocalManager) GetValidators(ctx context.Context) (validators.ProcessesValidators, error) {
	processesValidators := make(validators.ProcessesValidators, 0)

	var updateStoreErr error
	if m.validationCfg.RulesPath != "" {
		_, loadConfigErr := LoadConfig(m.validationCfg, func(process string, schema *RulesSchema, validators validators.Validators) {
			newValidatorLink, err := m.store.LinkFromSchema(ctx, process, schema)
			if err != nil {
				log.Error("Could not create link from validation rules", err)
				return
			}
			updateStoreErr = m.store.UpdateValidator(ctx, newValidatorLink)
			if updateStoreErr != nil {
				log.Errorf("Could not update validation rules in store for process %s: %s", process, updateStoreErr)
				return
			}
			processesValidators[process] = validators
		})
		if loadConfigErr != nil {
			return nil, errors.Wrapf(loadConfigErr, "Cannot load validator rules file %s", m.validationCfg.RulesPath)
		}
		if updateStoreErr != nil {
			return nil, updateStoreErr
		}
	}
	return processesValidators, nil
}

func (m *LocalManager) updateCurrent(validatorsMap validators.ProcessesValidators) {
	m.current = validators.NewMultiValidator(validatorsMap)
	m.Broadcast(m.current)
}
