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
package repo

import (
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/stratumn/go-indigocore/generator"
)

var (
	testUser  = "stratumn"
	testRepo  = "generators"
	testRef   = "sdk-test"
	testOwner = "stratumn"
	testInput = "test\n\nStephan\n\nStratumn\n\n\nstratumn\npurchase,shipment\n\n"
)

func TestUpdate(t *testing.T) {
	dir, err := ioutil.TempDir("", "generator")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	r := New(dir, testUser, testRepo, os.Getenv("GITHUB_TOKEN"), true)

	desc, updated, err := r.Update(testRef, false)
	require.NoError(t, err)
	assert.True(t, updated)
	assert.Equal(t, testOwner, desc.Owner)

	desc, updated, err = r.Update(testRef, false)
	require.NoError(t, err)
	assert.False(t, updated)
	assert.Equal(t, testOwner, desc.Owner)

	desc, updated, err = r.Update(testRef, true)
	require.NoError(t, err)
	assert.True(t, updated)
	assert.Equal(t, testOwner, desc.Owner)
}

func TestUpdate_notFound(t *testing.T) {
	dir, err := ioutil.TempDir("", "generator")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	r := New(dir, testUser, "404", os.Getenv("GITHUB_TOKEN"), true)
	_, _, err = r.Update(testRef, false)
	assert.Error(t, err)
}

func TestGetState(t *testing.T) {
	dir, err := ioutil.TempDir("", "generator")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	r := New(dir, testUser, testRepo, os.Getenv("GITHUB_TOKEN"), true)

	desc, err := r.GetState(testRef)
	require.NoError(t, err)
	require.Nil(t, desc)

	_, _, err = r.Update(testRef, false)
	assert.NoError(t, err)

	desc, err = r.GetState(testRef)
	require.NoError(t, err)
	assert.Equal(t, testOwner, desc.Owner)
}

func TestGetStateOrCreate(t *testing.T) {
	dir, err := ioutil.TempDir("", "generator")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	r := New(dir, testUser, testRepo, os.Getenv("GITHUB_TOKEN"), true)

	desc, err := r.GetStateOrCreate(testRef)
	require.NoError(t, err)
	assert.Equal(t, testOwner, desc.Owner)
}

func TestList(t *testing.T) {
	dir, err := ioutil.TempDir("", "generator")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	r := New(dir, testUser, testRepo, os.Getenv("GITHUB_TOKEN"), true)

	list, err := r.List(testRef)
	require.NoError(t, err)
	assert.NotEmpty(t, list)
}

func TestLocalList(t *testing.T) {
	// Get generators from git, it should be better
	dir, err := ioutil.TempDir("", "generator")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	r := New(dir, testUser, testRepo, os.Getenv("GITHUB_TOKEN"), true)

	_, err = r.GetStateOrCreate(testRef)
	require.NoError(t, err)

	r = New(path.Join(dir, "src", testRef), "foo", "bar", "nil", false)

	list, err := r.List("unread arg")
	require.NoError(t, err)
	assert.NotEmpty(t, list)
}

func TestNotFoundLocalList(t *testing.T) {
	r := New("/foo/bar", "foo", "bar", "nil", false)

	list, err := r.List("unread arg")
	assert.Error(t, err)
	assert.Empty(t, list)
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

	r := New(dir, testUser, testRepo, os.Getenv("GITHUB_TOKEN"), true)
	opts := generator.Options{
		Reader: strings.NewReader(testInput),
	}

	err = r.Generate("404", dst, &opts, testRef)
	if err == nil {
		t.Error("err: r.Generate(): err = nil want Error")
	}
}
