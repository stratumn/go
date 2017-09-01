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
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"strings"
	"testing"
)

func TestNewGeneratorFromFile_checkVariables(t *testing.T) {
	vars := map[string]interface{}{
		"os": runtime.GOOS,
	}
	gen, err := NewDefinitionFromFile("testdata/nodejs/generator.json", vars, nil)
	if err != nil {
		t.Fatalf("err: NewDefinitionFromFile(): %s", err)
	}
	got, ok := gen.Variables["os"]
	if !ok {
		t.Fatalf(`err: gen.Variables["os"]: ok = false want true`)
	}
	if want := runtime.GOOS; got != want {
		t.Errorf(`err: gen.Variables["os"] = %q want %q`, got, want)
	}
}

func TestNewGeneratorFromFile_checkStringInput(t *testing.T) {
	vars := map[string]interface{}{
		"os": runtime.GOOS,
	}
	gen, err := NewDefinitionFromFile("testdata/nodejs/generator.json", vars, nil)
	if err != nil {
		t.Fatalf("err: NewDefinitionFromFile(): %s", err)
	}
	got, ok := gen.Inputs["name"]
	if !ok {
		t.Errorf(`err: gen.Inputs["name"]: ok = false want true`)
	} else if input, ok := got.(*StringInput); !ok {
		t.Errorf(`err: gen.Inputs["name"] should be an StringInput but decoded %#v`, got)
	} else {
		if got, want := input.Prompt, "Project name:"; got != want {
			t.Errorf(`err: input.Prompt = %q want %q`, got, want)
		}
		if got, want := input.Format, ".+"; got != want {
			t.Errorf(`err: input.Format = %q want %q`, got, want)
		}
	}
}

func TestNewGeneratorFromFile_checkSelectInput(t *testing.T) {
	vars := map[string]interface{}{
		"os": runtime.GOOS,
	}
	gen, err := NewDefinitionFromFile("testdata/nodejs/generator.json", vars, nil)
	if err != nil {
		t.Fatalf("err: NewDefinitionFromFile(): %s", err)
	}
	got, ok := gen.Inputs["license"]
	if !ok {
		t.Errorf(`err: gen.Inputs["license"]: ok = false want true`)
	} else if input, ok := got.(*StringSelect); !ok {
		t.Errorf(`err: gen.Inputs["license"] should be an StringSelect but decoded %#v`, got)
	} else {
		if got, want := input.Prompt, "License:"; got != want {
			t.Errorf(`err: input.Prompt = %q want %q`, got, want)
		}
		if got, want := len(input.Options), 2; got != want {
			t.Errorf(`err: len(input.Options) = %q want %q`, got, want)
		}
		if got, want := input.Default, "mit"; got != want {
			t.Errorf(`err: input.Prompt = %q want %q`, got, want)
		}
	}
}

func TestNewGeneratorFromFile_checkSliceInput(t *testing.T) {
	vars := map[string]interface{}{
		"os": runtime.GOOS,
	}
	gen, err := NewDefinitionFromFile("testdata/nodejs/generator.json", vars, nil)
	if err != nil {
		t.Fatalf("err: NewDefinitionFromFile(): %s", err)
	}
	got, ok := gen.Inputs["process"]
	if !ok {
		t.Errorf(`err: gen.Inputs["process"]: ok = false want true`)
	} else if input, ok := got.(*StringSlice); !ok {
		t.Errorf(`err: gen.Inputs["process"] should be an StringSlice but decoded %#v`, got)
	} else {
		if got, want := input.Prompt, "List of process names:"; got != want {
			t.Errorf(`err: input.Prompt = %q want %q`, got, want)
		}
		if got, want := input.Format, "^[a-zA-Z].*$"; got != want {
			t.Errorf(`err: input.Format = %q want %q`, got, want)
		}
		if got, want := input.Separator, ","; got != want {
			t.Errorf(`err: input.Format = %q want %q`, got, want)
		}
	}
}

func TestNewFromDir(t *testing.T) {
	gen, err := NewFromDir("testdata/nodejs", &Options{})
	if err != nil {
		t.Fatalf("err: NewFromDir(): %s", err)
	}
	if gen == nil {
		t.Fatal("err: gen = nil want *Generator")
	}
}

func TestNewFromDir_notExist(t *testing.T) {
	_, err := NewFromDir("testdata/404", &Options{})
	if err == nil {
		t.Error("err: err = nil want Error")
	}
}

func TestNewFromDir_invalidDef(t *testing.T) {
	_, err := NewFromDir("testdata/invalid_def", &Options{})
	if err == nil {
		t.Error("err: err = nil want Error")
	}
}

func TestNewFromDir_invalidDefExec(t *testing.T) {
	_, err := NewFromDir("testdata/invalid_def_exec", &Options{})
	if err == nil {
		t.Error("err: err = nil want Error")
	}
}

func TestNewFromDir_invalidDefTpml(t *testing.T) {
	_, err := NewFromDir("testdata/custom_funcs", &Options{})
	if err == nil {
		t.Error("err: err = nil want Error")
	}
}

