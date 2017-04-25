// Copyright 2017 Stratumn SAS. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

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
