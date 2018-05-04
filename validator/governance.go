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

package validator

import (
	"bytes"
	"context"
	"encoding/json"
	"sync"

	"github.com/fsnotify/fsnotify"
	cj "github.com/gibson042/canonicaljson-go"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"

	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/store"
)

const (
	// GovernanceProcessName is the process name used for governance information storage
	governanceProcessName = "_governance"

	// ValidatorTag is the tag used to find validators in storage
	validatorTag = "validators"
)

var (
	// ErrNoFileWatcher is the error returned when the provided rules file could not be watched.
	ErrNoFileWatcher = errors.New("cannot listen for file updates: no file watcher")

	defaultPagination = store.Pagination{
		Offset: 0,
		Limit:  1, // store.DefaultLimit,
	}
)

// GovernanceManager defines the methods to implement to manage validations in an indigo network.
type GovernanceManager interface {

	// ListenAndUpdate will update the current validators whenever a change occurs in the governance rules.
	// This method must be run in a goroutine as it will wait for events from the network or file updates.
	ListenAndUpdate(ctx context.Context) error

	// AddListener adds a listener for validator changes.
	AddListener() <-chan Validator

	// RemoveListener removes a listener.
	RemoveListener(<-chan Validator)

	// Current returns the current version of the validator set.
	Current() Validator
}

// LocalGovernor manages governance for validation rules management in an indigo network.
type LocalGovernor struct {
	adapter store.Adapter

	validationCfg    *Config
	validatorWatcher *fsnotify.Watcher
	validators       map[string][]Validator
	current          Validator

	listenersMutex sync.RWMutex
	listeners      []chan Validator
}

// NewLocalGovernor enhances validator management with some governance concepts.
func NewLocalGovernor(ctx context.Context, a store.Adapter, validationCfg *Config) (GovernanceManager, error) {
	var err error
	var govMgr = LocalGovernor{
		adapter:       a,
		validators:    make(map[string][]Validator, 0),
		validationCfg: validationCfg,
	}

	govMgr.loadValidatorsFromFile(ctx)
	govMgr.loadValidatorsFromStore(ctx)
	if len(govMgr.validators) > 0 {
		govMgr.updateCurrent()
	}
	if validationCfg != nil && validationCfg.RulesPath != "" {
		if govMgr.validatorWatcher, err = fsnotify.NewWatcher(); err != nil {
			return nil, errors.Wrap(err, "cannot create a new filesystem watcher for validators")
		}
		if err := govMgr.validatorWatcher.Add(validationCfg.RulesPath); err != nil {
			return nil, errors.Wrapf(err, "cannot watch validator configuration file %s", validationCfg.RulesPath)
		}
	}

	return &govMgr, nil
}

