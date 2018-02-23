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

// func TestNewFromDir(t *testing.T) {
// 	gen, err := NewFromDir("testdata/nodejs", &Options{})
// 	if err != nil {
// 		t.Fatalf("err: NewFromDir(): %s", err)
// 	}
// 	if gen == nil {
// 		t.Fatal("err: gen = nil want *Generator")
// 	}
// }

// func TestNewFromDir_notExist(t *testing.T) {
// 	_, err := NewFromDir("testdata/404", &Options{})
// 	if err == nil {
// 		t.Error("err: err = nil want Error")
// 	}
// }

// func TestNewFromDir_invalidDef(t *testing.T) {
// 	_, err := NewFromDir("testdata/invalid_def", &Options{})
// 	if err == nil {
// 		t.Error("err: err = nil want Error")
// 	}
// }

// func TestNewFromDir_invalidDefExec(t *testing.T) {
// 	_, err := NewFromDir("testdata/invalid_def_exec", &Options{})
// 	if err == nil {
// 		t.Error("err: err = nil want Error")
// 	}
// }

// func TestNewFromDir_invalidDefTpml(t *testing.T) {
// 	_, err := NewFromDir("testdata/custom_funcs", &Options{})
// 	if err == nil {
// 		t.Error("err: err = nil want Error")
// 	}
// }

// func TestGeneratorExec(t *testing.T) {
// 	dst, err := ioutil.TempDir("", "generator")
// 	if err != nil {
// 		t.Fatalf("err: ioutil.TempDir(): %s", err)
// 	}
// 	defer os.RemoveAll(dst)

// 	r := strings.NewReader("test\nTest project\nStephan Florquin\n2017\nStratumn\n2\nProcess1,Process2\n")

// 	gen, err := NewFromDir("testdata/nodejs", &Options{Reader: r})
// 	if err != nil {
// 		t.Fatalf("err: NewFromDir(): %s", err)
// 	}

// 	if err := gen.Exec(dst); err != nil {
// 		t.Fatalf("err: gen.Exec(): %s", err)
// 	}

// 	cmpWalk(t, "testdata/nodejs_expected", dst, "testdata/nodejs_expected")
// }

// func TestGeneratorExec_ask(t *testing.T) {
// 	dst, err := ioutil.TempDir("", "generator")
// 	if err != nil {
// 		t.Fatalf("err: ioutil.TempDir(): %s", err)
// 	}
// 	defer os.RemoveAll(dst)

// 	r := strings.NewReader("\n\nTest Project\n")

// 	gen, err := NewFromDir("testdata/ask", &Options{Reader: r})
// 	if err != nil {
// 		t.Fatalf("err: NewFromDir(): %s", err)
// 	}

// 	if err := gen.Exec(dst); err != nil {
// 		t.Fatalf("err: gen.Exec(): %s", err)
// 	}

// 	cmpWalk(t, "testdata/ask_expected", dst, "testdata/ask_expected")
// }

// func TestGeneratorExec_tmplVars(t *testing.T) {
// 	dst, err := ioutil.TempDir("", "generator")
// 	if err != nil {
// 		t.Fatalf("err: ioutil.TempDir(): %s", err)
// 	}
// 	defer os.RemoveAll(dst)

// 	opts := Options{
// 		TmplVars: map[string]interface{}{"test": "hello"},
// 	}

// 	gen, err := NewFromDir("testdata/vars", &opts)
// 	if err != nil {
// 		t.Fatalf("err: NewFromDir(): %s", err)
// 	}

// 	if err := gen.Exec(dst); err != nil {
// 		t.Fatalf("err: gen.Exec(): %s", err)
// 	}

// 	cmpWalk(t, "testdata/vars_expected", dst, "testdata/vars_expected")
// }

// func TestGeneratorExec_customFuncs(t *testing.T) {
// 	dst, err := ioutil.TempDir("", "generator")
// 	if err != nil {
// 		t.Fatalf("err: ioutil.TempDir(): %s", err)
// 	}
// 	defer os.RemoveAll(dst)

// 	opts := Options{
// 		DefFuncs: map[string]interface{}{
// 			"custom": func() string { return "hello generator" },
// 		},
// 		TmplFuncs: map[string]interface{}{
// 			"custom": func() string { return "hello template" },
// 		},
// 	}

// 	gen, err := NewFromDir("testdata/custom_funcs", &opts)
// 	if err != nil {
// 		t.Fatalf("err: NewFromDir(): %s", err)
// 	}

