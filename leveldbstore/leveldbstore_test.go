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

	"github.com/stratumn/sdk/store"
)

func createTestStore() (store.KeyValueStore, error) {
	path, err := ioutil.TempDir("", "leveldbstore")
	if err != nil {
		return nil, err
	}
	db, err := New(&Config{Path: path})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func freeTestStore(s store.KeyValueStore) {
	a := s.(*LevelDBStore)
	defer os.RemoveAll(a.config.Path)
}

func TestKeyValueStore(t *testing.T) {
	testStore, err := createTestStore()
	defer freeTestStore(testStore)

	if err != nil {
		t.Error(err)
	}

	t.Run("SetValue correctly updates store", func(t *testing.T) {
		key := []byte("key1")
		initialValue := "value1"
		if err := testStore.SetValue(key, []byte(initialValue)); err != nil {
			t.Error(err)
		}

		value, err := testStore.GetValue(key)
		if err != nil {
			t.Error(err)
		}
		if got := string(value[:]); got != initialValue {
			t.Errorf("Invalid value: want %s, got %s", initialValue, got)
		}

		updatedValue := "value2"
		if err = testStore.SetValue(key, []byte(updatedValue)); err != nil {
			t.Error(err)
		}

		value, err = testStore.GetValue(key)
		if err != nil {
			t.Error(err)
		}

		if got := string(value[:]); got != updatedValue {
			t.Errorf("Invalid value: want %s, got %s", updatedValue, got)
		}
	})

	t.Run("DeleteValue correctly removes from store", func(t *testing.T) {
		key := []byte("to-delete")
		value := "I will be deleted"
		if err := testStore.SetValue(key, []byte(value)); err != nil {
			t.Error(err)
		}

		deleted, err := testStore.DeleteValue(key)
		if err != nil {
			t.Error(err)
		}

		if got := string(deleted[:]); got != value {
			t.Errorf("Invalid value: want %s, got %s", got, value)
		}

		notFound, err := testStore.GetValue(key)
		if err != nil {
			t.Error(err)
		}
		if notFound != nil {
			t.Error("Value should be nil after delete")
		}
	})

	t.Run("GetValue should return nil for unknown key", func(t *testing.T) {
		notFound, err := testStore.GetValue([]byte("You won't find me"))
		if err != nil {
			t.Error(err)
		}
		if notFound != nil {
			t.Error("Value should be nil for unknown key")
		}

	})
}
