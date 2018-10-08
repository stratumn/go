// Copyright 2017-2018 Stratumn SAS. All rights reserved.
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
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stratumn/go-chainscript/chainscripttest"
	"github.com/stratumn/go-core/validation/validationtesting"
	"github.com/stratumn/go-core/validation/validators"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScriptValidator(t *testing.T) {
	testLink := chainscripttest.NewLinkBuilder(t).
		WithProcess("test").
		WithStep("init").
		Build()

	t.Run("New", func(t *testing.T) {
		t.Run("missing plugin file", func(t *testing.T) {
			_, err := validators.NewScriptValidator(
				"test_process",
				"/var/tmp/",
				&validators.ScriptConfig{Hash: "42"},
			)

			assert.Equal(t, 0, strings.Index(err.Error(), validators.ErrLoadingPlugin.Error()))
		})

		t.Run("invalid plugin hash", func(t *testing.T) {
			pluginsDir, _ := validationtesting.CompilePlugin(t, validationtesting.PluginValidationSuccess)

			_, err := validators.NewScriptValidator(
				"test_process",
				pluginsDir,
				&validators.ScriptConfig{Hash: "not 42"},
			)

			assert.Equal(t, 0, strings.Index(err.Error(), validators.ErrLoadingPlugin.Error()))
		})

		t.Run("invalid plugin file format", func(t *testing.T) {
			pluginsDir, err := ioutil.TempDir("", "plugins")
			require.NoError(t, err)

			pluginHash := sha256.Sum256([]byte("this is not a valid go plugin's compiled content"))
			err = ioutil.WriteFile(
				filepath.Join(pluginsDir, fmt.Sprintf("%s.so", pluginHash)),
				[]byte("this is not a valid go plugin's compiled content"),
				os.ModePerm,
			)
			require.NoError(t, err)

			_, err = validators.NewScriptValidator(
				"test_process",
				pluginsDir,
				&validators.ScriptConfig{Hash: hex.EncodeToString(pluginHash[:])},
			)

			assert.Equal(t, 0, strings.Index(err.Error(), validators.ErrLoadingPlugin.Error()))
		})

		t.Run("invalid plugin entry point", func(t *testing.T) {
			pluginsDir, pluginHash := validationtesting.CompilePlugin(t, validationtesting.PluginMissingValidate)

			_, err := validators.NewScriptValidator(
				"test_process",
				pluginsDir,
				&validators.ScriptConfig{Hash: hex.EncodeToString(pluginHash)},
			)

			assert.Equal(t, 0, strings.Index(err.Error(), validators.ErrInvalidPlugin.Error()))
		})

		t.Run("invalid plugin signature", func(t *testing.T) {
			pluginsDir, pluginHash := validationtesting.CompilePlugin(t, validationtesting.PluginInvalidValidate)

			_, err := validators.NewScriptValidator(
				"test_process",
				pluginsDir,
				&validators.ScriptConfig{Hash: hex.EncodeToString(pluginHash)},
			)

			assert.Equal(t, 0, strings.Index(err.Error(), validators.ErrInvalidPlugin.Error()))
		})

		t.Run("valid plugin", func(t *testing.T) {
			pluginsDir, pluginHash := validationtesting.CompilePlugin(t, validationtesting.PluginValidationSuccess)

			v, err := validators.NewScriptValidator(
				"test_process",
				pluginsDir,
				&validators.ScriptConfig{Hash: hex.EncodeToString(pluginHash)},
			)

			require.NoError(t, err)
			assert.NotNil(t, v)
		})
	})

	t.Run("Hash", func(t *testing.T) {
		t.Run("depends on process", func(t *testing.T) {
			pluginsDir, pluginHash := validationtesting.CompilePlugin(t, validationtesting.PluginValidationSuccess)
			v1, err := validators.NewScriptValidator(
				"p1",
				pluginsDir,
				&validators.ScriptConfig{Hash: hex.EncodeToString(pluginHash)},
			)
			require.NoError(t, err)
			h1, _ := v1.Hash()

			v2, err := validators.NewScriptValidator(
				"p2",
				pluginsDir,
				&validators.ScriptConfig{Hash: hex.EncodeToString(pluginHash)},
			)
			require.NoError(t, err)
			h2, _ := v2.Hash()

			assert.NotEqual(t, h1, h2)
		})

		t.Run("depends on script", func(t *testing.T) {
			pluginsDir1, pluginHash1 := validationtesting.CompilePlugin(t, validationtesting.PluginValidationSuccess)
			v1, err := validators.NewScriptValidator(
				"test_process",
				pluginsDir1,
				&validators.ScriptConfig{Hash: hex.EncodeToString(pluginHash1)},
			)
			require.NoError(t, err)
			h1, _ := v1.Hash()

			pluginsDir2, pluginHash2 := validationtesting.CompilePlugin(t, validationtesting.PluginValidationError)
			v2, err := validators.NewScriptValidator(
				"test_process",
				pluginsDir2,
				&validators.ScriptConfig{Hash: hex.EncodeToString(pluginHash2)},
			)
			require.NoError(t, err)
			h2, _ := v2.Hash()

			assert.NotEqual(t, h1, h2)
		})
	})

	t.Run("ShouldValidate", func(t *testing.T) {
		pluginsDir, pluginHash := validationtesting.CompilePlugin(t, validationtesting.PluginValidationSuccess)

		t.Run("process mismatch", func(t *testing.T) {
			v, err := validators.NewScriptValidator(
				"some_random_process",
				pluginsDir,
				&validators.ScriptConfig{Hash: hex.EncodeToString(pluginHash)},
			)

			require.NoError(t, err)
			assert.False(t, v.ShouldValidate(testLink))
		})

		t.Run("process match", func(t *testing.T) {
			v, err := validators.NewScriptValidator(
				testLink.Meta.Process.Name,
				pluginsDir,
				&validators.ScriptConfig{Hash: hex.EncodeToString(pluginHash)},
			)

			require.NoError(t, err)
			assert.True(t, v.ShouldValidate(testLink))
		})
	})

	t.Run("Validate", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			pluginsDir, pluginHash := validationtesting.CompilePlugin(t, validationtesting.PluginValidationSuccess)
			v, err := validators.NewScriptValidator(
				testLink.Meta.Process.Name,
				pluginsDir,
				&validators.ScriptConfig{Hash: hex.EncodeToString(pluginHash)},
			)
			require.NoError(t, err)

			err = v.Validate(context.Background(), nil, testLink)
			assert.NoError(t, err)
		})

		t.Run("failure", func(t *testing.T) {
			pluginsDir, pluginHash := validationtesting.CompilePlugin(t, validationtesting.PluginValidationError)
			v, err := validators.NewScriptValidator(
				testLink.Meta.Process.Name,
				pluginsDir,
				&validators.ScriptConfig{Hash: hex.EncodeToString(pluginHash)},
			)
			require.NoError(t, err)

			err = v.Validate(context.Background(), nil, testLink)
			assert.Error(t, err)
		})
	})
}
