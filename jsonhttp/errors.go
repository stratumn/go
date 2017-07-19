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
	"encoding/json"
	"fmt"
	"net/http"
)

// ErrHTTP is an error with an HTTP status code.
type ErrHTTP struct {
	msg    string
	status int
}

// NewErrHTTP creates a, error with a message and HTTP status code.
func NewErrHTTP(msg string, status int) ErrHTTP {
	return ErrHTTP{msg, status}
}

// NewErrInternalServer creates an error with an internal server error HTTP
// status code.
// If the message is empty, the default is "internal server error".
func NewErrInternalServer(msg string) ErrHTTP {
	if msg == "" {
		msg = "internal server error"
	}
	return NewErrHTTP(msg, http.StatusInternalServerError)
}

// NewErrBadRequest creates an error with a bad request HTTP status code.
// If the message is empty, the default is "bad request".
func NewErrBadRequest(msg string) ErrHTTP {
	if msg == "" {
		msg = "bad request"
	}
	return NewErrHTTP(msg, http.StatusBadRequest)
}

// NewErrUnauthorized creates an error with an unauthorized HTTP status code.
// If the message is empty, the default is "unauthorized".
func NewErrUnauthorized(msg string) ErrHTTP {
	if msg == "" {
		msg = "unauthorized"
	}
	return NewErrHTTP(msg, http.StatusUnauthorized)
}

// NewErrNotFound creates an error with a not found HTTP status code.
// If the message is empty, the default is "not found".
func NewErrNotFound(msg string) ErrHTTP {
	if msg == "" {
		msg = "not found"
	}
	return NewErrHTTP(msg, http.StatusNotFound)
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
