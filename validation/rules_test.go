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

package validation_test

import (
	"encoding/json"
	"testing"

	"github.com/pkg/errors"
	"github.com/stratumn/go-core/validation"
	"github.com/stratumn/go-core/validation/validationtesting"
	"github.com/stratumn/go-core/validation/validators"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessRules(t *testing.T) {
	t.Run("ValidateTransitions()", func(t *testing.T) {
		t.Run("missing transitions", func(t *testing.T) {
			rules := &validation.ProcessRules{Steps: map[string]*validation.StepRules{
				"init": &validation.StepRules{
					Transitions: []string{""},
				},
				// Since 'init' uses transitions validation, 'update' should
				// also define transitions rules.
				"update": &validation.StepRules{},
			}}

			err := rules.ValidateTransitions()
			assert.EqualError(t, errors.Cause(err), validation.ErrMissingTransitions.Error())
		})

		t.Run("orphaned step", func(t *testing.T) {
			rules := &validation.ProcessRules{Steps: map[string]*validation.StepRules{
				"init": &validation.StepRules{
					Transitions: []string{""},
				},
				// Update can never be reached because the 'unknown-step'
				// cannot be reached since it's not even defined.
				"update": &validation.StepRules{
					Transitions: []string{"unknown-step"},
				},
			}}

			err := rules.ValidateTransitions()
			assert.EqualError(t, errors.Cause(err), validation.ErrInvalidTransitions.Error())
		})

		t.Run("self-referencing step", func(t *testing.T) {
			rules := &validation.ProcessRules{Steps: map[string]*validation.StepRules{
				"init": &validation.StepRules{
					Transitions: []string{""},
				},
				"update": &validation.StepRules{
					Transitions: []string{"update"},
				},
			}}

			err := rules.ValidateTransitions()
			assert.EqualError(t, errors.Cause(err), validation.ErrInvalidTransitions.Error())
		})
	})
}

func TestProcessesRules(t *testing.T) {
	var validSchema map[string]interface{}
	err := json.Unmarshal([]byte(`{"type": "object","properties": {"message": {"type": "string"}}, "required": ["message"]}`), &validSchema)
	require.NoError(t, err)

	t.Run("invalid transitions", func(t *testing.T) {
		rules := validation.ProcessesRules{
			"test_process": &validation.ProcessRules{
				Steps: map[string]*validation.StepRules{
					"init": &validation.StepRules{
						Transitions: []string{"black-hole"},
					},
				},
			},
		}

		v, err := rules.Validators("")
		assert.Error(t, err)
		assert.Nil(t, v)
	})

	t.Run("invalid PKI", func(t *testing.T) {
		rules := validation.ProcessesRules{
			"test_process": &validation.ProcessRules{
				PKI: validators.PKI{
					"alice": &validators.Identity{},
				},
				Steps: map[string]*validation.StepRules{
					"init": &validation.StepRules{
						Transitions: []string{""},
					},
				},
			},
		}

		v, err := rules.Validators("")
		assert.Error(t, err)
		assert.Nil(t, v)
	})

	t.Run("invalid script", func(t *testing.T) {
		rules := validation.ProcessesRules{
			"test_process": &validation.ProcessRules{
				Script: &validators.ScriptConfig{
					Hash: "4224",
				},
			},
		}

		v, err := rules.Validators("/tmp/validation/plugins")
		assert.Error(t, err)
		assert.Nil(t, v)
	})

	t.Run("invalid schema", func(t *testing.T) {
		rules := validation.ProcessesRules{
			"test_process": &validation.ProcessRules{
				Steps: map[string]*validation.StepRules{
					"init": &validation.StepRules{
						Schema: validSchema,
					},
					"update": &validation.StepRules{
						Schema: map[string]interface{}{
							"not": "a valid JSON schema",
						},
					},
				},
			},
		}

		v, err := rules.Validators("")
		assert.Error(t, err)
		assert.Nil(t, v)
	})

	t.Run("valid multi-process rules", func(t *testing.T) {
		rules := validation.ProcessesRules{
			"p1": &validation.ProcessRules{
				PKI: validators.PKI{
					"alice": &validators.Identity{
						Keys:  []string{validationtesting.AlicePublicKey},
						Roles: []string{"employee"},
					},
					"bob": &validators.Identity{
						Keys:  []string{validationtesting.BobPublicKey},
						Roles: []string{"manager", "senior"},
					},
				},
				Steps: map[string]*validation.StepRules{
					"init": &validation.StepRules{
						Signatures:  []string{"manager"},
						Schema:      validSchema,
						Transitions: []string{""},
					},
					"launch": &validation.StepRules{
						Signatures:  []string{"employee", "manager"},
						Schema:      validSchema,
						Transitions: []string{"init"},
					},
				},
			},
			"p2": &validation.ProcessRules{
				Steps: map[string]*validation.StepRules{
					"init": &validation.StepRules{
						Schema:      validSchema,
						Transitions: []string{""},
					},
					"end": &validation.StepRules{
						Schema:      validSchema,
						Transitions: []string{"init", ""},
					},
				},
			},
		}

		v, err := rules.Validators("")
		require.NoError(t, err)
		require.NotNil(t, v)
		assert.Len(t, v, 2)

		p1, ok := v["p1"]
		require.True(t, ok)
		assert.Len(t, p1, 6)

		p2, ok := v["p2"]
		require.True(t, ok)
		assert.Len(t, p2, 4)
	})
}
