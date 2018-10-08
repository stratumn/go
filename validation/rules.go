// Copyright 2016-2018 Stratumn SAS. All rights reserved.
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
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/stratumn/go-core/validation/validators"
)

// Errors used by the validation rules parsers.
var (
	ErrInvalidStepValidationRules = errors.New("a step validator requires a JSON schema, a signature or a transition criteria to be valid")
	ErrMissingPKI                 = errors.New("PKI is missing from validation rules")
	ErrMissingTransitions         = errors.New("transitions must be defined for every step")
	ErrInvalidTransitions         = errors.New("invalid step transition")
)

// ProcessesRules maps processes to their validation rules.
type ProcessesRules map[string]*ProcessRules

// ProcessRules contains the validation rules that apply to a specific process.
type ProcessRules struct {
	PKI    validators.PKI           `json:"pki"`
	Steps  map[string]*StepRules    `json:"steps"`
	Script *validators.ScriptConfig `json:"script"`
}

// StepRules contains the validation rules that apply to a specific step inside
// a given process.
type StepRules struct {
	Signatures  []string               `json:"signatures"`
	Schema      map[string]interface{} `json:"schema"`
	Transitions []string               `json:"transitions"`
}

// Validators creates the validators corresponding to the configured rules.
func (r ProcessesRules) Validators(pluginsPath string) (validators.ProcessesValidators, error) {
	var err error
	validators := make(validators.ProcessesValidators)
	for process, processRules := range r {
		validators[process], err = processRules.Validators(process, pluginsPath)
		if err != nil {
			return nil, err
		}
	}

	return validators, nil
}

// ValidateTransitions checks for human errors in the transitions definitions
// (for example some steps that cannot be reached).
func (r *ProcessRules) ValidateTransitions() error {
	// If transition validation is defined for one step in the process, it has
	// to be defined for every other step.
	// Additionally if some steps are orphaned (unreachable) we raise an error
	// to warn the user that he likely made a mistake.
	transitionsEnabled := false
	stepsMap := make(map[string]struct{})
	for step, rules := range r.Steps {
		stepsMap[step] = struct{}{}
		if len(rules.Transitions) > 0 {
			transitionsEnabled = true
		}
	}

	if !transitionsEnabled {
		return nil
	}

	for step, rules := range r.Steps {
		if len(rules.Transitions) == 0 {
			return errors.Wrapf(ErrMissingTransitions, "%s has no transitions", step)
		}

		if len(rules.Transitions) == 1 && rules.Transitions[0] == step {
			return errors.Wrapf(ErrInvalidTransitions, "%s can only be reached from itself", step)
		}

		// Verify that transitions are defined on existing steps.
		for _, transition := range rules.Transitions {
			if len(transition) > 0 {
				_, ok := stepsMap[transition]
				if !ok {
					return errors.Wrapf(ErrInvalidTransitions, "%s -> %s is invalid: %s doesn't exist in the process", transition, step, transition)
				}
			}
		}
	}

	return nil
}

// Validators creates the validators corresponding to the configured process' rules.
func (r *ProcessRules) Validators(process string, pluginsPath string) (validators.Validators, error) {
	var processValidators validators.Validators

	if r.PKI != nil {
		err := r.PKI.Validate()
		if err != nil {
			return nil, err
		}
	}

	if err := r.ValidateTransitions(); err != nil {
		return nil, err
	}

	if r.Script != nil {
		scriptValidator, err := validators.NewScriptValidator(process, pluginsPath, r.Script)
		if err != nil {
			return nil, err
		}

		processValidators = append(processValidators, scriptValidator)
	}

	for step, stepRules := range r.Steps {
		stepValidators, err := stepRules.Validators(process, step, r.PKI)
		if err != nil {
			return nil, err
		}

		processValidators = append(processValidators, stepValidators...)
	}

	return processValidators, nil
}

// Validators creates the validators corresponding to the configured step's rules.
func (r *StepRules) Validators(process, step string, pki validators.PKI) (validators.Validators, error) {
	if len(r.Signatures) == 0 && len(r.Transitions) == 0 && r.Schema == nil {
		return nil, ErrInvalidStepValidationRules
	}

	processStepValidator, err := validators.NewProcessStepValidator(process, step)
	if err != nil {
		return nil, err
	}

	var stepValidators validators.Validators

	if r.Schema != nil {
		jsonSchema, err := json.Marshal(r.Schema)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		schemaValidator, err := validators.NewSchemaValidator(processStepValidator, jsonSchema)
		if err != nil {
			return nil, err
		}

		stepValidators = append(stepValidators, schemaValidator)
	}

	if len(r.Signatures) > 0 {
		if pki == nil {
			return nil, ErrMissingPKI
		}

		pkiValidator := validators.NewPKIValidator(processStepValidator, r.Signatures, pki)
		stepValidators = append(stepValidators, pkiValidator)
	}

	if len(r.Transitions) > 0 {
		transitionValidator := validators.NewTransitionValidator(processStepValidator, r.Transitions)
		stepValidators = append(stepValidators, transitionValidator)
	}

	return stepValidators, nil
}
