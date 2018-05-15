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
	"bytes"
	"context"
	"encoding/json"

	cj "github.com/gibson042/canonicaljson-go"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"

	"github.com/prometheus/common/log"
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/validation/validators"
)

const (
	// GovernanceProcessName is the process name used for governance information storage.
	GovernanceProcessName = "_governance"

	// ProcessMetaKey is the key used to store the governed process name in the link's meta data.
	ProcessMetaKey = "process"

	// ValidatorTag is the tag used to find validators in storage.
	ValidatorTag = "validators"
)

var (
	// ErrNoFileWatcher is the error returned when the provided rules file could not be watched.
	ErrNoFileWatcher = errors.New("cannot listen for file updates: no file watcher")

	// ErrValidatorNotFound is the error returned when no governance segment was found for a process.
	ErrValidatorNotFound = errors.New("could not find governance segments")

	// ErrBadGovernanceSegment is the error returned when the governance segment has a bad format
	ErrBadGovernanceSegment = errors.New("governance segment is badly formatted")

	defaultPagination = store.Pagination{
		Offset: 0,
		Limit:  1,
	}
)

// Store stores validation rules in an indigo store.
type Store struct {
	store store.Adapter

	validationCfg *Config
}

// NewStore returns a new governance store.
func NewStore(adapter store.Adapter, validationCfg *Config) *Store {
	return &Store{
		store:         adapter,
		validationCfg: validationCfg,
	}
}

// GetValidators returns the list of validators for each process by fetching them from the store.
func (s *Store) GetValidators(ctx context.Context) (validators.ProcessesValidators, error) {
	var err error
	validators := make(validators.ProcessesValidators, 0)

	for _, process := range s.GetAllProcesses(ctx) {
		validators[process], err = s.getProcessValidators(ctx, process)
		if err != nil {
			return nil, err
		}
	}

	return validators, nil
}

// GetAllProcesses returns the list of processes for which governance rules have been found.
func (s *Store) GetAllProcesses(ctx context.Context) []string {
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

func (s *Store) getProcessValidators(ctx context.Context, process string) (validators.Validators, error) {
	segments, err := s.store.FindSegments(ctx, &store.SegmentFilter{
		Pagination: defaultPagination,
		Process:    GovernanceProcessName,
		Tags:       []string{process, ValidatorTag},
	})
	if err != nil || len(segments) == 0 {
		return nil, ErrValidatorNotFound
	}
	linkState := segments[0].Link.State

	var pki validators.PKI
	if err := mapToStruct(linkState["pki"], &pki); err != nil {
		return nil, ErrBadGovernanceSegment
	}
	var types map[string]TypeSchema
	if err := mapToStruct(linkState["types"], &types); err != nil {
		return nil, ErrBadGovernanceSegment
	}

	return LoadProcessRules(&RulesSchema{
		PKI:   &pki,
		Types: types,
	}, process, s.validationCfg.PluginsPath, nil)
}

// UpdateValidator replaces the current validation rules in the store by the provided ones.
// If none was found in the store, they will be created.
func (s *Store) UpdateValidator(ctx context.Context, process string, schema *RulesSchema) error {
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

func (s *Store) uploadValidator(ctx context.Context, process string, schema *RulesSchema, prevLink *cs.Link) error {
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
		Data:         map[string]interface{}{ProcessMetaKey: process},
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

func canonicalCompare(metaData interface{}, fileData interface{}) error {
	if metaData == nil {
		return errors.Errorf("missing data to compare")
	}
	canonStoreData, err := cj.Marshal(metaData)
	if err != nil {
		return errors.Wrapf(err, "cannot canonical marshal store data")
	}
	canonFileData, err := cj.Marshal(fileData)
	if err != nil {
		return errors.Wrapf(err, "cannot canonical marshal file data")
	}

	if !bytes.Equal(canonStoreData, canonFileData) {
		return errors.New("data different from file and from store")
	}
	return nil
}

func mapToStruct(src interface{}, dest interface{}) error {
	bytes, err := json.Marshal(src)
	if err != nil {
		return err
	}

	return json.Unmarshal(bytes, dest)
}
