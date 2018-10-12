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

package testutil

import (
	"strings"
	"testing"

	"github.com/stratumn/go-core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// AssertWrappedErrorEqual checks that the innermost error equals the given
// expected error.
func AssertWrappedErrorEqual(t *testing.T, err error, expected error) {
	require.NotNil(t, err)
	require.IsType(t, &types.Error{}, err)

	structErr := err.(*types.Error)
	inner := error(structErr)

	for {
		structInner, ok := inner.(*types.Error)
		if !ok {
			break
		}

		if structInner.Wrapped == nil {
			break
		}

		inner = structInner.Wrapped
	}

	assert.EqualError(t, inner, expected.Error())
}

// AssertErrorContains checks that the error contains a given sub-error.
func AssertErrorContains(t *testing.T, err error, expected error) {
	i := strings.Index(err.Error(), expected.Error())
	assert.Truef(t, i > 0, "'%s' should contain '%s'", err.Error(), expected.Error())
}
