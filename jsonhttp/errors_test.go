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

package jsonhttp

import (
	"net/http"
	"testing"
)

func TestNewErrInternalServer(t *testing.T) {
	testErrStatus(t, NewErrInternalServer(""), http.StatusInternalServerError)
	testErrError(t, NewErrInternalServer(""), "internal server error")
	testErrError(t, NewErrInternalServer("test"), "test")
}

func TestNewErrBadRequest(t *testing.T) {
	testErrStatus(t, NewErrBadRequest(""), http.StatusBadRequest)
	testErrError(t, NewErrBadRequest(""), "bad request")
	testErrError(t, NewErrBadRequest("test"), "test")
}

func TestNewErrUnauthorized(t *testing.T) {
	testErrStatus(t, NewErrUnauthorized(""), http.StatusUnauthorized)
	testErrError(t, NewErrUnauthorized(""), "unauthorized")
	testErrError(t, NewErrUnauthorized("test"), "test")
}

func TestNewErrNotFound(t *testing.T) {
	testErrStatus(t, NewErrNotFound(""), http.StatusNotFound)
	testErrError(t, NewErrNotFound(""), "not found")
	testErrError(t, NewErrNotFound("test"), "test")
}
