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

package validators_test

import (
	"testing"

	"github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-chainscript/chainscripttest"
	"github.com/stratumn/go-core/validation/validators"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBaseConfig(t *testing.T) {
	process := "p1"
	linkType := "sell"

	type testCaseCfg struct {
		id            string
		process       string
		linkType      string
		schema        []byte
		valid         bool
		expectedError error
	}

	testCases := []testCaseCfg{{
		id:            "missing-process",
		process:       "",
		linkType:      linkType,
		valid:         false,
		expectedError: validators.ErrMissingProcess,
	}, {
		id:            "missing-link-type",
		process:       process,
		linkType:      "",
		valid:         false,
		expectedError: validators.ErrMissingLinkStep,
	}, {
		id:       "valid-config",
		process:  process,
		linkType: linkType,
		valid:    true,
	},
	}

	for _, tt := range testCases {
		t.Run(tt.id, func(t *testing.T) {
			cfg, err := validators.NewValidatorBaseConfig(
				tt.process,
				tt.linkType,
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
	linkStep := "sell"
	cfg, err := validators.NewValidatorBaseConfig(
		process,
		linkStep,
	)
	require.NoError(t, err)

	type testCase struct {
		name           string
		link           *chainscript.Link
		shouldValidate bool
	}

	testCases := []testCase{
		{
			name:           "valid-link",
			shouldValidate: true,
			link:           chainscripttest.NewLinkBuilder(t).WithProcess(process).WithStep(linkStep).WithSignature(t, "").Build(),
		},
		{
			name:           "no-process",
			shouldValidate: false,
			link:           chainscripttest.NewLinkBuilder(t).WithProcess("").WithStep(linkStep).WithSignature(t, "").Build(),
		},
		{
			name:           "process-no-match",
			shouldValidate: false,
			link:           chainscripttest.NewLinkBuilder(t).WithProcess("test").WithStep(linkStep).WithSignature(t, "").Build(),
		},
		{
			name:           "no-type",
			shouldValidate: false,
			link:           chainscripttest.NewLinkBuilder(t).WithProcess(process).WithStep("").WithSignature(t, "").Build(),
		},
		{
			name:           "type-no-match",
			shouldValidate: false,
			link:           chainscripttest.NewLinkBuilder(t).WithProcess(process).WithStep("test").WithSignature(t, "").Build(),
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			res := cfg.ShouldValidate(tt.link)
			assert.Equal(t, tt.shouldValidate, res)
		})
	}
}
