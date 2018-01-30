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

package cs_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stratumn/sdk/cs"
)

func TestGetSignatures(t *testing.T) {
	signatures := cs.Signatures{
		&cs.Signature{
			PublicKey: "one",
		},
		&cs.Signature{
			PublicKey: "two",
		},
	}
	got := signatures.Get("one")
	assert.EqualValues(t, signatures[0], got, "signatures.Get()")
}

func TestGetSignatures_NotFound(t *testing.T) {
	signatures := cs.Signatures{
		&cs.Signature{
			PublicKey: "one",
		},
		&cs.Signature{
			PublicKey: "two",
		},
	}
	got := signatures.Get("wrong")
	assert.Nil(t, got, "s.Get(()")
}
