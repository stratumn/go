// Copyright 2016-2018 Stratumn SAS. All rights reserved.
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
	"strings"
	"testing"

	"github.com/stratumn/go-core/monitoring/errorcode"
	"github.com/stratumn/go-core/types"
	"github.com/stretchr/testify/assert"
)

func TestNewErr(t *testing.T) {
	assert.Equal(t, http.StatusBadRequest, NewErrHTTP(&types.Error{Code: errorcode.InvalidArgument}).Status())
	assert.Equal(t, http.StatusBadRequest, NewErrHTTP(&types.Error{Code: errorcode.FailedPrecondition}).Status())
	assert.Equal(t, http.StatusServiceUnavailable, NewErrHTTP(&types.Error{Code: errorcode.Unavailable}).Status())
	assert.Equal(t, http.StatusNotImplemented, NewErrHTTP(&types.Error{Code: errorcode.Unimplemented}).Status())
}

func TestNewErrNotFound(t *testing.T) {
	assert.Equal(t, http.StatusNotFound, NewErrNotFound().Status())
	assert.Equal(t, "Not Found", NewErrNotFound().Error())

	customErr := NewErrHTTP(&types.Error{Code: errorcode.NotFound, Message: "test"})
	assert.True(t, strings.Contains(customErr.Error(), "test"))
}

func TestNewErrInternalServer(t *testing.T) {
	assert.Equal(t, http.StatusInternalServerError, NewErrInternalServer().Status())
	assert.Equal(t, "Internal Server Error", NewErrInternalServer().Error())

	customErr := NewErrHTTP(&types.Error{Code: errorcode.Internal, Message: "test"})
	assert.True(t, strings.Contains(customErr.Error(), "test"))
}
