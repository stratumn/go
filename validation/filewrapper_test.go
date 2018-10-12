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
	"context"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-chainscript/chainscripttest"
	"github.com/stratumn/go-core/dummystore"
	"github.com/stratumn/go-core/testutil"
	"github.com/stratumn/go-core/utils"
	"github.com/stratumn/go-core/validation"
	"github.com/stratumn/go-core/validation/validationtesting"
	"github.com/stratumn/go-core/validation/validators"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStoreWithConfigFile(t *testing.T) {
	testStore := dummystore.New(nil)

	t.Run("without configuration file", func(t *testing.T) {
		t.Run("invalid link", func(t *testing.T) {
			va, err := validation.WrapStoreWithConfigFile(testStore, &validation.Config{})
			require.NoError(t, err)

			link := chainscripttest.NewLinkBuilder(t).WithProcess("").Build()
			_, err = va.CreateLink(context.Background(), link)
			testutil.AssertWrappedErrorEqual(t, err, chainscript.ErrMissingProcess)
		})

		t.Run("valid link", func(t *testing.T) {
			va, err := validation.WrapStoreWithConfigFile(testStore, &validation.Config{})
			require.NoError(t, err)

			link := chainscripttest.NewLinkBuilder(t).WithRandomData().Build()
			_, err = va.CreateLink(context.Background(), link)
			assert.NoError(t, err)
		})
	})

	t.Run("with configuration file", func(t *testing.T) {
		rules := utils.CreateTempFile(t, validationtesting.TestJSONRules)
		defer os.Remove(rules)

		testValidationStore, err := validation.WrapStoreWithConfigFile(
			testStore,
			&validation.Config{RulesPath: rules},
		)
		require.NoError(t, err)
		require.NotNil(t, testValidationStore)

		t.Run("invalid link", func(t *testing.T) {
			link := chainscripttest.NewLinkBuilder(t).
				WithProcess("").
				Build()

			_, err := testValidationStore.CreateLink(context.Background(), link)
			testutil.AssertWrappedErrorEqual(t, err, chainscript.ErrMissingProcess)
		})

		t.Run("missing parent", func(t *testing.T) {
			link := chainscripttest.NewLinkBuilder(t).
				WithProcess("chat").
				WithParentHash(chainscripttest.RandomHash()).
				Build()

			_, err := testValidationStore.CreateLink(context.Background(), link)
			testutil.AssertWrappedErrorEqual(t, err, validators.ErrParentNotFound)
		})

		t.Run("invalid transition", func(t *testing.T) {
			ctx := context.Background()
			init := chainscripttest.NewLinkBuilder(t).
				WithProcess("chat").
				WithStep("init").
				WithSignatureFromKey(t, []byte(validationtesting.BobPrivateKey), "").
				Build()

			_, err := testValidationStore.CreateLink(ctx, init)
			require.NoError(t, err)

			init2 := chainscripttest.NewLinkBuilder(t).
				WithProcess("chat").
				WithStep("init").
				WithParent(t, init).
				WithSignatureFromKey(t, []byte(validationtesting.BobPrivateKey), "").
				Build()

			_, err = testValidationStore.CreateLink(ctx, init2)
			testutil.AssertWrappedErrorEqual(t, err, validators.ErrInvalidTransition)
		})

		t.Run("invalid signature", func(t *testing.T) {
			init := chainscripttest.NewLinkBuilder(t).
				WithProcess("chat").
				WithStep("init").
				WithSignatureFromKey(t, []byte(validationtesting.AlicePrivateKey), "").
				Build()

			_, err := testValidationStore.CreateLink(context.Background(), init)
			testutil.AssertWrappedErrorEqual(t, err, validators.ErrMissingSignature)
		})

		t.Run("invalid schema", func(t *testing.T) {
			init := chainscripttest.NewLinkBuilder(t).
				WithProcess("auction").
				WithStep("init").
				WithData(t, map[string]string{
					"seller": "alice",
				}).
				WithSignatureFromKey(t, []byte(validationtesting.AlicePrivateKey), "").
				Build()

			_, err := testValidationStore.CreateLink(context.Background(), init)
			testutil.AssertWrappedErrorEqual(t, err, validators.ErrInvalidLinkSchema)
		})

		t.Run("invalid configuration file", func(t *testing.T) {
			invalidRules := utils.CreateTempFile(t, "validate all the things")
			defer os.Remove(invalidRules)

			v, err := validation.WrapStoreWithConfigFile(
				testStore,
				&validation.Config{RulesPath: invalidRules},
			)

			assert.Nil(t, v)
			assert.Error(t, err)
		})

		t.Run("config file update", func(t *testing.T) {
			ctx := context.Background()
			rules := utils.CreateTempFile(t, `{
				"drivers": {
				  "steps": {
					"add": {
					  "schema": {
						"type": "object",
						"properties": {
						  "name": {
							"type": "string"
						  },
						  "age": {
							  "type": "string"
						  }
						},
						"required": ["name", "age"]
					  }
					}
				  }
				}
			  }`)
			defer os.Remove(rules)

			v, err := validation.WrapStoreWithConfigFile(
				testStore,
				&validation.Config{RulesPath: rules},
			)
			require.NoError(t, err)
			require.NotNil(t, v)

			link := chainscripttest.NewLinkBuilder(t).
				WithProcess("drivers").
				WithStep("add").
				WithData(t, map[string]interface{}{
					"name": "ryan",
					"age":  33,
				}).
				Build()

			_, err = v.CreateLink(ctx, link)
			testutil.AssertWrappedErrorEqual(t, err, validators.ErrInvalidLinkSchema)

			err = ioutil.WriteFile(rules, []byte(`{
				"drivers": {
				  "steps": {
					"add": {
					  "schema": {
						"type": "object",
						"properties": {
						  "name": {
							"type": "string"
						  },
						  "age": {
							  "type": "integer"
						  }
						},
						"required": ["name", "age"]
					  }
					}
				  }
				}
			  }`), os.ModePerm)
			require.NoError(t, err)

			// Since we don't control the filewatcher's signals, we have no
			// other solution than waiting a bit for the write to be flushed
			// properly to disk.
			<-time.After(50 * time.Millisecond)

			_, err = v.CreateLink(ctx, link)
			assert.NoError(t, err)
		})
	})
}
