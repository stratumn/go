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

package filestore

import (
	"io/ioutil"
	"testing"
)

func createAdapter(tb testing.TB) *FileStore {
	path, err := ioutil.TempDir("", "filestore")
	if err != nil {
		tb.Fatalf("ioutil.TempDir(): err: %s", err)
	}
	fs, err := New(&Config{Path: path})
	if err != nil {
		tb.Fatalf("New(): err: %s", err)
	}
	return fs
}
