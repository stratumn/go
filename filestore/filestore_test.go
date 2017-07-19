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
	"os"
	"testing"

	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/store/storetestcases"
)

func TestFilestore(t *testing.T) {
	storetestcases.Factory{
		New: func() (store.Adapter, error) {
			return createAdapter(t), nil
		},
		Free: func(s store.Adapter) {
			a := s.(*FileStore)
			defer os.RemoveAll(a.config.Path)
		},
	}.RunTests(t)
}
