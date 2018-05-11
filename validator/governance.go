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

	cj "github.com/gibson042/canonicaljson-go"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"

	"github.com/prometheus/common/log"
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/validator/validators"
)

const (
	// GovernanceProcessName is the process name used for governance information storage
	GovernanceProcessName = "_governance"

	// ValidatorTag is the tag used to find validators in storage
	ValidatorTag = "validators"

	// DefaultFilename is the default filename for the file with the rules of validation
	DefaultFilename = "/data/validation/rules.json"

	// DefaultPluginsDirectory is the default directory where validation plugins are located
	DefaultPluginsDirectory = "/data/validation/"
)

var (
	// ErrNoFileWatcher is the error returned when the provided rules file could not be watched.
	ErrNoFileWatcher = errors.New("cannot listen for file updates: no file watcher")

	defaultPagination = store.Pagination{
		Offset: 0,
		Limit:  1,
	}
)

// Config contains the path of the rules JSON file and the directory where the validator scripts are located.
type Config struct {
	RulesPath   string
	PluginsPath string
}

// GovernanceManager defines the methods to implement to manage validations in an indigo network.
type GovernanceManager interface {

	// ListenAndUpdate will update the current validators whenever a change occurs in the governance rules.
	// This method must be run in a goroutine as it will wait for events from the network or file updates.
	ListenAndUpdate(ctx context.Context) error

	// AddListener adds a listener for validator changes.
	AddListener() <-chan validators.Validator

	// RemoveListener removes a listener.
	RemoveListener(<-chan validators.Validator)

	// Current returns the current version of the validator set.
	Current() validators.Validator
}

// GovernanceStore stores validation rules in an indigo store.
type GovernanceStore struct {
	store store.Adapter

	validationCfg *Config
}

// NewGovernanceStore returns a new governance store.
func NewGovernanceStore(adapter store.Adapter, validationCfg *Config) *GovernanceStore {
	return &GovernanceStore{
		store:         adapter,
		validationCfg: validationCfg,
	}
}

// GetValidators returns the list of validators for each process by fetching them from the store.
func (s *GovernanceStore) GetValidators(ctx context.Context) ([]validators.Validators, error) {
	validators := make([]validators.Validators, 0)
	for _, process := range s.getAllProcesses(ctx) {
		processValidators, err := s.getProcessValidators(ctx, process)
		if err != nil {
			return nil, err
		}
		validators = append(validators, processValidators)
	}

	return validators, nil
}

func (s *GovernanceStore) getAllProcesses(ctx context.Context) []string {
	processSet := make(map[string]interface{}, 0)
	for offset := 0; offset >= 0; {
		segments, err := s.store.FindSegments(ctx, &store.SegmentFilter{
			Pagination: store.Pagination{Offset: offset, Limit: store.MaxLimit},
			Process:    GovernanceProcessName,
			Tags:       []string{ValidatorTag},
		})
		if err != nil {
			log.Errorf("Cannot retrieve governance segments: %+v", errors.WithStack(err))
			return []string{}
		}
		for _, segment := range segments {
			for _, tag := range segment.Link.Meta.Tags {
				if tag != ValidatorTag {
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

func (s *GovernanceStore) getProcessValidators(ctx context.Context, process string) (validators.Validators, error) {
	segments, err := s.store.FindSegments(ctx, &store.SegmentFilter{
		Pagination: defaultPagination,
		Process:    GovernanceProcessName,
		Tags:       []string{process, ValidatorTag},
	})
	if err != nil || len(segments) == 0 {
		return nil, errors.New("could not find governance segments")
	}
	linkState := segments[0].Link.State
	pki, ok := linkState["pki"]
	types, ok2 := linkState["types"]
	if !ok || !ok2 {
		return nil, errors.New("governance segment is incomplete")
	}
	rawPKI, ok := pki.(json.RawMessage)
	rawTypes, ok2 := types.(json.RawMessage)
	if !ok || !ok2 {
		return nil, errors.New("governance segment is badly formatted")
	}
	return LoadProcessRules(processesRules{
		process: RulesSchema{
			PKI:   rawPKI,
			Types: rawTypes,
		},
	}, s.validationCfg.PluginsPath, nil)
}

func (s *GovernanceStore) updateValidatorInStore(ctx context.Context, process string, schema RulesSchema, validators []validators.Validator) error {
	segments, err := s.store.FindSegments(ctx, &store.SegmentFilter{
		Pagination: defaultPagination,
		Process:    GovernanceProcessName,
		Tags:       []string{process, ValidatorTag},
	})
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "Cannot retrieve governance segments")
	}
	if len(segments) == 0 {
		log.Warnf("No governance segments found for process %s", process)
		if err = s.uploadValidator(ctx, process, schema, nil); err != nil {
			return errors.Wrap(err, "Cannot upload validator")
		}
		return nil
	}
	link := segments[0].Link
	if canonicalCompare(link.State["pki"], schema.PKI) != nil ||
		canonicalCompare(link.State["types"], schema.Types) != nil {
		log.Infof("Validator or process %s has to be updated in store", process)
		if err = s.uploadValidator(ctx, process, schema, &link); err != nil {
			log.Warnf("Cannot upload validator: %s", err)
			return err
		}
	}

	return nil
}

func (s *GovernanceStore) uploadValidator(ctx context.Context, process string, schema RulesSchema, prevLink *cs.Link) error {
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
		Process:      GovernanceProcessName,
		MapID:        mapID,
		PrevLinkHash: prevLinkHash,
		Priority:     priority,
		Tags:         []string{process, ValidatorTag},
	}

	link := &cs.Link{
		State:      linkState,
		Meta:       linkMeta,
		Signatures: cs.Signatures{},
	}

	hash, err := s.store.CreateLink(ctx, link)
	if err != nil {
		return errors.Wrapf(err, "cannot create link for process governance %s", process)
	}
	log.Infof("New validator rules store for process %s: %q", process, hash)
	return nil
}

func canonicalCompare(metaData interface{}, fileData json.RawMessage) error {
	if metaData == nil {
		return errors.Errorf("missing data to compare")
	}
	canonStoreData, err := cj.Marshal(metaData)
	if err != nil {
		return errors.Wrapf(err, "cannot canonical marshal store data")
	}

	var typedData interface{}
	if err := json.Unmarshal(fileData, &typedData); err != nil {
		return err
	}
	canonFileData, err := cj.Marshal(typedData)
	if err != nil {
		return errors.Wrapf(err, "cannot canonical marshal file data")
	}

	if !bytes.Equal(canonStoreData, canonFileData) {
		return errors.New("data different from file and from store")
	}
	return nil
}
