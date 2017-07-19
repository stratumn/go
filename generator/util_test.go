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
	"path/filepath"
	"reflect"
	"testing"
)

func cmpWalk(t *testing.T, src, dst, dir string) {
	files, err := filepath.Glob(filepath.Join(dir, "*"))
	if err != nil {
		t.Fatalf("err: filepath.Glob(): %s", err)
	}
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			t.Fatalf("err: os.Stat(): %s", err)
		}
		if info.IsDir() {
			cmpWalk(t, src, dst, file)
			continue
		}
		rel, err := filepath.Rel(src, file)
		if err != nil {
			t.Fatalf("err: filepath.Rel(): %s", err)
		}
		srcPath := filepath.Join(src, rel)
		srcInfo, err := os.Stat(srcPath)
		if err != nil {
			t.Fatalf("err: os.Stat(): %s", err)
		}
		dstPath := filepath.Join(dst, rel)
		dstInfo, err := os.Stat(dstPath)
		if err != nil {
			t.Fatalf("err: os.Stat(): %s", err)
		}
		if got, want := srcInfo.Mode(), dstInfo.Mode(); got != want {
			t.Errorf("err: srcInfo.Mode() = %d want %d", got, want)
		}
		got, err := ioutil.ReadFile(srcPath)
		if err != nil {
			t.Fatalf("err: ioutil.ReadFile(): %s", err)
		}
		want, err := ioutil.ReadFile(dstPath)
		if err != nil {
			t.Fatalf("err: ioutil.ReadFile(): %s", err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("err: content =\n%q\nwant\n%q", got, want)
		}
	}
}
