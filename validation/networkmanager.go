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

	"github.com/pkg/errors"
	"github.com/prometheus/common/log"
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/validation/validators"
)

var (
	// ErrMissingProcess is the error returned when the ProcessMetaKey could not be found in the link's meta data.
	ErrMissingProcess = errors.New("governed process name is missing in the link's meta data")

	// ErrNoNetworkListener is returned when the provided channel is nil.
	ErrNoNetworkListener = errors.New("missing network listener")
)

// NetworkManager manages governance for validation rules management in an indigo network.
type NetworkManager struct {
	*UpdateBroadcaster
	store *Store

	validationCfg *Config
	current       validators.Validator

	networkListener <-chan *cs.Link
}

// NewNetworkManager returns a new NetworkManager able to listen to the network and update governance rules.
func NewNetworkManager(ctx context.Context, a store.Adapter, networkListener <-chan *cs.Link, validationCfg *Config) (Manager, error) {
	var err error
	var govMgr = NetworkManager{
		UpdateBroadcaster: NewUpdateBroadcaster(),
		store:             NewStore(a, validationCfg),
		validationCfg:     validationCfg,
		networkListener:   networkListener,
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

// ListenAndUpdate implements github.com/go-indigocore/validation.Manager.ListenAndUpdate.
// It will update the current validators whenever the provided rule file is updated.
// This method must be run in a goroutine as it will wait for write events on the file.
func (m *NetworkManager) ListenAndUpdate(ctx context.Context) error {
	if m.networkListener == nil {
		return ErrNoNetworkListener
	}

	for {
		select {
		case link := <-m.networkListener:
			if isGovernanceLink(link) {
				if validators, err := m.GetValidators(ctx, link); err == nil {
					m.updateCurrent(validators)
				} else {
					log.Error(err)
				}
			}

		case <-ctx.Done():
			m.Close()
			return ctx.Err()
		}
	}
}

// GetValidators extract the config from a link, parses it and returns as list of validators.
func (m *NetworkManager) GetValidators(ctx context.Context, link *cs.Link) (validators.ProcessesValidators, error) {
	jsonRules, err := json.Marshal(link.State)
	if err != nil {
		return nil, err
	}
	var schema RulesSchema
	if err := json.Unmarshal(jsonRules, &schema); err != nil {
		return nil, err
	}

	process, ok := link.Meta.Data["process"].(string)
	if !ok {
		return nil, ErrMissingProcess
	}

	processesValidators := make(validators.ProcessesValidators, 0)
	processRulesUpdate := func(process string, schema *RulesSchema, validators validators.Validators) {
		m.store.UpdateValidator(ctx, process, schema)
		processesValidators[process] = validators
	}
	if _, err = LoadProcessRules(&schema, process, m.validationCfg.PluginsPath, processRulesUpdate); err != nil {
		return nil, err
	}

	return m.store.GetValidators(ctx)
}

// Current implements github.com/go-indigocore/validation.Manager.Current.
// It returns the current validator set
func (m *NetworkManager) Current() validators.Validator {
	return m.current
}

func (m *NetworkManager) updateCurrent(validatorsMap validators.ProcessesValidators) {
	m.current = validators.NewMultiValidator(validatorsMap)
	m.Broadcast(m.current)
}

func isGovernanceLink(link *cs.Link) bool {
	if link.Meta.Process == GovernanceProcessName {
		for _, tag := range link.Meta.Tags {
			if tag == ValidatorTag {
				return true
			}
		}
	}
	return false
}
