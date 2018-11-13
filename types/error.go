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

package types

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/pkg/errors"
)

const (
	// Component name for monitoring.
	Component = "types"
)

// Error structure to provide accurate metrics and logs.
// Errors coming from external dependencies should always be wrapped inside an
// instance of this structured error type (in particular to capture a stack
// trace).
// Errors coming from dependencies inside go-core should either be propagated
// as-is or wrapped depending on the code's ability to add new relevant
// context.
type Error struct {
	Code      int    // Error code (see monitoring/errorcode.go).
	Component string // Component that triggered the error (e.g. "fossilizer").
	Message   string // Custom error message.
	Wrapped   error  // Inner error.
}

// NewError creates a new error and captures a stack trace.
func NewError(code int, component string, message string) *Error {
	return &Error{
		Code:      code,
		Component: component,
		Message:   message,
		Wrapped:   errors.New(""), // this empty error captures a stack trace.
	}
}

// NewErrorf creates a new error and captures a stack trace.
func NewErrorf(code int, component string, format string, args ...interface{}) *Error {
	return &Error{
		Code:      code,
		Component: component,
		Message:   fmt.Sprintf(format, args...),
		Wrapped:   errors.New(""), // this empty error captures a stack trace.
	}
}

// WrapError wraps an error and adds a stack trace.
func WrapError(err error, code int, component string, message string) *Error {
	e := &Error{
		Code:      code,
		Component: component,
		Message:   message,
	}

	switch err.(type) {
	case *Error:
		e.Wrapped = err
	default:
		e.Wrapped = errors.WithStack(err)
	}

	return e
}

// WrapErrorf wraps an error and adds a stack trace.
func WrapErrorf(err error, code int, component string, format string, args ...interface{}) *Error {
	e := &Error{
		Code:      code,
		Component: component,
		Message:   fmt.Sprintf(format, args...),
	}

	switch err.(type) {
	case *Error:
		e.Wrapped = err
	default:
		e.Wrapped = errors.WithStack(err)
	}

	return e
}

// Error returns a friendly error message.
func (e *Error) Error() string {
	if e.Wrapped != nil && e.Wrapped.Error() != "" {
		return fmt.Sprintf("%s error %d: %s: %s", e.Component, e.Code, e.Message, e.Wrapped.Error())
	}

	return fmt.Sprintf("%s error %d: %s", e.Component, e.Code, e.Message)
}

// String returns a friendly error message.
func (e *Error) String() string {
	return e.Error()
}

// Format the error when printing.
// %+v prints the full stack trace of the deepest error in the stack in case an
// internal error was wrapped.
func (e *Error) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			_, _ = io.WriteString(s, fmt.Sprintf("%s error %d: %s", e.Component, e.Code, e.Message))
			if e.Wrapped != nil {
				_, _ = io.WriteString(s, fmt.Sprintf("\n%+v", e.Wrapped))
			}
			return
		}
		fallthrough
	case 's':
		_, _ = io.WriteString(s, e.Error())
	}
}

// MarshalJSON recursively checks inner errors to marshal them properly.
func (e *Error) MarshalJSON() ([]byte, error) {
	marshalled := make(map[string]interface{})
	marshalled["code"] = e.Code
	marshalled["category"] = e.Component
	marshalled["message"] = e.Message

	if e.Wrapped != nil {
		switch wrapped := e.Wrapped.(type) {
		case *Error:
			marshalled["inner"] = wrapped
		default:
			marshalled["inner"] = wrapped.Error()
		}
	}

	return json.Marshal(marshalled)
}
