package validator

import (
	"context"
	"os"
	"testing"

	"github.com/pkg/errors"
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/cs/cstesting"
	"github.com/stratumn/go-indigocore/dummystore"
	"github.com/stratumn/go-indigocore/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScriptValidator(t *testing.T) {
	testLink := cstesting.RandomLinkWithProcess("test")
	testLink.Meta.Type = "init"

	t.Run("New", func(t *testing.T) {

		type testCase struct {
			name      string
			baseCfg   *validatorBaseConfig
			scriptCfg *scriptConfig
			valid     bool
			err       string
		}

		testCases := []testCase{
			{
				name: "valid-config",
				baseCfg: &validatorBaseConfig{
					Process:  "test",
					LinkType: "init",
				},
				scriptCfg: &scriptConfig{
					File: pluginFile,
					Type: "go",
				},
				valid: true,
			},
			{
				name: "bad-script-type",
				baseCfg: &validatorBaseConfig{
					Process:  "test",
					LinkType: "invalid",
				},
				scriptCfg: &scriptConfig{
					File: pluginFile,
					Type: "bad",
				},
				valid: false,
				err:   "Validation engine does not handle script of type bad, valid types are [go]",
			},
			{
				name: "script-not-found",
				baseCfg: &validatorBaseConfig{
					Process:  "test",
					LinkType: "invalid",
				},
				scriptCfg: &scriptConfig{
					File: "test",
					Type: "go",
				},
				valid: false,
				err:   errors.Wrapf(errors.New("plugin.Open(\"test\"): realpath failed"), ErrLoadingPlugin, "test", "invalid").Error(),
			},
			{
				name: "unknown-script-validator-for-linkType",
				baseCfg: &validatorBaseConfig{
					Process:  "test",
					LinkType: "unknown",
				},
				scriptCfg: &scriptConfig{
					File: pluginFile,
					Type: "go",
				},
				valid: false,
			},
			{
				name: "bad-script-function-signature",
				baseCfg: &validatorBaseConfig{
					Process:  "test",
					LinkType: "badSignature",
				},
				scriptCfg: &scriptConfig{
					File: pluginFile,
					Type: "go",
				},
				valid: false,
				err:   errors.Wrapf(errors.New(ErrBadPlugin), ErrLoadingPlugin, "test", "badSignature").Error(),
			},
		}

		for _, tt := range testCases {
			t.Run(tt.name, func(t *testing.T) {
				_, err := newScriptValidator(tt.baseCfg, tt.scriptCfg, "")
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
		baseCfg, err := newValidatorBaseConfig("test", "init")
		require.NoError(t, err)

		byzantinePluginFile, err := testutil.CompileGoPlugin(pluginsPath, "byzantine_validator.go")
		require.NoError(t, err)
		defer os.Remove(byzantinePluginFile)

		scriptCfg1 := &scriptConfig{
			Type: "go",
			File: pluginFile,
		}
		scriptCfg2 := &scriptConfig{
			Type: "go",
			File: byzantinePluginFile,
		}

		v1, err := newScriptValidator(baseCfg, scriptCfg1, "")
		require.NoError(t, err)
		v2, err := newScriptValidator(baseCfg, scriptCfg2, "")
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
			baseCfg   *validatorBaseConfig
			scriptCfg *scriptConfig
			link      *cs.Link
		}

		scriptConfig := &scriptConfig{
			File: pluginFile,
			Type: "go",
		}

		testCases := []testCase{
			{
				name: "has to validate",
				ret:  true,
				baseCfg: &validatorBaseConfig{
					Process:  "test",
					LinkType: "init",
				},
				link: testLink,
			},
			{
				name: "bad process",
				ret:  false,
				baseCfg: &validatorBaseConfig{
					Process:  "bad",
					LinkType: "init",
				},
				link: cstesting.RandomLinkWithProcess("test"),
			},
			{
				name: "bad type",
				ret:  false,
				baseCfg: &validatorBaseConfig{
					Process:  "test",
					LinkType: "invalid",
				},
				link: cstesting.RandomLinkWithType("test"),
			},
		}

		for _, tt := range testCases {
			t.Run(tt.name, func(t *testing.T) {
				v, err := newScriptValidator(tt.baseCfg, scriptConfig, "")
				require.NoError(t, err)
				assert.Equal(t, tt.ret, v.ShouldValidate(tt.link))
			})
		}
	})

	t.Run("Validate", func(t *testing.T) {

		type testCase struct {
			name     string
			testLink *cs.Link
			baseCfg  *validatorBaseConfig
			valid    bool
			err      string
		}

		scriptCfg := &scriptConfig{
			File: pluginFile,
			Type: "go",
		}
		testCases := []testCase{
			{
				name: "valid-link",
				baseCfg: &validatorBaseConfig{
					Process:  "test",
					LinkType: "init",
				},
				testLink: testLink,
				valid:    true,
			},
			{
				name: "fetch-link",
				baseCfg: &validatorBaseConfig{
					Process:  "test",
					LinkType: "fetchLink",
				},
				testLink: cstesting.RandomLinkWithType("fetchLink"),
				valid:    true,
			},
			{
				name: "validation-fails",
				baseCfg: &validatorBaseConfig{
					Process:  "test",
					LinkType: "invalid",
				},
				testLink: cstesting.RandomLinkWithType("invalid"),
				valid:    false,
				err:      "error",
			},
		}
		for _, tt := range testCases {
			t.Run(tt.name, func(t *testing.T) {
				sv, err := newScriptValidator(tt.baseCfg, scriptCfg, "")
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
