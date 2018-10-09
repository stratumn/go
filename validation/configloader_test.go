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
	"os"
	"testing"

	"github.com/stratumn/go-core/utils"
	"github.com/stratumn/go-core/validation"
	"github.com/stratumn/go-core/validation/validationtesting"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadFromFile(t *testing.T) {
	t.Run("missing file", func(t *testing.T) {
		v, err := validation.LoadFromFile(context.Background(), &validation.Config{
			RulesPath: "/tmp/validation/missing-rules.json",
		})

		assert.Nil(t, v)
		assert.Error(t, err)
	})

	t.Run("invalid file content", func(t *testing.T) {
		testFile := utils.CreateTempFile(t, "validatie-wixie kawaii rules-y")
		defer os.Remove(testFile)

		v, err := validation.LoadFromFile(context.Background(), &validation.Config{
			RulesPath: testFile,
		})

		assert.Nil(t, v)
		assert.Error(t, err)
	})

	t.Run("valid configuration file", func(t *testing.T) {
		testFile := utils.CreateTempFile(t, validationtesting.TestJSONRules)
		defer os.Remove(testFile)

		v, err := validation.LoadFromFile(context.Background(), &validation.Config{
			RulesPath: testFile,
		})

		require.NoError(t, err)
		require.NotNil(t, v)

		require.Len(t, v, 2)

		auctionValidators, ok := v["auction"]
		require.True(t, ok)
		assert.Len(t, auctionValidators, 5)

		chatValidators, ok := v["chat"]
		require.True(t, ok)
		assert.Len(t, chatValidators, 4)
	})
}
