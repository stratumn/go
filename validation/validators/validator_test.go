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
	"fmt"
	"os"
	"testing"

	"github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-chainscript/chainscripttest"
	"github.com/stratumn/go-core/dummystore"
	"github.com/stratumn/go-core/store"
	"github.com/stratumn/go-core/testutil"
	"github.com/stratumn/go-core/utils"
	"github.com/stratumn/go-core/validation"
	"github.com/stratumn/go-core/validation/testutils"
	"github.com/stratumn/go-core/validation/validators"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testCase struct {
	name  string
	link  *chainscript.Link
	valid bool
}

var pluginFile string

const (
	pluginsPath      = "../testutils/plugins"
	pluginSourceFile = "custom_validator.go"
)

func TestMain(m *testing.M) {
	var res int
	defer os.Exit(res)

	var err error
	pluginFile, err = testutil.CompileGoPlugin(pluginsPath, pluginSourceFile)
	if err != nil {
		fmt.Println("could not launch validator tests: error while compiling validation plugin")
		os.Exit(2)
	}
	defer os.Remove(pluginFile)

	res = m.Run()
}

func initTestCases(t *testing.T) (store.Adapter, []testCase) {
	store := dummystore.New(nil)
	data := map[string]interface{}{
		"buyer":        "alice",
		"seller":       "bob",
		"lot":          "painting",
		"initialPrice": 12,
	}

	initAuctionLink := chainscripttest.NewLinkBuilder(t).
		WithProcess("auction").
		WithStep("init").
		WithoutParent().
		WithData(t, data).
		WithSignatureFromKey(t, []byte(testutils.AlicePrivateKey), "").
		Build()

	initAuctionLinkHash, err := store.CreateLink(context.Background(), initAuctionLink)
	require.NoError(t, err)

	var testCases = []testCase{{
		name:  "valid-init-link",
		link:  initAuctionLink,
		valid: true,
	}, {
		name: "valid-link",
		link: chainscripttest.NewLinkBuilder(t).
			WithData(t, map[string]interface{}{
				"buyer":    "alice",
				"bidPrice": 42,
			}).
			WithParentHash(initAuctionLinkHash).
			WithProcess("auction").
			WithStep("bid").
			Build(),
		valid: true,
	}, {
		name:  "no-validator-match",
		link:  chainscripttest.NewLinkBuilder(t).WithProcess("auction").WithStep("unknown").Build(),
		valid: false,
	}, {
		name: "missing-required-field",
		link: chainscripttest.NewLinkBuilder(t).
			WithData(t, map[string]interface{}{
				"to": "bob",
			}).
			WithProcess("chat").
			WithStep("message").
			Build(),
		valid: false,
	}, {
		name: "invalid-field-value",
		link: chainscripttest.NewLinkBuilder(t).
			WithData(t, map[string]interface{}{
				"buyer":    "alice",
				"bidPrice": -10,
			}).
			WithProcess("auction").
			WithStep("bid").
			Build(),
		valid: false,
	}}
	return store, testCases
}

func TestValidator(t *testing.T) {
	testFile := utils.CreateTempFile(t, testutils.ValidJSONConfig)
	defer os.Remove(testFile)

	children, err := validation.LoadConfig(&validation.Config{
		RulesPath:   testFile,
		PluginsPath: pluginsPath,
	}, nil)
	require.NoError(t, err, "LoadConfig()")

	v := validators.NewMultiValidator(children)

	store, testCases := initTestCases(t)
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Validate(context.Background(), store, tt.link)
			if tt.valid {
				assert.NoError(t, err, "v.Validate()")
			} else {
				assert.Error(t, err, "v.Validate()")
			}
		})
	}
}
