// Copyright 2017 Stratumn SAS. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package repo

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stratumn/sdk/generator"
)

func TestUpdate(t *testing.T) {
	dir, err := ioutil.TempDir("", "generator")
	if err != nil {
		t.Fatalf("err: ioutil.TempDir(): %s", err)
	}
	defer os.RemoveAll(dir)

	r := New(dir, "stratumn", "generators", os.Getenv("GITHUB_TOKEN"))
	desc, updated, err := r.Update("master", false)
	if err != nil {
		t.Fatalf("err: r.Update(): %s", err)
	}

	if got, want := updated, true; got != want {
		t.Errorf("err: r.Update(): updated = %v want %v", got, want)
	}

	if got, want := desc.Owner, "stratumn"; got != want {
		t.Errorf("err: r.Update(): owner = %q want %q", got, want)
	}

	desc, updated, err = r.Update("master", false)
	if err != nil {
		t.Fatalf("err: r.Update(): %s", err)
	}

	if got, want := updated, false; got != want {
		t.Errorf("err: r.Update(): updated = %v want %v", got, want)
	}

	if got, want := desc.Owner, "stratumn"; got != want {
		t.Errorf("err: r.Update(): owner = %q want %q", got, want)
	}

	desc, updated, err = r.Update("master", true)
	if err != nil {
		t.Fatalf("err: r.Update(): %s", err)
	}

	if got, want := updated, true; got != want {
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

	r := New(dir, "stratumn", "404", os.Getenv("GITHUB_TOKEN"))
	_, _, err = r.Update("master", false)
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

	r := New(dir, "stratumn", "generators", os.Getenv("GITHUB_TOKEN"))

	desc, err := r.GetState("master")
	if err != nil {
		t.Fatalf("err: r.GetState(): %s", err)
	}
	if desc != nil {
		t.Fatalf("err: r.GetState(): desc = %#v want nil", desc)
	}

	_, _, err = r.Update("master", false)
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

	r := New(dir, "stratumn", "generators", os.Getenv("GITHUB_TOKEN"))

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

	r := New(dir, "stratumn", "generators", os.Getenv("GITHUB_TOKEN"))

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

	r := New(dir, "stratumn", "generators", os.Getenv("GITHUB_TOKEN"))
	opts := generator.Options{
		Reader: strings.NewReader("test\n\nStephan\n\nStratumn\n\n\nstratumn\n\n"),
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

	r := New(dir, "stratumn", "generators", os.Getenv("GITHUB_TOKEN"))
	opts := generator.Options{
		Reader: strings.NewReader("test\n\nStephan\n\nStratumn\n\n\n\n"),
	}

	err = r.Generate("404", dst, &opts, "master")
	if err == nil {
		t.Error("err: r.Generate(): err = nil want Error")
	}
}
