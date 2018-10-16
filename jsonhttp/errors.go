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
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/stratumn/go-core/monitoring/errorcode"
	"github.com/stratumn/go-core/types"
)

var (
	// ErrorCodeToHTTPCode maps internal error codes to http status code.
	ErrorCodeToHTTPCode = map[int]int{
		errorcode.Ok:                 http.StatusOK,
		errorcode.InvalidArgument:    http.StatusBadRequest,
		errorcode.FailedPrecondition: http.StatusBadRequest,
		errorcode.OutOfRange:         http.StatusBadRequest,
		errorcode.AlreadyExists:      http.StatusConflict,
		errorcode.Aborted:            http.StatusConflict,
		errorcode.PermissionDenied:   http.StatusForbidden,
		errorcode.DeadlineExceeded:   http.StatusGatewayTimeout,
		errorcode.Unknown:            http.StatusInternalServerError,
		errorcode.Internal:           http.StatusInternalServerError,
		errorcode.DataLoss:           http.StatusInternalServerError,
		errorcode.NotFound:           http.StatusNotFound,
		errorcode.Unimplemented:      http.StatusNotImplemented,
		errorcode.Unavailable:        http.StatusServiceUnavailable,
		errorcode.ResourceExhausted:  http.StatusTooManyRequests,
		errorcode.Unauthenticated:    http.StatusUnauthorized,
	}
)

// ErrHTTP is an error with an HTTP status code.
type ErrHTTP struct {
	msg    string
	status int
}

// NewErrHTTP creates an http error from an internal error.
func NewErrHTTP(err error) ErrHTTP {
	switch e := err.(type) {
	case *types.Error:
		status, ok := ErrorCodeToHTTPCode[e.Code]
		if ok {
			return ErrHTTP{
				msg:    e.Error(),
				status: status,
			}
		}
	}

	return ErrHTTP{
		msg:    err.Error(),
		status: http.StatusInternalServerError,
	}
}

// NewErrNotFound creates an error not found.
func NewErrNotFound() ErrHTTP {
	return ErrHTTP{
		msg:    http.StatusText(http.StatusNotFound),
		status: http.StatusNotFound,
	}
}

// NewErrInternalServer creates an internal server error.
func NewErrInternalServer() ErrHTTP {
	return ErrHTTP{
		msg:    http.StatusText(http.StatusInternalServerError),
		status: http.StatusInternalServerError,
	}
}

// Status returns the HTTP status code of the error.
func (e ErrHTTP) Status() int {
	return e.status
}

// Error implements error.Error.
func (e ErrHTTP) Error() string {
	return e.msg
}

var internalServerJSON = fmt.Sprintf(`{"error:":"internal server error","status":%d}`, http.StatusInternalServerError)

// JSONMarshal marshals an error to JSON.
func (e ErrHTTP) JSONMarshal() []byte {
	js, err := json.Marshal(map[string]interface{}{
		"error":  e.msg,
		"status": e.status,
	})
	if err != nil {
		msg := internalServerJSON
		return []byte(msg)
	}

	return js
}
