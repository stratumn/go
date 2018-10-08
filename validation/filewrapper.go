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
	"os"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-core/store"
	"github.com/stratumn/go-core/validation/validators"
)

// StoreWithConfigFile wraps a store adapter with a layer of validations
// based on a local configuration file.
type StoreWithConfigFile struct {
	store.Adapter

	defaultValidator validators.Validator

	lock            sync.RWMutex
	customValidator validators.Validator
}

// WrapStoreWithConfigFile wraps a store adapter with a layer of validations
// based on a local configuration file.
func WrapStoreWithConfigFile(a store.Adapter, cfg *Config) (store.Adapter, error) {
	// The default validator validates the structure of links and that the
	// chainscript graph stays coherent (no missing references for example).
	// This validator applies to all links, regardless of custom rules.
	defaultValidator := validators.NewMultiValidator([]validators.Validator{
		validators.NewRefsValidator(),
	})

	wrapped := &StoreWithConfigFile{
		Adapter:          a,
		defaultValidator: defaultValidator,
	}

	if cfg == nil || len(cfg.RulesPath) == 0 {
		log.Warn("No custom validation rules provided. Only default link validations will be applied.")
		return wrapped, nil
	}

	if _, err := os.Stat(cfg.RulesPath); os.IsNotExist(err) {
		log.Warnf("Invalid custom validation rules path: could not load rules at %s", cfg.RulesPath)
		return wrapped, nil
	}

	v, err := LoadFromFile(cfg)
	if err != nil {
		return nil, err
	}

	wrapped.customValidator = validators.NewMultiValidator(v.Flatten())

	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if err := w.Add(cfg.RulesPath); err != nil {
		return nil, errors.WithStack(err)
	}

	go wrapped.watchRules(w, cfg)

	return wrapped, nil
}

// CreateLink applies validations before creating the link.
func (a *StoreWithConfigFile) CreateLink(ctx context.Context, link *chainscript.Link) (chainscript.LinkHash, error) {
	if err := link.Validate(ctx); err != nil {
		return nil, err
	}

	if err := a.defaultValidator.Validate(ctx, a, link); err != nil {
		return nil, err
	}

	if err := a.validateCustom(ctx, link); err != nil {
		return nil, err
	}

	return a.Adapter.CreateLink(ctx, link)
}

func (a *StoreWithConfigFile) validateCustom(ctx context.Context, link *chainscript.Link) error {
	a.lock.RLock()
	defer a.lock.RUnlock()

	if a.customValidator == nil {
		return nil
	}

	return a.customValidator.Validate(ctx, a, link)
}

func (a *StoreWithConfigFile) watchRules(w *fsnotify.Watcher, cfg *Config) {
	for e := range w.Events {
		if e.Op != fsnotify.Write {
			continue
		}

		newValidators, err := LoadFromFile(cfg)
		if err != nil {
			continue
		}

		a.lock.Lock()
		a.customValidator = validators.NewMultiValidator(newValidators.Flatten())
		a.lock.Unlock()
	}
}
