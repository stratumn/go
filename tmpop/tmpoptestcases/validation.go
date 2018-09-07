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

package tmpoptestcases

import (
	"os"
	"testing"

	"github.com/stratumn/go-chainscript/chainscripttest"
	"github.com/stratumn/go-indigocore/tmpop"
	"github.com/stratumn/go-indigocore/utils"
	"github.com/stratumn/go-indigocore/validation"
	"github.com/stretchr/testify/assert"
)

const testValidationConfig = `
{
	"testProcess": {
	    "pki": {
		"alice.vandenbudenmayer@stratumn.com": {
		    "keys": ["-----BEGIN ED25519 PUBLIC KEY-----\nMCowBQYDK2VwAyEAO0U2B1DjM7k+AWLUBl9oK+ZhX/gpwrx5Z7RxCUgccDo=\n-----END ED25519 PUBLIC KEY-----\n"],
		    "roles": ["employee"]
		}
	    },
	    "types": {
			"init": {
				"schema": {
					"type": "object",
					"properties": {
						"string": {
							"type": "string"
						}
					}
				},
				"transitions": [""]
			}
		},
		"action": {
		    "signatures": ["it"]
	    }
	}
}
`

// TestValidation tests what happens when validating a segment from a json-schema based validation file
func (f Factory) TestValidation(t *testing.T) {
	testFilename := utils.CreateTempFile(t, testValidationConfig)
	defer os.Remove(testFilename)

	h, req := f.newTMPop(t, &tmpop.Config{Validation: &validation.Config{RulesPath: testFilename}})
	defer f.free()

	h.BeginBlock(req)
	data := map[string]string{"string": "test"}

	t.Run("Validation succeeded", func(t *testing.T) {
		l := chainscripttest.NewLinkBuilder(t).
			WithProcess("testProcess").
			WithStep("init").
			WithoutParent().
			WithData(t, data).
			WithSignature(t, "").
			Build()
		tx := makeCreateLinkTx(t, l)
		res := h.DeliverTx(tx)

		assert.False(t, res.IsErr(), "a.DeliverTx(): failed")
	})

	t.Run("Link does not match any validator", func(t *testing.T) {
		l := chainscripttest.NewLinkBuilder(t).
			WithProcess("testProcess").
			WithStep("notfound").
			WithoutParent().
			Build()
		tx := makeCreateLinkTx(t, l)
		res := h.DeliverTx(tx)

		assert.True(t, res.IsErr(), "a.DeliverTx(): want error")
		assert.Equal(t, tmpop.CodeTypeValidation, res.Code, "res.Code")
	})

	t.Run("Schema validation failed", func(t *testing.T) {
		badData := map[string]interface{}{"string": 42}
		l := chainscripttest.NewLinkBuilder(t).
			WithProcess("testProcess").
			WithStep("init").
			WithoutParent().
			WithData(t, badData).
			Build()
		tx := makeCreateLinkTx(t, l)
		res := h.DeliverTx(tx)

		assert.True(t, res.IsErr(), "a.DeliverTx(): want error")
		assert.Equal(t, tmpop.CodeTypeValidation, res.Code, "res.Code")
	})

	t.Run("Signature validation failed", func(t *testing.T) {
		l := chainscripttest.NewLinkBuilder(t).
			WithProcess("testProcess").
			WithStep("init").
			WithoutParent().
			WithData(t, data).
			WithSignature(t, "").
			Build()

		l.Signatures[0].Signature = []byte(`-----BEGIN MESSAGE-----
BEDZR29+Zk8M72ZlgWstb3o96MdKNXeT0Q7LfzDFQKjv9dLjeHpRL4BSjkjPWbuA
Kmq1nHIk7T7bpLBohyy0lRYO
-----END MESSAGE-----`)

		tx := makeCreateLinkTx(t, l)
		res := h.DeliverTx(tx)

		assert.True(t, res.IsErr(), "a.DeliverTx(): want error")
		assert.Equal(t, tmpop.CodeTypeValidation, res.Code, "res.Code")

	})

}
