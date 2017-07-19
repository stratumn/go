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

package fossilizerhttp

import (
	"io/ioutil"
	"net"
	"net/http"
	"testing"

	"github.com/stratumn/sdk/fossilizer/fossilizertesting"
	"github.com/stratumn/sdk/jsonhttp"
)

func createServer() (*Server, *fossilizertesting.MockAdapter) {
	a := &fossilizertesting.MockAdapter{}
	s := New(a, &Config{
		MinDataLen: 2,
		MaxDataLen: 16,
	}, &jsonhttp.Config{})

	return s, a
}

type resultHandler struct {
	t        *testing.T
	listener net.Listener
	want     string
	done     chan struct{}
}

func (h *resultHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer h.listener.Close()

	w.Write([]byte("thanks"))
	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		h.t.Fatalf("ioutil.ReadAll(): err: %s", err)
		return
	}

	if got, want := string(body), h.want; got != want {
		h.t.Errorf("h.ServerHTTP(): body = %q want %q", got, want)
	}

	h.done <- struct{}{}
}
