// Copyright 2017 Stratumn SAS. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

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
