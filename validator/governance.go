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
	"encoding/json"

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

var defaultPagination = store.Pagination{
	Offset: 0,
	Limit:  1, // store.DefaultLimit,
}

// GovernanceManager manages governance for validation rules management.
type GovernanceManager struct {
	adapter store.Adapter

	validatorWatcher *fsnotify.Watcher
	validatorChan    chan Validator
	validators       map[string][]Validator
}

// NewGovernanceManager enhances validator management with some governance concepts.
func NewGovernanceManager(a store.Adapter, filename string) (*GovernanceManager, error) {
	var err error
	var govMgr = GovernanceManager{
		adapter:       a,
		validatorChan: make(chan Validator, 1),
		validators:    make(map[string][]Validator, 0),
	}

	govMgr.loadValidatorsFromFile(filename)
	govMgr.loadValidatorsFromStore()
	if len(govMgr.validators) > 0 {
		govMgr.sendValidators()
	}
	if filename != "" {
		if govMgr.validatorWatcher, err = fsnotify.NewWatcher(); err != nil {
			return nil, errors.Wrap(err, "cannot create a new filesystem watcher for validators")
		}
		if err := govMgr.validatorWatcher.Add(filename); err != nil {
			return nil, errors.Wrapf(err, "cannot watch validator configuration file %s", filename)
		}
	}

	return &govMgr, nil
}

func (m *GovernanceManager) loadValidatorsFromFile(filename string) (err error) {
	if filename != "" {
		_, err = LoadConfig(filename, func(process string, schema rulesSchema, validators []Validator) {
			m.validators[process] = m.updateValidatorInStore(process, schema, validators)
		})
		if err != nil {
			log.Errorf("Cannot load validator rules file %s: %s", filename, err)
		}
	}
	return err
}

func (m *GovernanceManager) loadValidatorsFromStore() {
	for _, process := range m.getAllProcesses() {
		if _, exist := m.validators[process]; !exist {
			m.validators[process] = m.getValidators(process)
		}
	}
}

func (m *GovernanceManager) sendValidators() {
	v4ch := make([]Validator, 0)
	for _, v := range m.validators {
		v4ch = append(v4ch, v...)
	}
	m.validatorChan <- NewMultiValidator(v4ch)
}

func (m *GovernanceManager) getAllProcesses() []string {
	processSet := make(map[string]interface{}, 0)
	for offset := 0; offset >= 0; {
		segments, err := m.adapter.FindSegments(&store.SegmentFilter{
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

func (m *GovernanceManager) getValidators(process string) []Validator {
	segments, err := m.adapter.FindSegments(&store.SegmentFilter{
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
	}, nil)
	if err != nil {
		return v
	}
	return nil
}

func (m *GovernanceManager) updateValidatorInStore(process string, schema rulesSchema, validators []Validator) []Validator {
	segments, err := m.adapter.FindSegments(&store.SegmentFilter{
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
		if err = m.uploadValidator(process, schema, nil); err != nil {
			log.Warnf("Cannot upload validator: %s", err)
		}
		return validators
	}
	link := segments[0].Link
	if m.compareFromStore(link.State, "pki", schema.PKI) != nil ||
		m.compareFromStore(link.State, "types", schema.Types) != nil {
		log.Infof("Validator or process %s has to be updated in store", process)
		if err = m.uploadValidator(process, schema, &link); err != nil {
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

func (m *GovernanceManager) compareFromStore(meta map[string]interface{}, key string, fileData json.RawMessage) error {
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

func (m *GovernanceManager) uploadValidator(process string, schema rulesSchema, prevLink *cs.Link) error {
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

	hash, err := m.adapter.CreateLink(link)
	if err != nil {
		return errors.Wrapf(err, "cannot create link for process governance %s", process)
	}
	log.Infof("New validator rules store for process %s: %q", process, hash)
	return nil
}

// UpdateValidators will replace validator if a new one is available
func (m *GovernanceManager) UpdateValidators(v *Validator) bool {
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
				if m.loadValidatorsFromFile(validatorFile) == nil {
					m.sendValidators()
				}
			}()
		}
	}
	select {
	case validator := <-m.validatorChan:
		*v = validator
		return true
	default:
		return false
	}
}
