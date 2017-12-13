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

package tmpop

import (
	"fmt"

	abci "github.com/tendermint/abci/types"
)

const (
	// CodeTypeValidation is the ABCI error code for a validation error.
	CodeTypeValidation uint32 = 400

	// CodeTypeInternalError is the ABCI error code for an internal error.
	CodeTypeInternalError uint32 = 500

	// CodeTypeNotImplemented is the ABCI error code for a feature not yet implemented.
	CodeTypeNotImplemented uint32 = 501
)

// ABCIError is a structured error used inside TMPoP.
// The error codes are close to standard http status codes.
type ABCIError struct {
	Code uint32
	Log  string
}

// IsOK returns true if no error occurred.
func (e *ABCIError) IsOK() bool {
	return e == nil || e.Code == abci.CodeTypeOK
}

func (e *ABCIError) Error() string {
	if e.IsOK() {
		return ""
	}

	return fmt.Sprintf("Code: %d. Log: %s", e.Code, e.Log)
}
