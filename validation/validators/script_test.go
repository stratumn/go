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
	"context"
	"os"
	"testing"

	"github.com/pkg/errors"
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/cs/cstesting"
	"github.com/stratumn/go-indigocore/dummystore"
	"github.com/stratumn/go-indigocore/testutil"
	"github.com/stratumn/go-indigocore/validation/validators"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScriptValidator(t *testing.T) {

	testLink := cstesting.NewLinkBuilder().WithProcess("test").WithType("init").Build()

	t.Run("New", func(t *testing.T) {

		type testCase struct {
			name      string
			baseCfg   *validators.ValidatorBaseConfig
			scriptCfg *validators.ScriptConfig
			valid     bool
			err       string
		}

		testCases := []testCase{
			{
				name: "valid-config",
				baseCfg: &validators.ValidatorBaseConfig{
					Process:  "test",
					LinkType: "init",
				},
				scriptCfg: &validators.ScriptConfig{
					File: pluginFile,
					Type: "go",
				},
				valid: true,
			},
			{
				name: "bad-script-type",
				baseCfg: &validators.ValidatorBaseConfig{
					Process:  "test",
					LinkType: "invalid",
				},
				scriptCfg: &validators.ScriptConfig{
					File: pluginFile,
					Type: "bad",
				},
				valid: false,
				err:   errors.Errorf(validators.ErrBadScriptType, "bad", validators.ValidScriptTypes).Error(),
			},
			{
				name: "script-not-found",
				baseCfg: &validators.ValidatorBaseConfig{
					Process:  "test",
					LinkType: "invalid",
				},
				scriptCfg: &validators.ScriptConfig{
					File: "test",
					Type: "go",
				},
				valid: false,
				err:   errors.Wrapf(errors.New("plugin.Open(\"test\"): realpath failed"), validators.ErrLoadingPlugin, "test", "invalid").Error(),
			},
			{
				name: "unknown-script-validator-for-linkType",
				baseCfg: &validators.ValidatorBaseConfig{
					Process:  "test",
					LinkType: "unknown",
				},
				scriptCfg: &validators.ScriptConfig{
					File: pluginFile,
					Type: "go",
				},
				valid: false,
			},
			{
				name: "bad-script-function-signature",
				baseCfg: &validators.ValidatorBaseConfig{
					Process:  "test",
					LinkType: "badSignature",
				},
				scriptCfg: &validators.ScriptConfig{
					File: pluginFile,
					Type: "go",
				},
				valid: false,
				err:   errors.Wrapf(errors.New(validators.ErrBadPlugin), validators.ErrLoadingPlugin, "test", "badSignature").Error(),
			},
		}

		for _, tt := range testCases {
			t.Run(tt.name, func(t *testing.T) {
				_, err := validators.NewScriptValidator(tt.baseCfg, tt.scriptCfg, "")
				if tt.valid {
					assert.NoError(t, err)
				} else {
					if tt.err != "" {
						assert.EqualError(t, err, tt.err)
					} else {
						assert.Error(t, err)
					}
				}
			})
		}
	})

	t.Run("Hash", func(t *testing.T) {
		// in this test, we try to load the same symbol from different files to check that the hash are different
		baseCfg, err := validators.NewValidatorBaseConfig("test", "init")
		require.NoError(t, err)

		byzantinePluginFile, err := testutil.CompileGoPlugin(pluginsPath, "byzantine_validator.go")
		require.NoError(t, err)
		defer os.Remove(byzantinePluginFile)

		scriptCfg1 := &validators.ScriptConfig{
			Type: "go",
			File: pluginFile,
		}
		scriptCfg2 := &validators.ScriptConfig{
			Type: "go",
			File: byzantinePluginFile,
		}

		v1, err := validators.NewScriptValidator(baseCfg, scriptCfg1, "")
		require.NoError(t, err)
		v2, err := validators.NewScriptValidator(baseCfg, scriptCfg2, "")
		require.NoError(t, err)

		hash1, err1 := v1.Hash()
		hash2, err2 := v2.Hash()
		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.NotNil(t, hash1)
		assert.NotNil(t, hash2)
		assert.NotEqual(t, hash1.String(), hash2.String())
	})

	t.Run("ShouldValidate", func(t *testing.T) {

		type testCase struct {
			name      string
			ret       bool
			baseCfg   *validators.ValidatorBaseConfig
			scriptCfg *validators.ScriptConfig
			link      *cs.Link
		}

		scriptConfig := &validators.ScriptConfig{
			File: pluginFile,
			Type: "go",
		}

		testCases := []testCase{
			{
				name: "has to validate",
				ret:  true,
				baseCfg: &validators.ValidatorBaseConfig{
					Process:  "test",
					LinkType: "init",
				},
				link: testLink,
			},
			{
				name: "bad process",
				ret:  false,
				baseCfg: &validators.ValidatorBaseConfig{
					Process:  "bad",
					LinkType: "init",
				},
				link: cstesting.NewLinkBuilder().WithProcess("test").Build(),
			},
			{
				name: "bad type",
				ret:  false,
				baseCfg: &validators.ValidatorBaseConfig{
					Process:  "test",
					LinkType: "invalid",
				},
				link: cstesting.NewLinkBuilder().WithType("init").Build(),
			},
		}

		for _, tt := range testCases {
			t.Run(tt.name, func(t *testing.T) {
				v, err := validators.NewScriptValidator(tt.baseCfg, scriptConfig, "")
				require.NoError(t, err)
				assert.Equal(t, tt.ret, v.ShouldValidate(tt.link))
			})
		}
	})

	t.Run("Validate", func(t *testing.T) {

		type testCase struct {
			name     string
			testLink *cs.Link
			baseCfg  *validators.ValidatorBaseConfig
			valid    bool
			err      string
		}

		scriptCfg := &validators.ScriptConfig{
			File: pluginFile,
			Type: "go",
		}
		testCases := []testCase{
			{
				name: "valid-link",
				baseCfg: &validators.ValidatorBaseConfig{
					Process:  "test",
					LinkType: "init",
				},
				testLink: testLink,
				valid:    true,
			},
			{
				name: "fetch-link",
				baseCfg: &validators.ValidatorBaseConfig{
					Process:  "test",
					LinkType: "fetchLink",
				},
				testLink: cstesting.NewLinkBuilder().WithType("fetchLink").Build(),
				valid:    true,
			},
			{
				name: "validation-fails",
				baseCfg: &validators.ValidatorBaseConfig{
					Process:  "test",
					LinkType: "invalid",
				},
				testLink: cstesting.NewLinkBuilder().WithType("invalid").Build(),
				valid:    false,
				err:      "error",
			},
		}
		for _, tt := range testCases {
			t.Run(tt.name, func(t *testing.T) {
				sv, err := validators.NewScriptValidator(tt.baseCfg, scriptCfg, "")
				require.NoError(t, err)
				err = sv.Validate(context.Background(), dummystore.New(nil), tt.testLink)
				if tt.valid {
					assert.NoError(t, err)
				} else {
					assert.EqualError(t, err, tt.err)
				}
			})
		}
	})
}
