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

	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-core/store"
	"github.com/stratumn/go-core/validation/validators"
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

	// ErrBadPriority is returned when the new governance link's priority is less than or equal to the previous governance link's priority.
	ErrBadPriority = errors.New("priority has to be higher than previous governance link")

	// ErrBadPrevLinkHash is returned when the new governance link's prevLInkHash is different from the previous governance link's hash.
	ErrBadPrevLinkHash = errors.New("prevLinkHash does not match previous governance link")

	// ErrBadMapID is returned when the governance link's mapID does not match the previous link's mapID.
	ErrBadMapID = errors.New("governance rules for a single process must belong to the same map")

	// ErrBadProcess is returned when the governance link's process does not match the previous link's process.
	ErrBadProcess = errors.New("process does not match the previous governance link")

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
	validators := make(validators.ProcessesValidators)

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
	processSet := make(map[string]interface{})
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
		for _, segment := range segments.Segments {
			for _, tag := range segment.Link.Meta.Tags {
				if tag != ValidatorTag {
					processSet[tag] = nil
				}
			}
		}
		if len(segments.Segments)+offset*store.MaxLimit >= segments.TotalCount {
			break
		}
		offset += store.MaxLimit
	}
	ret := make([]string, 0, len(processSet))
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
	if err != nil || len(segments.Segments) == 0 {
		return nil, ErrValidatorNotFound
	}

	var linkData map[string]interface{}
	err = segments.Segments[0].Link.StructurizeData(&linkData)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var pki validators.PKI
	if err := mapToStruct(linkData["pki"], &pki); err != nil {
		return nil, ErrBadGovernanceSegment
	}
	var types map[string]TypeSchema
	if err := mapToStruct(linkData["types"], &types); err != nil {
		return nil, ErrBadGovernanceSegment
	}

	return LoadProcessRules(&RulesSchema{
		PKI:   &pki,
		Types: types,
	}, process, s.validationCfg.PluginsPath, nil)
}

// UpdateValidator replaces the current validation rules in the store by the provided ones.
// It checks that the provided link correcly references the previous rules' link.
func (s *Store) UpdateValidator(ctx context.Context, link *chainscript.Link) error {
	var linkMetadata map[string]string
	err := link.StructurizeMetadata(&linkMetadata)
	if err != nil {
		return errors.WithStack(err)
	}

	process, ok := linkMetadata[ProcessMetaKey]
	if !ok {
		return ErrMissingProcess
	}

	segments, err := s.store.FindSegments(ctx, &store.SegmentFilter{
		Pagination: defaultPagination,
		Process:    GovernanceProcessName,
		Tags:       []string{process, ValidatorTag},
	})
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "Cannot retrieve governance segments")
	}

	if len(segments.Segments) == 0 {
		log.Infof("No governance segments found for process %s, creating validator", process)
		if len(link.Meta.PrevLinkHash) > 0 {
			return ErrBadPrevLinkHash
		}

		return s.uploadValidator(ctx, link)
	}

	lastGovernanceLink := segments.Segments[0].Link
	var lastGovernanceLinkMetadata map[string]string
	err = lastGovernanceLink.StructurizeMetadata(&lastGovernanceLinkMetadata)
	if err != nil {
		return errors.WithStack(err)
	}

	if !bytes.Equal(link.Data, lastGovernanceLink.Data) {
		log.Infof("Validator of process %s has to be updated in store", process)
		if link.Meta.Priority <= lastGovernanceLink.Meta.Priority {
			return ErrBadPriority
		}

		lastGovernanceLinkHash, _ := lastGovernanceLink.Hash()
		if !bytes.Equal(link.Meta.PrevLinkHash, lastGovernanceLinkHash) {
			return ErrBadPrevLinkHash
		}

		if link.Meta.MapId != lastGovernanceLink.Meta.MapId {
			return ErrBadMapID
		}

		if process != lastGovernanceLinkMetadata[ProcessMetaKey] {
			return ErrBadProcess
		}

		return s.uploadValidator(ctx, link)
	}
	return nil
}

func (s *Store) uploadValidator(ctx context.Context, link *chainscript.Link) error {
	var linkMetadata map[string]string
	err := link.StructurizeMetadata(&linkMetadata)
	if err != nil {
		return errors.WithStack(err)
	}

	hash, err := s.store.CreateLink(ctx, link)
	if err != nil {
		return errors.Wrapf(err, "cannot create link for process governance %s", linkMetadata[ProcessMetaKey])
	}

	log.Infof("New validator rules store for process %s: %q", linkMetadata[ProcessMetaKey], hash)
	return nil
}

// LinkFromSchema creates a chainscript link from a PKI and a set of rules.
// It first tries to fetch the previous governance link for this process and
// builds the new one on top of it.
// If no previous governance link exists, a link from a new map is created.
func (s *Store) LinkFromSchema(ctx context.Context, process string, schema *RulesSchema) (*chainscript.Link, error) {
	var lastGovernanceLink *chainscript.Link
	segments, err := s.store.FindSegments(ctx, &store.SegmentFilter{
		Pagination: defaultPagination,
		Process:    GovernanceProcessName,
		Tags:       []string{process, ValidatorTag},
	})
	if err != nil {
		return nil, errors.Wrap(errors.WithStack(err), "Cannot retrieve governance segments")
	}

	linkData := map[string]interface{}{
		"pki":   schema.PKI,
		"types": schema.Types,
	}

	// If we have a previous governance link, extend it.
	if len(segments.Segments) > 0 {
		lastGovernanceLink = segments.Segments[0].Link
		priority := lastGovernanceLink.Meta.Priority + 1.
		mapID := lastGovernanceLink.Meta.MapId

		prevLinkHash, err := lastGovernanceLink.Hash()
		if err != nil {
			return nil, errors.Wrapf(err, "cannot get previous hash for process governance %s", process)
		}

		return chainscript.NewLinkBuilder(GovernanceProcessName, mapID).
			WithParent(prevLinkHash).
			WithPriority(priority).
			WithTags(process, ValidatorTag).
			WithData(linkData).
			WithMetadata(map[string]string{ProcessMetaKey: process}).
			Build()
	}

	// Otherwise initialize a new map.
	mapID := uuid.NewV4().String()
	return chainscript.NewLinkBuilder(GovernanceProcessName, mapID).
		WithTags(process, ValidatorTag).
		WithData(linkData).
		WithMetadata(map[string]string{ProcessMetaKey: process}).
		Build()
}

func mapToStruct(src interface{}, dest interface{}) error {
	bytes, err := json.Marshal(src)
	if err != nil {
		return err
	}

	return json.Unmarshal(bytes, dest)
}
