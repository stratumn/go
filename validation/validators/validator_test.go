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

package validators_test

import (
	"testing"

	"github.com/stratumn/go-core/validation/validators"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlatten(t *testing.T) {
	pv := make(validators.ProcessesValidators)

	v1, _ := validators.NewProcessStepValidator("p1", "s1")
	pv["p1"] = []validators.Validator{v1}

	v2, _ := validators.NewProcessStepValidator("p2", "s2")
	v3, _ := validators.NewProcessStepValidator("p3", "s3")
	pv["p2"] = []validators.Validator{v2, v3}

	pv["p3"] = make(validators.Validators, 0)

	flattened := pv.Flatten()
	require.Len(t, flattened, 3)
	assert.ElementsMatch(t, flattened, []validators.Validator{v1, v2, v3})
}