// 	if err := gen.Exec(dst); err != nil {
// 		t.Fatalf("err: gen.Exec(): %s", err)
// 	}

// 	cmpWalk(t, "testdata/custom_funcs_expected", dst, "testdata/custom_funcs_expected")
// }

// func TestGeneratorExec_inputError(t *testing.T) {
// 	dst, err := ioutil.TempDir("", "generator")
// 	if err != nil {
// 		t.Fatalf("err: ioutil.TempDir(): %s", err)
// 	}
// 	defer os.RemoveAll(dst)

// 	gen, err := NewFromDir("testdata/nodejs", &Options{})
// 	if err != nil {
// 		t.Fatalf("err: NewFromDir(): %s", err)
// 	}

// 	if err := gen.Exec(dst); err == nil {
// 		t.Error("err: err = nil want Error")
// 	}
// }

// func TestGeneratorExec_askError(t *testing.T) {
// 	dst, err := ioutil.TempDir("", "generator")
// 	if err != nil {
// 		t.Fatalf("err: ioutil.TempDir(): %s", err)
// 	}
// 	defer os.RemoveAll(dst)

// 	gen, err := NewFromDir("testdata/ask", &Options{})
// 	if err != nil {
// 		t.Fatalf("err: NewFromDir(): %s", err)
// 	}

// 	if err := gen.Exec(dst); err == nil {
// 		t.Error("err: err = nil want Error")
// 	}
// }

// func TestGeneratorExec_askInvalid(t *testing.T) {
// 	dst, err := ioutil.TempDir("", "generator")
// 	if err != nil {
// 		t.Fatalf("err: ioutil.TempDir(): %s", err)
// 	}
// 	defer os.RemoveAll(dst)

// 	gen, err := NewFromDir("testdata/ask_invalid", &Options{})
// 	if err != nil {
// 		t.Fatalf("err: NewFromDir(): %s", err)
// 	}

// 	if err := gen.Exec(dst); err == nil {
// 		t.Error("err: err = nil want Error")
// 	}
// }

// func TestGeneratorExec_invalidTpml(t *testing.T) {
// 	dst, err := ioutil.TempDir("", "generator")
// 	if err != nil {
// 		t.Fatalf("err: ioutil.TempDir(): %s", err)
// 	}
// 	defer os.RemoveAll(dst)

// 	gen, err := NewFromDir("testdata/invalid_tmpl", &Options{})
// 	if err != nil {
// 		t.Fatalf("err: NewFromDir(): %s", err)
// 	}

// 	if err := gen.Exec(dst); err == nil {
// 		t.Error("err: err = nil want Error")
// 	}
// }

// func TestGeneratorExec_invalidPartial(t *testing.T) {
// 	dst, err := ioutil.TempDir("", "generator")
// 	if err != nil {
// 		t.Fatalf("err: ioutil.TempDir(): %s", err)
// 	}
// 	defer os.RemoveAll(dst)

// 	r := strings.NewReader("test\nTest project\nStephan Florquin\n2017\nStratumn\n2\nProcess1,Process2\n")

// 	gen, err := NewFromDir("testdata/invalid_partial", &Options{Reader: r})
// 	if err != nil {
// 		t.Fatalf("err: NewFromDir(): %s", err)
// 	}

// 	if err := gen.Exec(dst); err == nil {
// 		t.Error("err: err = nil want Error")
// 	}
// }

// func TestGeneratorExec_invalidPartialExec(t *testing.T) {
// 	dst, err := ioutil.TempDir("", "generator")
// 	if err != nil {
// 		t.Fatalf("err: ioutil.TempDir(): %s", err)
// 	}
// 	defer os.RemoveAll(dst)

// 	r := strings.NewReader("test\nTest project\nStephan Florquin\n2017\nStratumn\n2\nProcess1,Process2\n")

// 	gen, err := NewFromDir("testdata/invalid_partial_exec", &Options{Reader: r})
// 	if err != nil {
// 		t.Fatalf("err: NewFromDir(): %s", err)
// 	}

// 	if err := gen.Exec(dst); err == nil {
// 		t.Error("err: err = nil want Error")
// 	}
// }

// func TestGeneratorExec_undefinedInput(t *testing.T) {
// 	dst, err := ioutil.TempDir("", "generator")
// 	if err != nil {
// 		t.Fatalf("err: ioutil.TempDir(): %s", err)
// 	}
// 	defer os.RemoveAll(dst)

