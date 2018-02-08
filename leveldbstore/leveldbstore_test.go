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

package leveldbstore

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/store/storetestcases"
)

func TestLevelDBStore(t *testing.T) {
	factory := storetestcases.Factory{
		NewKeyValueStore: func() (store.KeyValueStore, error) {
			path, err := ioutil.TempDir("", "leveldbstore")
			if err != nil {
				return nil, err
			}
			db, err := New(&Config{Path: path})
			if err != nil {
				return nil, err
			}
			return db, nil
		},
		FreeKeyValueStore: func(s store.KeyValueStore) {
			a := s.(*LevelDBStore)
			defer os.RemoveAll(a.config.Path)
		},
	}

	factory.RunKeyValueStoreTests(t)
}