func TestGeneratorExec(t *testing.T) {
	dst, err := ioutil.TempDir("", "generator")
	if err != nil {
		t.Fatalf("err: ioutil.TempDir(): %s", err)
	}
	defer os.RemoveAll(dst)

	r := strings.NewReader("test\nTest project\nStephan Florquin\n2016\nStratumn\n2\nProcess1,Process2\n")

	gen, err := NewFromDir("testdata/nodejs", &Options{Reader: r})
	if err != nil {
		t.Fatalf("err: NewFromDir(): %s", err)
	}

	if err := gen.Exec(dst); err != nil {
		t.Fatalf("err: gen.Exec(): %s", err)
	}

	cmpWalk(t, "testdata/nodejs_expected", dst, "testdata/nodejs_expected")
}

func TestGeneratorExec_ask(t *testing.T) {
	dst, err := ioutil.TempDir("", "generator")
	if err != nil {
		t.Fatalf("err: ioutil.TempDir(): %s", err)
	}
	defer os.RemoveAll(dst)

	r := strings.NewReader("\n\nTest Project\n")

	gen, err := NewFromDir("testdata/ask", &Options{Reader: r})
	if err != nil {
		t.Fatalf("err: NewFromDir(): %s", err)
	}

	if err := gen.Exec(dst); err != nil {
		t.Fatalf("err: gen.Exec(): %s", err)
	}

	cmpWalk(t, "testdata/ask_expected", dst, "testdata/ask_expected")
}

func TestGeneratorExec_tmplVars(t *testing.T) {
	dst, err := ioutil.TempDir("", "generator")
	if err != nil {
		t.Fatalf("err: ioutil.TempDir(): %s", err)
	}
	defer os.RemoveAll(dst)

	opts := Options{
		TmplVars: map[string]interface{}{"test": "hello"},
	}

	gen, err := NewFromDir("testdata/vars", &opts)
	if err != nil {
		t.Fatalf("err: NewFromDir(): %s", err)
	}

	if err := gen.Exec(dst); err != nil {
		t.Fatalf("err: gen.Exec(): %s", err)
	}

	cmpWalk(t, "testdata/vars_expected", dst, "testdata/vars_expected")
}

func TestGeneratorExec_customFuncs(t *testing.T) {
	dst, err := ioutil.TempDir("", "generator")
	if err != nil {
		t.Fatalf("err: ioutil.TempDir(): %s", err)
	}
	defer os.RemoveAll(dst)

	opts := Options{
		DefFuncs: map[string]interface{}{
			"custom": func() string { return "hello generator" },
		},
		TmplFuncs: map[string]interface{}{
			"custom": func() string { return "hello template" },
		},
	}

	gen, err := NewFromDir("testdata/custom_funcs", &opts)
	if err != nil {
		t.Fatalf("err: NewFromDir(): %s", err)
	}

	if err := gen.Exec(dst); err != nil {
		t.Fatalf("err: gen.Exec(): %s", err)
	}

	cmpWalk(t, "testdata/custom_funcs_expected", dst, "testdata/custom_funcs_expected")
}

func TestGeneratorExec_inputError(t *testing.T) {
	dst, err := ioutil.TempDir("", "generator")
	if err != nil {
		t.Fatalf("err: ioutil.TempDir(): %s", err)
	}
	defer os.RemoveAll(dst)

	gen, err := NewFromDir("testdata/nodejs", &Options{})
	if err != nil {
		t.Fatalf("err: NewFromDir(): %s", err)
	}

	if err := gen.Exec(dst); err == nil {
		t.Error("err: err = nil want Error")
	}
}

func TestGeneratorExec_askError(t *testing.T) {
	dst, err := ioutil.TempDir("", "generator")
	if err != nil {
		t.Fatalf("err: ioutil.TempDir(): %s", err)
	}
	defer os.RemoveAll(dst)

	gen, err := NewFromDir("testdata/ask", &Options{})
	if err != nil {
		t.Fatalf("err: NewFromDir(): %s", err)
	}

	if err := gen.Exec(dst); err == nil {
		t.Error("err: err = nil want Error")
	}
}

func TestGeneratorExec_askInvalid(t *testing.T) {
	dst, err := ioutil.TempDir("", "generator")
	if err != nil {
		t.Fatalf("err: ioutil.TempDir(): %s", err)
	}
	defer os.RemoveAll(dst)

	gen, err := NewFromDir("testdata/ask_invalid", &Options{})
	if err != nil {
		t.Fatalf("err: NewFromDir(): %s", err)
	}

	if err := gen.Exec(dst); err == nil {
		t.Error("err: err = nil want Error")
	}
}

func TestGeneratorExec_invalidTpml(t *testing.T) {
	dst, err := ioutil.TempDir("", "generator")
	if err != nil {
		t.Fatalf("err: ioutil.TempDir(): %s", err)
	}
	defer os.RemoveAll(dst)

	gen, err := NewFromDir("testdata/invalid_tmpl", &Options{})
	if err != nil {
		t.Fatalf("err: NewFromDir(): %s", err)
	}

	if err := gen.Exec(dst); err == nil {
		t.Error("err: err = nil want Error")
	}
}