// 	gen, err := NewFromDir("testdata/undefined_input", &Options{})
// 	if err != nil {
// 		t.Fatalf("err: NewFromDir(): %s", err)
// 	}

// 	if err := gen.Exec(dst); err == nil {
// 		t.Error("err: err = nil want Error")
// 	}
// }

// func TestGeneratorExec_invalidPartialArgs(t *testing.T) {
// 	dst, err := ioutil.TempDir("", "generator")
// 	if err != nil {
// 		t.Fatalf("err: ioutil.TempDir(): %s", err)
// 	}
// 	defer os.RemoveAll(dst)

// 	gen, err := NewFromDir("testdata/invalid_partial_args", &Options{})
// 	if err != nil {
// 		t.Fatalf("err: NewFromDir(): %s", err)
// 	}

// 	if err := gen.Exec(dst); err == nil {
// 		t.Error("err: err = nil want Error")
// 	}
// }

// func TestGeneratorExec_filenameSubstitution(t *testing.T) {
// 	dst, err := ioutil.TempDir("", "generator")
// 	if err != nil {
// 		t.Fatalf("err: ioutil.TempDir(): %s", err)
// 	}
// 	defer os.RemoveAll(dst)

// 	r := strings.NewReader("Process1,Process2\nTheTest\n")

// 	gen, err := NewFromDir("testdata/filename_subst", &Options{Reader: r})
// 	if err != nil {
// 		t.Fatalf("err: NewFromDir(): %s", err)
// 	}

// 	if err := gen.Exec(dst); err != nil {
// 		t.Fatalf("err: gen.Exec(): %s", err)
// 	}

// 	if _, err := os.Stat(path.Join(dst, "file-Process1.js")); err != nil {
// 		t.Errorf("err: %s", err.Error())
// 	}

// 	if _, err := os.Stat(path.Join(dst, "file-Process2.js")); err != nil {
// 		t.Errorf("err: %s", err.Error())
// 	}

// 	substitutedJSONFile := path.Join(dst, "file-TheTest.json")
// 	if _, err := os.Stat(substitutedJSONFile); err != nil {
// 		t.Errorf("err: %s", err.Error())
// 	}

// 	jsonTestFile, err := os.Open(substitutedJSONFile)
// 	if err != nil {
// 		t.Errorf("err: %s", err.Error())
// 	}

// 	var jsonTestContent struct {
// 		Content string `json:"content"`
// 	}

// 	if err := json.NewDecoder(jsonTestFile).Decode(&jsonTestContent); err != nil {
// 		t.Errorf("err: %s", err.Error())
// 	}

// 	if jsonTestContent.Content != "TheTest" {
// 		t.Errorf("err: want %s got %s", "TheTest", jsonTestContent.Content)
// 	}
// }

// func TestGeneratorExec_invalidFilenameSubstitution(t *testing.T) {
// 	dst, err := ioutil.TempDir("", "generator")
// 	if err != nil {
// 		t.Fatalf("err: ioutil.TempDir(): %s", err)
// 	}
// 	defer os.RemoveAll(dst)

// 	//r := strings.NewReader("test\nTest project\nAlex\n2017\nStratumn\n2\nProcess1,Process2\n")
// 	r := strings.NewReader("Process1,Process2\nTheTest\n")

// 	gen, err := NewFromDir("testdata/invalid_filename_subst", &Options{Reader: r})
// 	if err != nil {
// 		t.Fatalf("err: NewFromDir(): %s", err)
// 	}

// 	if err := gen.Exec(dst); err == nil {
// 		t.Fatal("err: gen.Exec() must return an error")
// 	}
// }

// func TestSecret(t *testing.T) {
// 	s, err := secret(16)
// 	if err != nil {
// 		t.Fatalf("err: secret(): %s", err)
// 	}
// 	if got, want := len(s), 16; got != want {
// 		t.Errorf("err: len(s) = %d want %d", got, want)
// 	}
// OUTER_LOOP:
// 	for _, c := range s {
// 		for _, r := range letters {
// 			if c == r {
// 				continue OUTER_LOOP
// 			}
// 		}
// 		t.Errorf("err: unexpected rune '%c'", c)
// 	}
// }

// func TestSecret_invalidSize(t *testing.T) {
// 	if _, err := secret(-1); err == nil {
// 		t.Error("err: err = nil want Error")
// 	}
// }
