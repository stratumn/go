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

package storehttp

import (
	"time"

	"github.com/stratumn/sdk/jsonhttp"
	"github.com/stratumn/sdk/jsonws"
	"github.com/stratumn/sdk/store/storetesting"
)

func createServer() (*Server, *storetesting.MockAdapter) {
	a := &storetesting.MockAdapter{}
	s := New(a, &Config{}, &jsonhttp.Config{}, &jsonws.BasicConfig{}, &jsonws.BufferedConnConfig{
		Size:         256,
		WriteTimeout: 10 * time.Second,
		PongTimeout:  70 * time.Second,
		PingInterval: time.Minute,
		MaxMsgSize:   1024,
	})

	return s, a
}
