// Copyright 2016 Stratumn SAS. All rights reserved.
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
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"testing"
)

func TestNewGeneratorFromFile(t *testing.T) {
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

func TestDefaultGeneratorFuncs(t *testing.T) {
	if got := DefaultDefinitionFuncs(); got == nil {
		t.Errorf("err: DefaultDefinitionFuncs() = %q want FuncMap", got)
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

func TestGeneratorExec(t *testing.T) {
	dst, err := ioutil.TempDir("", "generator")
	if err != nil {
		t.Fatalf("err: ioutil.TempDir(): %s", err)
	}
	fmt.Println(dst)
	defer os.RemoveAll(dst)

	r := strings.NewReader("test\nTest project\nStephan Florquin\n2016\nStratumn\n2\n")

	gen, err := NewFromDir("testdata/nodejs", &Options{Reader: r})
	if err != nil {
		t.Fatalf("err: NewFromDir(): %s", err)
	}

	if err := gen.Exec(dst); err != nil {
		t.Fatalf("err: gen.Exec(): %s", err)
	}

	cmpWalk(t, "testdata/nodejs_expected", dst, "testdata/nodejs_expected")
}
