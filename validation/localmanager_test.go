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

package validation_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stratumn/go-indigocore/dummystore"
	"github.com/stratumn/go-indigocore/store/storetesting"
	"github.com/stratumn/go-indigocore/utils"
	"github.com/stratumn/go-indigocore/validation"
	"github.com/stratumn/go-indigocore/validation/testutils"
	"github.com/stratumn/go-indigocore/validation/validators"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalManager(t *testing.T) {

	t.Run("New", func(t *testing.T) {
		t.Run("Governance without file", func(t *testing.T) {
			var v validators.Validator
			a := new(storetesting.MockAdapter)
			gov, err := validation.NewLocalManager(context.Background(), a, &validation.Config{})
			assert.NoError(t, err, "Gouvernance is initialized by store")
			assert.NotNil(t, gov, "Gouvernance is initialized by store")

			v = gov.Current()

			assert.Nil(t, v, "No validator loaded")
		})

		t.Run("Governance without file but store", func(t *testing.T) {
			var v validators.Validator
			a := dummystore.New(nil)
			populateStoreWithValidData(t, a)
			gov, err := validation.NewLocalManager(context.Background(), a, &validation.Config{
				PluginsPath: pluginsPath,
			})
			assert.NoError(t, err, "Gouvernance is initialized by store")
			require.NotNil(t, gov, "Gouvernance is initialized by store")

			v = gov.Current()
			assert.NotNil(t, v, "Validator loaded from store")
		})

		t.Run("Governance with valid file", func(t *testing.T) {
			var v validators.Validator
			a := dummystore.New(nil)
			testFile := utils.CreateTempFile(t, testutils.ValidJSONConfig)
			defer os.Remove(testFile)

			gov, err := validation.NewLocalManager(context.Background(), a, &validation.Config{
				RulesPath:   testFile,
				PluginsPath: pluginsPath,
			})
			assert.NoError(t, err, "Gouvernance is initialized by file and store")
			assert.NotNil(t, gov, "Gouvernance is initialized by file and store")

			v = gov.Current()

			assert.NotNil(t, v, "Validator loaded from file")
		})

		t.Run("Governance with invalid file", func(t *testing.T) {
			a := new(storetesting.MockAdapter)
			gov, err := validation.NewLocalManager(context.Background(), a, &validation.Config{
				RulesPath: "localmanager_test.go",
			})
			assert.EqualError(t, err, "Cannot load validator rules file localmanager_test.go: invalid character '/' looking for beginning of value")
			require.NotNil(t, gov, "Gouvernance is initialized by store")
		})

		t.Run("Governance with unexisting file", func(t *testing.T) {
			a := new(storetesting.MockAdapter)
			gov, err := validation.NewLocalManager(context.Background(), a, &validation.Config{
				RulesPath: "foo/bar",
			})
			assert.Error(t, err, "Cannot initialize gouvernance with bad file")
			assert.Nil(t, gov, "Cannot initialize gouvernance with bad file")
		})

		t.Run("New validator uploaded at startup", func(t *testing.T) {
			var v validators.Validator
			a := dummystore.New(nil)
			populateStoreWithValidData(t, a)
			checkLastValidatorPriority(t, a, "auction", 1.)
			testFile := utils.CreateTempFile(t, testutils.ValidJSONConfig)
			defer os.Remove(testFile)
			gov, err := validation.NewLocalManager(context.Background(), a, &validation.Config{
				RulesPath:   testFile,
				PluginsPath: pluginsPath,
			})
			require.NotNil(t, gov, "Gouvernance is initialized by file and store")
			assert.NoError(t, err, "Validator updated")

			v = gov.Current()
			assert.NotNil(t, v, "Validator loaded from file")
			checkLastValidatorPriority(t, a, "auction", 2.)
		})
	})

	t.Run("ListenAndUpdate", func(t *testing.T) {
		t.Run("Update rules in store on file update", func(t *testing.T) {
			var v validators.Validator
			ctx := context.Background()
			validJSON := fmt.Sprintf(`{%s}`, testutils.ValidChatJSONConfig)
			a := dummystore.New(nil)
			testFile := utils.CreateTempFile(t, validJSON)
			defer os.Remove(testFile)
			gov, err := validation.NewLocalManager(ctx, a, &validation.Config{
				RulesPath: testFile,
			})
			require.NotNil(t, gov, "Gouvernance is initialized by file and store")
			go gov.ListenAndUpdate(ctx)
			waitValidator := gov.Subscribe()
			v = <-waitValidator
			assert.NotNil(t, v, "Validator loaded from file")

			checkLastValidatorPriority(t, a, "chat", 0.)

			chatJSON := testutils.CreateValidatorJSON("chat",
				strings.Replace(testutils.ValidChatJSONPKIConfig, "Bob", "Dave", -1),
				testutils.ValidChatJSONTypesConfig)
			validJSON = fmt.Sprintf(`{%s}`, chatJSON)
			f, err := os.OpenFile(testFile, os.O_WRONLY, 0)
			require.NoErrorf(t, err, "cannot modify file %s", &validation.Config{
				RulesPath: testFile,
			})
			defer f.Close()
			_, err = f.WriteString(validJSON)
			require.NoError(t, err, "tmpfile.WriteString()")

			v = <-waitValidator
			assert.NotNil(t, v, "Validator reloaded from file")

			checkLastValidatorPriority(t, a, "chat", 1.)
		})

		t.Run("closes subscribing channels on context cancel", func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())

			testFile := utils.CreateTempFile(t, testutils.ValidJSONConfig)
			defer os.Remove(testFile)
			gov, err := validation.NewLocalManager(ctx, dummystore.New(nil), &validation.Config{
				RulesPath:   testFile,
				PluginsPath: pluginsPath,
			})
			require.NoError(t, err)

			done := make(chan struct{})
			go func() {
				require.EqualError(t, gov.ListenAndUpdate(ctx), context.Canceled.Error())
				done <- struct{}{}
			}()
			cancel()
			<-done
		})

		t.Run("return an error when no file watcher is set", func(t *testing.T) {
			ctx := context.Background()
			testFile := utils.CreateTempFile(t, "")
			defer os.Remove(testFile)
			gov, err := validation.NewLocalManager(ctx, dummystore.New(nil), &validation.Config{})
			require.NoError(t, err)

			done := make(chan struct{})
			go func() {
				require.EqualError(t, gov.ListenAndUpdate(ctx), validation.ErrNoFileWatcher.Error())
				done <- struct{}{}
			}()
			<-done
		})
	})

	t.Run("Current", func(t *testing.T) {
		t.Run("returns the current validator set", func(t *testing.T) {
			ctx := context.Background()
			testFile := utils.CreateTempFile(t, "{}")
			defer os.Remove(testFile)
			gov, err := validation.NewLocalManager(ctx, dummystore.New(nil), &validation.Config{
				RulesPath:   testFile,
				PluginsPath: pluginsPath,
			})
			require.NoError(t, err)

			go gov.ListenAndUpdate(ctx)
			assert.Nil(t, gov.Current())

			ioutil.WriteFile(testFile, []byte(testutils.ValidJSONConfig), os.ModeTemporary)
			newValidator := <-gov.Subscribe()
			assert.Equal(t, newValidator, gov.Current())
		})
	})
}
