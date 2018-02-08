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

// Package testutil contains helpers for tests.
package testutil

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"net/http"
	"net/http/httptest"

	"github.com/stratumn/go-indigocore/types"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// RandomHash creates a random hash.
func RandomHash() *types.Bytes32 {
	var hash types.Bytes32
	for i := range hash {
		hash[i] = byte(letters[rand.Intn(len(letters))])
	}
	return &hash
}

// RandomString generates a random string.
func RandomString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// ContainsString checks if an array contains a string.
func ContainsString(a []string, s string) bool {
	for _, v := range a {
		if v == s {
			return true
		}
	}
	return false
}

// RequestJSON does a request expecting a JSON response.
func RequestJSON(h http.HandlerFunc, method, target string, payload, dst interface{}) (*httptest.ResponseRecorder, error) {
	var req *http.Request

	if payload != nil {
		body, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		req = httptest.NewRequest(method, target, bytes.NewReader(body))
	} else {
		req = httptest.NewRequest(method, target, nil)
	}

	w := httptest.NewRecorder()
	h(w, req)

	if dst != nil {
		if err := json.NewDecoder(w.Body).Decode(&dst); err != nil {
			return nil, err
		}
	}

	return w, nil
}

// RandomValue returns a random byte array (with max 1024 bytes)
func RandomValue() []byte {
	c := rand.Intn(1024)
	b := make([]byte, c)
	rand.Read(b)
	return b
}

// RandomKey returns a random byte array (with max 64 bytes)
func RandomKey() []byte {
	c := rand.Intn(63) + 1
	b := make([]byte, c)
	rand.Read(b)
	return b
}
