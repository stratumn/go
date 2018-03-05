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

package validator

import (
	"testing"

	"github.com/stratumn/go-indigocore/cs/cstesting"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInvalidValidator(t *testing.T) {
	t.Parallel()
	const errorMessage = "FooBar"
	baseCfg, err := newValidatorBaseConfig("process", "linkType")
	require.NoError(t, err)
	v := newInvalidValidator(baseCfg, errorMessage)
	require.NotNil(t, v)
	l := cstesting.RandomLink()
	l.Meta.Process = "bad process"
	assert.False(t, v.ShouldValidate(l))
	l.Meta.Process = "process"
	l.Meta.Type = "linkType"
	assert.True(t, v.ShouldValidate(l))
	assert.EqualError(t, v.Validate(nil, nil), errorMessage)
}

func TestInvalidHash(t *testing.T) {
	t.Parallel()
	baseCfg, err := newValidatorBaseConfig("process", "linkType")
	require.NoError(t, err)
	v1 := newInvalidValidator(baseCfg, "foo")
	v2 := newInvalidValidator(baseCfg, "bar")

	hash1, err1 := v1.Hash()
	hash2, err2 := v2.Hash()
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NotNil(t, hash1)
	assert.NotNil(t, hash2)
	assert.NotEqual(t, hash1.String(), hash2.String())
}
