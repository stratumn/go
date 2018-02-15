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
	"encoding/json"
	"reflect"

	"github.com/fsnotify/fsnotify"
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

var defaultPagination = store.Pagination{
	Offset: 0,
	Limit:  1, // store.DefaultLimit,
}

// GovernanceManager manages governance for validation rules management.
type GovernanceManager struct {
	adapter store.Adapter

	validatorWatcher  *fsnotify.Watcher
	validatorChan     chan Validator
	processLastHeight map[string]int
}

// NewGovernanceManager enhances validator management with some governance concepts.
func NewGovernanceManager(a store.Adapter, filename string) (*GovernanceManager, error) {
	var govMgr = GovernanceManager{
		adapter:       a,
		validatorChan: make(chan Validator, 1),
	}
	if filename != "" {
		err := govMgr.loadValidatorFile(filename, false)
		if err != nil {
			log.Warn(errors.Wrapf(err, "cannot load validator configuration file %s", filename))
		}
		if govMgr.validatorWatcher, err = fsnotify.NewWatcher(); err != nil {
			return nil, errors.Wrap(err, "cannot create a new filesystem watcher for validators")
		}
		if err = govMgr.validatorWatcher.Add(filename); err != nil {
			return nil, errors.Wrapf(err, "cannot watch validator configuration file %s", filename)
		}
	}

	return &govMgr, nil
}

func (m *GovernanceManager) loadValidatorFile(filename string, updateStore bool) error {
	var v4ch = make([]Validator, 0)
	_, err := LoadConfig(filename, func(process string, schema rulesSchema, validators []Validator) {
		log.Infof("Here is a new validator: %s", process)
		v := m.getValidatorFromStore(process, schema, validators, updateStore)
		if v != nil {
			v4ch = append(v4ch, v...)
		}
	})
	if err == nil {
		select {
		case m.validatorChan <- NewMultiValidator(v4ch):
		}
	}
	return err
}

func (m *GovernanceManager) getValidatorFromStore(process string, schema rulesSchema, validators []Validator, updateStore bool) []Validator {
	segments, err := m.adapter.FindSegments(&store.SegmentFilter{
		Pagination: defaultPagination,
		Process:    governanceProcessName,
		Tags:       []string{process, validatorTag},
	})
	if err != nil {
		log.Errorf("Cannot retrieve gouvernance segments: %+v", errors.WithStack(err))
		return validators
	}
	if len(segments) == 0 {
		log.Warnf("No gouvernance segments found for process %s", process)
		if err = m.uploadValidator(process, schema, nil); err != nil {
			log.Warnf("Cannot upload validator %s", err)
		}
		return validators
	}
	link := segments[0].Link
	var hasToUpdateStore bool
	if err := m.compareFromStore(link.Meta, "pki", schema.PKI); err != nil {
		log.Errorf("Problem when loading pki: %s", err)
		hasToUpdateStore = true
	}
	if err := m.compareFromStore(link.Meta, "types", schema.Types); err != nil {
		log.Errorf("Problem when loading types: %s", err)
		hasToUpdateStore = true
	}
	// if updateStore && hasToUpdateStore {
	if hasToUpdateStore {
		if err = m.uploadValidator(process, schema, &link); err != nil {
			log.Warnf("Cannot upload validator %s", err)
		}
	}

	return validators
}

func (m *GovernanceManager) uploadValidator(process string, schema rulesSchema, prevLink *cs.Link) error {
	priority := 0.
	mapID := ""
	prevLinkHash := ""
	if prevLink != nil {
		priority = prevLink.GetPriority() + 1
		mapID = prevLink.GetMapID()
		var err error
		if prevLinkHash, err = prevLink.HashString(); err != nil {
			return errors.Wrapf(err, "cannot get previous hash for process governance %s", process)
		}
	} else {
		mapID = uuid.NewV4().String()
	}
	linkMeta := map[string]interface{}{
		"process":  governanceProcessName,
		"mapId":    mapID,
		"priority": priority,
		"tags":     []interface{}{process, validatorTag},
		"pki":      schema.PKI,
		"types":    schema.Types,
	}

	if prevLink != nil {
		linkMeta["prevLinkHash"] = prevLinkHash
	}

	link := &cs.Link{
		State:      map[string]interface{}{},
		Meta:       linkMeta,
		Signatures: cs.Signatures{},
	}

	hash, err := m.adapter.CreateLink(link)
	if err != nil {
		return errors.Wrapf(err, "cannot create link for process governance %s", process)
	}
	log.Infof("New validator rules store for process %s: %q", process, hash)
	return nil
}

func (m *GovernanceManager) compareFromStore(meta map[string]interface{}, key string, fileData json.RawMessage) error {
	metaData, ok := meta[key]
	if !ok {
		return errors.Errorf("%s is missing on segment", key)
	}
	storeData, err := json.Marshal(metaData)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(storeData, fileData) {
		return errors.New("data different from file and from store")
	}
	return nil
}

// UpdateValidators will replace validator if a new one is available
func (m *GovernanceManager) UpdateValidators(v *Validator) {
	if m.validatorWatcher != nil {
		var validatorFile string
		select {
		case event := <-m.validatorWatcher.Events:
			if event.Op&fsnotify.Write == fsnotify.Write {
				validatorFile = event.Name
			}
		case err := <-m.validatorWatcher.Errors:
			log.Warnf("Validator file watcher error caught: %s", err)
		default:
			break
		}
		if validatorFile != "" {
			go func() {
				if err := m.loadValidatorFile(validatorFile, true); err != nil {
					log.Warnf("cannot load validator configuration file %s: %+v", validatorFile, err)
				}
			}()
		}
	}
	select {
	case validator := <-m.validatorChan:
		*v = validator
	default:
		return
	}
}
