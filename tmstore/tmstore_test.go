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

package tmstore

import (
	"testing"

	"github.com/stratumn/go/store"
	"github.com/stratumn/go/store/storetestcases"
)

func TestTMStore(t *testing.T) {
	storetestcases.Factory{
		New: func() (store.Adapter, error) {
			config := &Config{
				Endpoint: GetConfig().GetString("rpc_laddr"),
			}
			s := New(config)
			go s.StartWebsocket()
			return s, nil
		},
		Free: func(s store.Adapter) {
			s.(*TMStore).StopWebsocket()
			Reset()
		},
	}.RunTests(t)
}