// ListenAndUpdate will update the current validators whenever the provided rule file is updated.
// This method must be run in a goroutine as it will wait for write events on the file.
func (m *LocalGovernor) ListenAndUpdate(ctx context.Context) error {
	if m.validatorWatcher == nil {
		return ErrNoFileWatcher
	}

	for {
		select {
		case event := <-m.validatorWatcher.Events:
			if event.Op&fsnotify.Write == fsnotify.Write && event.Name != "" {
				if m.loadValidatorsFromFile(ctx) == nil {
					m.updateCurrent()
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
func (m *LocalGovernor) Current() Validator {
	return m.current
}

// AddListener return a listener that will be notified when the validator changes.
func (m *LocalGovernor) AddListener() <-chan Validator {
	m.listenersMutex.Lock()
	defer m.listenersMutex.Unlock()

	subscribeChan := make(chan Validator)
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
func (m *LocalGovernor) RemoveListener(c <-chan Validator) {
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

func (m *LocalGovernor) loadValidatorsFromFile(ctx context.Context) (err error) {
	if m.validationCfg != nil && m.validationCfg.RulesPath != "" {
		_, err = LoadConfig(m.validationCfg, func(process string, schema rulesSchema, validators []Validator) {
			m.validators[process] = m.updateValidatorInStore(ctx, process, schema, validators)
		})
		if err != nil {
			log.Errorf("Cannot load validator rules file %s: %s", m.validationCfg.RulesPath, err)
		}
	}
	return err
}

func (m *LocalGovernor) loadValidatorsFromStore(ctx context.Context) {
	for _, process := range m.getAllProcesses(ctx) {
		if _, exist := m.validators[process]; !exist {
			m.validators[process] = m.getValidators(ctx, process)
		}
	}
}

func (m *LocalGovernor) updateCurrent() {
	m.listenersMutex.RLock()
	defer m.listenersMutex.RUnlock()

	v4ch := make([]Validator, 0)
	for _, v := range m.validators {
		v4ch = append(v4ch, v...)
	}
	m.current = NewMultiValidator(v4ch)
	for _, listener := range m.listeners {
		go func(listener chan Validator) {
			listener <- m.current
		}(listener)
	}
}

func (m *LocalGovernor) getAllProcesses(ctx context.Context) []string {
	processSet := make(map[string]interface{}, 0)
	for offset := 0; offset >= 0; {
		segments, err := m.adapter.FindSegments(ctx, &store.SegmentFilter{
			Pagination: store.Pagination{Offset: offset, Limit: store.MaxLimit},
			Process:    governanceProcessName,
			Tags:       []string{validatorTag},
		})
		if err != nil {
			log.Errorf("Cannot retrieve governance segments: %+v", errors.WithStack(err))
			return []string{}
		}
		for _, segment := range segments {
			for _, tag := range segment.Link.Meta.Tags {
				if tag != validatorTag {
					processSet[tag] = nil
				}
			}
		}
		if len(segments) == store.MaxLimit {
			offset += store.MaxLimit
		} else {
			break
		}
	}
	ret := make([]string, 0)
	for p := range processSet {
		ret = append(ret, p)
	}
	return ret
}

func (m *LocalGovernor) getValidators(ctx context.Context, process string) []Validator {
	segments, err := m.adapter.FindSegments(ctx, &store.SegmentFilter{
		Pagination: defaultPagination,
		Process:    governanceProcessName,
		Tags:       []string{process, validatorTag},
	})
	if err != nil || len(segments) == 0 {
		return nil
	}
	linkState := segments[0].Link.State
	pki, ok := linkState["pki"]
	types, ok2 := linkState["types"]
	if !ok || !ok2 {
		return nil
	}
	rawPKI, ok := pki.(json.RawMessage)
	rawTypes, ok2 := types.(json.RawMessage)
	if !ok || !ok2 {
		return nil
	}
	v, err := LoadProcessRules(processesRules{
		"process": rulesSchema{
			PKI:   rawPKI,
			Types: rawTypes,
		},
	}, m.validationCfg.PluginsPath, nil)
	if err != nil {
		return v
	}
	return nil
}

func (m *LocalGovernor) updateValidatorInStore(ctx context.Context, process string, schema rulesSchema, validators []Validator) []Validator {
	segments, err := m.adapter.FindSegments(ctx, &store.SegmentFilter{
		Pagination: defaultPagination,
		Process:    governanceProcessName,
		Tags:       []string{process, validatorTag},
	})
	if err != nil {
		log.Errorf("Cannot retrieve governance segments: %+v", errors.WithStack(err))
		return validators
	}
	if len(segments) == 0 {
		log.Warnf("No governance segments found for process %s", process)
		if err = m.uploadValidator(ctx, process, schema, nil); err != nil {
			log.Warnf("Cannot upload validator: %s", err)
		}
		return validators
	}
	link := segments[0].Link
	if m.compareFromStore(link.State, "pki", schema.PKI) != nil ||
		m.compareFromStore(link.State, "types", schema.Types) != nil {
		log.Infof("Validator or process %s has to be updated in store", process)
		if err = m.uploadValidator(ctx, process, schema, &link); err != nil {
			log.Warnf("Cannot upload validator: %s", err)
		}
	}

	return validators
}

func getCanonicalJSONFromData(rawData json.RawMessage) (json.RawMessage, error) {
	var typedData interface{}
	err := json.Unmarshal(rawData, &typedData)
	if err != nil {
		return nil, err
	}
	return cj.Marshal(typedData)
}

func (m *LocalGovernor) compareFromStore(meta map[string]interface{}, key string, fileData json.RawMessage) error {
	metaData, ok := meta[key]
	if !ok {
		return errors.Errorf("%s is missing on segment", key)
	}
	canonStoreData, err := cj.Marshal(metaData)
	if err != nil {
		return errors.Wrapf(err, "cannot canonical marshal %s store data", key)
	}
	canonFileData, err := getCanonicalJSONFromData(fileData)
	if err != nil {
		return errors.Wrapf(err, "cannot canonical marshal %s file data", key)
	}
	if !bytes.Equal(canonStoreData, canonFileData) {
		return errors.New("data different from file and from store")
	}
	return nil
}

func (m *LocalGovernor) uploadValidator(ctx context.Context, process string, schema rulesSchema, prevLink *cs.Link) error {
	priority := 0.
	mapID := ""
	prevLinkHash := ""
	if prevLink != nil {
		priority = prevLink.Meta.Priority + 1.
		mapID = prevLink.Meta.MapID
		var err error
		if prevLinkHash, err = prevLink.HashString(); err != nil {
			return errors.Wrapf(err, "cannot get previous hash for process governance %s", process)
		}
	} else {
		mapID = uuid.NewV4().String()
	}
	linkState := map[string]interface{}{
		"pki":   schema.PKI,
		"types": schema.Types,
	}
	linkMeta := cs.LinkMeta{
		Process:      governanceProcessName,
		MapID:        mapID,
		PrevLinkHash: prevLinkHash,
		Priority:     priority,
		Tags:         []string{process, validatorTag},
	}

	link := &cs.Link{
		State:      linkState,
		Meta:       linkMeta,
		Signatures: cs.Signatures{},
	}

	hash, err := m.adapter.CreateLink(ctx, link)
	if err != nil {
		return errors.Wrapf(err, "cannot create link for process governance %s", process)
	}
	log.Infof("New validator rules store for process %s: %q", process, hash)
	return nil
}
