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
package repo

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stratumn/go/generator"
)

func TestUpdate(t *testing.T) {
	dir, err := ioutil.TempDir("", "generator")
	if err != nil {
		t.Fatalf("err: ioutil.TempDir(): %s", err)
	}
	defer os.RemoveAll(dir)

	r := New(dir, "stratumn", "generators")
	desc, updated, err := r.Update("master")
	if err != nil {
		t.Fatalf("err: r.Update(): %s", err)
	}

	if got, want := updated, true; got != want {
		t.Errorf("err: r.Update(): updated = %v want %v", got, want)
	}

	if got, want := desc.Owner, "stratumn"; got != want {
		t.Errorf("err: r.Update(): owner = %q want %q", got, want)
	}

	desc, updated, err = r.Update("master")
	if err != nil {
		t.Fatalf("err: r.Update(): %s", err)
	}

	if got, want := updated, false; got != want {
		t.Errorf("err: r.Update(): updated = %v want %v", got, want)
	}

	if got, want := desc.Owner, "stratumn"; got != want {
		t.Errorf("err: r.Update(): owner = %q want %q", got, want)
	}
}

func TestUpdate_notFound(t *testing.T) {
	dir, err := ioutil.TempDir("", "generator")
	if err != nil {
		t.Fatalf("err: ioutil.TempDir(): %s", err)
	}
	defer os.RemoveAll(dir)

	r := New(dir, "stratumn", "404")
	_, _, err = r.Update("master")
	if err == nil {
		t.Error("err: r.Update(): err = nil want Error")
	}
}

func TestGetState(t *testing.T) {
	dir, err := ioutil.TempDir("", "generator")
	if err != nil {
		t.Fatalf("err: ioutil.TempDir(): %s", err)
	}
	defer os.RemoveAll(dir)

	r := New(dir, "stratumn", "generators")

	desc, err := r.GetState("master")
	if err != nil {
		t.Fatalf("err: r.GetState(): %s", err)
	}
	if desc != nil {
		t.Fatalf("err: r.GetState(): desc = %#v want nil", desc)
	}

	_, _, err = r.Update("master")
	if err != nil {
		t.Fatalf("err: r.Update(): %s", err)
	}

	desc, err = r.GetState("master")
	if err != nil {
		t.Fatalf("err: r.GetState(): %s", err)
	}

	if got, want := desc.Owner, "stratumn"; got != want {
		t.Errorf("err: r.GetState(): owner = %q want %q", got, want)
	}
}

func TestGetStateOrCreate(t *testing.T) {
	dir, err := ioutil.TempDir("", "generator")
	if err != nil {
		t.Fatalf("err: ioutil.TempDir(): %s", err)
	}
	defer os.RemoveAll(dir)

	r := New(dir, "stratumn", "generators")

	desc, err := r.GetStateOrCreate("master")
	if err != nil {
		t.Fatalf("err: r.GetStateOrCreate(): %s", err)
	}

	if got, want := desc.Owner, "stratumn"; got != want {
		t.Errorf("err: r.GetStateOrCreate(): owner = %q want %q", got, want)
	}
}

func TestList(t *testing.T) {
	dir, err := ioutil.TempDir("", "generator")
	if err != nil {
		t.Fatalf("err: ioutil.TempDir(): %s", err)
	}
	defer os.RemoveAll(dir)

	r := New(dir, "stratumn", "generators")

	list, err := r.List("master")
	if err != nil {
		t.Fatalf("err: r.List(): %s", err)
	}

	if got := len(list); got < 1 {
		t.Errorf("err: len() %d want > 0", got)
	}
}

func TestGenerate(t *testing.T) {
	dir, err := ioutil.TempDir("", "generator")
	if err != nil {
		t.Fatalf("err: ioutil.TempDir(): %s", err)
	}
	defer os.RemoveAll(dir)

	dst, err := ioutil.TempDir("", "generator")
	if err != nil {
		t.Fatalf("err: ioutil.TempDir(): %s", err)
	}
	defer os.RemoveAll(dst)

	r := New(dir, "stratumn", "generators")
	opts := generator.Options{
		Reader: strings.NewReader("test\n\nStephan\n\nStratumn\n\n\n"),
	}

	err = r.Generate("agent-basic-js", dst, &opts, "master")
	if err != nil {
		t.Fatalf("err: r.Generate(): %s", err)
	}
}

func TestGenerate_notFound(t *testing.T) {
	dir, err := ioutil.TempDir("", "generator")
	if err != nil {
		t.Fatalf("err: ioutil.TempDir(): %s", err)
	}
	defer os.RemoveAll(dir)

	dst, err := ioutil.TempDir("", "generator")
	if err != nil {
		t.Fatalf("err: ioutil.TempDir(): %s", err)
	}
	defer os.RemoveAll(dst)

	r := New(dir, "stratumn", "generators")
	opts := generator.Options{
		Reader: strings.NewReader("test\n\nStephan\n\nStratumn\n\n\n"),
	}

	err = r.Generate("404", dst, &opts, "master")
	if err == nil {
		t.Error("err: r.Generate(): err = nil want Error")
	}
}
