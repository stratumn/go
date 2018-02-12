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
	"io/ioutil"
	"os"
	"testing"

	"github.com/stratumn/go-indigocore/cs/cstesting"
	"github.com/stratumn/go-indigocore/tmpop"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testValidationConfig = `
{
	"testProcess": {
	    "pki": {
		"alice.vandenbudenmayer@stratumn.com": {
		    "keys": [
			"TESTKEY1"
		    ],
		    "roles": [
			"employee"
		    ]
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
		    }
		},
		"action": {
		    "signatures": [
			"it"
		    ]
		}
	    }
	}
    }
`

func createValidationFile(t *testing.T) string {
	tmpfile, err := ioutil.TempFile("", "validation-config")
	require.NoError(t, err, "ioutil.TempFile()")

	_, err = tmpfile.WriteString(testValidationConfig)
	require.NoError(t, err, "tmpfile.WriteString()")

	return tmpfile.Name()
}

// TestValidation tests what happens when validating a segment from a json-schema based validation file
func (f Factory) TestValidation(t *testing.T) {
	testFilename := createValidationFile(t)
	defer os.Remove(testFilename)

	h, req := f.newTMPop(t, &tmpop.Config{ValidatorFilename: testFilename})
	defer f.free()

	h.BeginBlock(req)

	t.Run("Validation succeeded", func(t *testing.T) {
		l := cstesting.RandomLinkWithProcess("testProcess")
		l.Meta["type"] = "init"
		l.State["string"] = "test"
		l = cstesting.SignLink(l)
		tx := makeCreateLinkTx(t, l)
		res := h.DeliverTx(tx)

		assert.False(t, res.IsErr(), "a.DeliverTx(): failed")
	})

	t.Run("Link does not match any validator", func(t *testing.T) {
		l := cstesting.RandomLinkWithProcess("testProcess")
		l.Meta["type"] = "notfound"
		tx := makeCreateLinkTx(t, l)
		res := h.DeliverTx(tx)

		assert.True(t, res.IsErr(), "a.DeliverTx(): want error")
		assert.Equal(t, tmpop.CodeTypeValidation, res.Code, "res.Code")
	})

	t.Run("Schema validation failed", func(t *testing.T) {
		l := cstesting.RandomLinkWithProcess("testProcess")
		l.Meta["type"] = "init"
		l.State["string"] = 42
		tx := makeCreateLinkTx(t, l)
		res := h.DeliverTx(tx)

		assert.True(t, res.IsErr(), "a.DeliverTx(): want error")
		assert.Equal(t, tmpop.CodeTypeValidation, res.Code, "res.Code")
	})

	t.Run("Signature validation failed", func(t *testing.T) {
		l := cstesting.RandomLinkWithProcess("testProcess")
		l.Meta["type"] = "init"
		l.State["string"] = "test"
		l = cstesting.SignLink(l)
		l.Signatures[0].Signature = "test"
		tx := makeCreateLinkTx(t, l)
		res := h.DeliverTx(tx)

		assert.True(t, res.IsErr(), "a.DeliverTx(): want error")
		assert.Equal(t, tmpop.CodeTypeValidation, res.Code, "res.Code")

	})

}
