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

package generator

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGeneratorFromFile_(t *testing.T) {
	vars := map[string]interface{}{
		"os": runtime.GOOS,
	}
	gen, err := NewDefinitionFromFile("testdata/nodejs/generator.json", vars, nil)
	require.NoError(t, err, "NewDefinitionFromFile()")

	t.Run("checkVariables", func(t *testing.T) {
		got, ok := gen.Variables["os"]
		require.True(t, ok, `gen.Variables["os"]`)
		assert.Equal(t, runtime.GOOS, got, `gen.Variables["os"]`)
	})

	t.Run("checkStringInput", func(t *testing.T) {
		got, ok := gen.Inputs["name"]
		require.True(t, ok, `gen.Inputs["name"]`)
		assert.IsType(t, &StringInput{}, got, `gen.Inputs["name"] should be an StringInput`)
		input, _ := got.(*StringInput)
		assert.Equal(t, "Project name", input.Prompt, `input.Prompt`)
		assert.Equal(t, ".+", input.Format, `input.Format`)
	})

	t.Run("checkIntInput", func(t *testing.T) {
		got, ok := gen.Inputs["nodes"]
		require.True(t, ok, `gen.Inputs["nodes"]`)
		assert.IsType(t, &IntInput{}, got, `bad type for gen.Inputs["nodes"]`)
		input, _ := got.(*IntInput)
		assert.Equal(t, "Number of nodes", input.Prompt, `input.Prompt`)
		assert.Equal(t, 4, input.Default, `input.Format`)
	})

	t.Run("checkSelectInput", func(t *testing.T) {
		got, ok := gen.Inputs["license"]
		require.True(t, ok, `gen.Inputs["license"]`)
		assert.IsType(t, &StringSelect{}, got, `bad type for gen.Inputs["license"]`)
		input, _ := got.(*StringSelect)
		assert.Equal(t, "License", input.Prompt, `input.Prompt`)
		assert.Len(t, input.Options, 3, `input.Options`)
		assert.Equal(t, "apache", input.Default, `input.Default`)
	})

	t.Run("checkSelectMultiInput", func(t *testing.T) {
		got, ok := gen.Inputs["fossilizer"]
		require.True(t, ok, `gen.Inputs["fossilizer"]`)
		assert.IsType(t, &StringSelectMulti{}, got, `bad type for gen.Inputs["fossilizer"]`)
		input, _ := got.(*StringSelectMulti)
		assert.Equal(t, "List of fossilizers", input.Prompt, `input.Prompt`)
		assert.Len(t, input.Options, 2, `input.Options`)
		assert.True(t, input.IsRequired, `input.IsRequired`)
	})

	t.Run("checkSliceInput", func(t *testing.T) {
		got, ok := gen.Inputs["process"]
		require.True(t, ok, `gen.Inputs["process"]`)
		assert.IsType(t, &StringSlice{}, got, `bad type for gen.Inputs["process"]`)
		input, _ := got.(*StringSlice)
		assert.Equal(t, "List of process names", input.Prompt, `input.Prompt`)
		assert.Equal(t, "^[a-zA-Z].*$", input.Format, `input.Format`)
	})
}

func TestNewFromDir(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		gen, err := NewFromDir("testdata/nodejs", &Options{})
		require.NoError(t, err)
		assert.NotNil(t, gen)
	})

	t.Run("notExist", func(t *testing.T) {
		_, err := NewFromDir("testdata/404", &Options{})
		require.Error(t, err)
	})

	t.Run("invalidDef", func(t *testing.T) {
		_, err := NewFromDir("testdata/invalid_def", &Options{})
		require.Error(t, err)
	})

	t.Run("invalidDefExec", func(t *testing.T) {
		_, err := NewFromDir("testdata/invalid_def_exec", &Options{})
		require.Error(t, err)
	})

	t.Run("invalidDefTpml", func(t *testing.T) {
		_, err := NewFromDir("testdata/custom_funcs", &Options{})
		require.Error(t, err)
	})
}
