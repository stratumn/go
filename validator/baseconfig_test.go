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
	"testing"

	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/cs/cstesting"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBaseConfig(t *testing.T) {
	process := "p1"
	action := "sell"

	type testCaseCfg struct {
		id            string
		process       string
		action        string
		schema        []byte
		valid         bool
		expectedError error
	}

	testCases := []testCaseCfg{{
		id:            "missing-process",
		process:       "",
		action:        action,
		valid:         false,
		expectedError: ErrMissingProcess,
	}, {
		id:            "",
		process:       process,
		action:        action,
		valid:         false,
		expectedError: ErrMissingIdentifier,
	}, {
		id:            "missing-link-type",
		process:       process,
		action:        "",
		valid:         false,
		expectedError: ErrMissingLinkType,
	}, {
		id:      "valid-config",
		process: process,
		action:  action,
		valid:   true,
	},
	}

	for _, tt := range testCases {
		t.Run(tt.id, func(t *testing.T) {
			cfg, err := newValidatorBaseConfig(
				tt.process,
				tt.id,
				tt.action,
			)

			if tt.valid {
				assert.NotNil(t, cfg)
				assert.NoError(t, err)
			} else {
				assert.Nil(t, cfg)
				assert.Error(t, err)
				if tt.expectedError != nil {
					assert.EqualError(t, err, tt.expectedError.Error())
				}
			}
		})
	}
}

func TestBaseConfig_ShouldValidate(t *testing.T) {
	process := "p1"
	action := "sell"
	cfg, err := newValidatorBaseConfig(
		process,
		"test",
		action,
	)
	require.NoError(t, err)

	createValidLink := func() *cs.Link {
		l := cstesting.RandomLink()
		l.Meta["process"] = process
		l.Meta["action"] = action
		return cstesting.SignLink(l)
	}
	type testCase struct {
		name           string
		link           func() *cs.Link
		shouldValidate bool
	}

	testCases := []testCase{
		{
			name:           "valid-link",
			shouldValidate: true,
			link:           createValidLink,
		},
		{
			name:           "no-process",
			shouldValidate: false,
			link: func() *cs.Link {
				l := createValidLink()
				delete(l.Meta, "process")
				return l
			},
		},
		{
			name:           "process-no-match",
			shouldValidate: false,
			link: func() *cs.Link {
				l := createValidLink()
				l.Meta["process"] = "test"
				return l
			},
		},
		{
			name:           "no-action",
			shouldValidate: false,
			link: func() *cs.Link {
				l := createValidLink()
				delete(l.Meta, "action")
				return l
			},
		},
		{
			name:           "action-no-match",
			shouldValidate: false,
			link: func() *cs.Link {
				l := createValidLink()
				l.Meta["action"] = "test"
				return l
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			res := cfg.ShouldValidate(tt.link())
			assert.Equal(t, tt.shouldValidate, res)
		})
	}
}