func TestGeneratorExec_invalidPartial(t *testing.T) {
	dst, err := ioutil.TempDir("", "generator")
	if err != nil {
		t.Fatalf("err: ioutil.TempDir(): %s", err)
	}
	defer os.RemoveAll(dst)

	r := strings.NewReader("test\nTest project\nStephan Florquin\n2016\nStratumn\n2\nProcess1,Process2\n")

	gen, err := NewFromDir("testdata/invalid_partial", &Options{Reader: r})
	if err != nil {
		t.Fatalf("err: NewFromDir(): %s", err)
	}

	if err := gen.Exec(dst); err == nil {
		t.Error("err: err = nil want Error")
	}
}

func TestGeneratorExec_invalidPartialExec(t *testing.T) {
	dst, err := ioutil.TempDir("", "generator")
	if err != nil {
		t.Fatalf("err: ioutil.TempDir(): %s", err)
	}
	defer os.RemoveAll(dst)

	r := strings.NewReader("test\nTest project\nStephan Florquin\n2016\nStratumn\n2\nProcess1,Process2\n")

	gen, err := NewFromDir("testdata/invalid_partial_exec", &Options{Reader: r})
	if err != nil {
		t.Fatalf("err: NewFromDir(): %s", err)
	}

	if err := gen.Exec(dst); err == nil {
		t.Error("err: err = nil want Error")
	}
}

func TestGeneratorExec_undefinedInput(t *testing.T) {
	dst, err := ioutil.TempDir("", "generator")
	if err != nil {
		t.Fatalf("err: ioutil.TempDir(): %s", err)
	}
	defer os.RemoveAll(dst)

	gen, err := NewFromDir("testdata/undefined_input", &Options{})
	if err != nil {
		t.Fatalf("err: NewFromDir(): %s", err)
	}

	if err := gen.Exec(dst); err == nil {
		t.Error("err: err = nil want Error")
	}
}

func TestGeneratorExec_invalidPartialArgs(t *testing.T) {
	dst, err := ioutil.TempDir("", "generator")
	if err != nil {
		t.Fatalf("err: ioutil.TempDir(): %s", err)
	}
	defer os.RemoveAll(dst)

	gen, err := NewFromDir("testdata/invalid_partial_args", &Options{})
	if err != nil {
		t.Fatalf("err: NewFromDir(): %s", err)
	}

	if err := gen.Exec(dst); err == nil {
		t.Error("err: err = nil want Error")
	}
}

func TestGeneratorExec_filenameSubstitution(t *testing.T) {
	dst, err := ioutil.TempDir("", "generator")
	if err != nil {
		t.Fatalf("err: ioutil.TempDir(): %s", err)
	}
	defer os.RemoveAll(dst)

	//r := strings.NewReader("test\nTest project\nAlex\n2017\nStratumn\n2\nProcess1,Process2\n")
	r := strings.NewReader("Process1,Process2\nTheTest\n")

	gen, err := NewFromDir("testdata/filename_subst", &Options{Reader: r})
	if err != nil {
		t.Fatalf("err: NewFromDir(): %s", err)
	}

	if err := gen.Exec(dst); err != nil {
		t.Fatalf("err: gen.Exec(): %s", err)
	}

	if _, err := os.Stat(path.Join(dst, "file-Process1.js")); err != nil {
		t.Errorf("err: err %s want Error", err.Error())
	}

	if _, err := os.Stat(path.Join(dst, "file-Process2.js")); err != nil {
		t.Errorf("err: err %s want Error", err.Error())
	}

	if _, err := os.Stat(path.Join(dst, "file-TheTest.json")); err != nil {
		t.Errorf("err: err %s want Error", err.Error())
	}
}

func TestGeneratorExec_invalidFilenameSubstitution(t *testing.T) {
	dst, err := ioutil.TempDir("", "generator")
	if err != nil {
		t.Fatalf("err: ioutil.TempDir(): %s", err)
	}
	defer os.RemoveAll(dst)

	//r := strings.NewReader("test\nTest project\nAlex\n2017\nStratumn\n2\nProcess1,Process2\n")
	r := strings.NewReader("Process1,Process2\nTheTest\n")

	gen, err := NewFromDir("testdata/invalid_filename_subst", &Options{Reader: r})
	if err != nil {
		t.Fatalf("err: NewFromDir(): %s", err)
	}

	if err := gen.Exec(dst); err == nil {
		t.Fatal("err: gen.Exec() must return an error")
	}
}

func TestSecret(t *testing.T) {
	s, err := secret(16)
	if err != nil {
		t.Fatalf("err: secret(): %s", err)
	}
	if got, want := len(s), 16; got != want {
		t.Errorf("err: len(s) = %d want %d", got, want)
	}
OUTER_LOOP:
	for _, c := range s {
		for _, r := range letters {
			if c == r {
				continue OUTER_LOOP
			}
		}
		t.Errorf("err: unexpected rune '%c'", c)
	}
}

func TestSecret_invalidSize(t *testing.T) {
	if _, err := secret(-1); err == nil {
		t.Error("err: err = nil want Error")
	}
}
