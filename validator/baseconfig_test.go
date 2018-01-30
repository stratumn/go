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

	"github.com/stretchr/testify/assert"
)

func TestBaseConfig(t *testing.T) {
	process := "p1"
	linkType := "sell"

	type testCaseCfg struct {
		name          string
		process       string
		linkType      string
		schema        []byte
		valid         bool
		expectedError error
	}

	testCases := []testCaseCfg{{
		name:          "missing-process",
		process:       "",
		linkType:      linkType,
		valid:         false,
		expectedError: ErrMissingProcess,
	}, {
		name:          "missing-link-type",
		process:       process,
		linkType:      "",
		valid:         false,
		expectedError: ErrMissingLinkType,
	}, {
		name:     "valid-config",
		process:  process,
		linkType: linkType,
		valid:    true,
	},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := newValidatorBaseConfig(
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
