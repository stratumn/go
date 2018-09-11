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

package types_test

import (
	"testing"

	"github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-indigocore/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvidenceSlice(t *testing.T) {
	t.Run("AddEvidence", func(t *testing.T) {
		t.Run("new evidence provider", func(t *testing.T) {
			e := make(types.EvidenceSlice, 1, 1)
			e[0] = &chainscript.Evidence{Backend: "btc", Provider: "testnet"}

			err := e.AddEvidence(&chainscript.Evidence{Backend: "btc", Provider: "mainnet"})
			require.NoError(t, err)

			assert.Len(t, e, 2)
			assert.Equal(t, "testnet", e[0].Provider)
			assert.Equal(t, "mainnet", e[1].Provider)
		})

		t.Run("existing backend provider", func(t *testing.T) {
			e := make(types.EvidenceSlice, 1, 1)
			e[0] = &chainscript.Evidence{Backend: "btc", Provider: "testnet"}

			err := e.AddEvidence(&chainscript.Evidence{Backend: "btc", Provider: "testnet"})
			assert.Error(t, err)
			assert.Len(t, e, 1)
		})
	})

	t.Run("GetEvidence", func(t *testing.T) {
		t.Run("missing evidence", func(t *testing.T) {
			e := types.EvidenceSlice{&chainscript.Evidence{Backend: "btc", Provider: "testnet"}}
			res := e.GetEvidence("ethereum", "mainnet")
			assert.Nil(t, res)
		})

		t.Run("provider mismatch", func(t *testing.T) {
			e := types.EvidenceSlice{&chainscript.Evidence{Backend: "btc", Provider: "testnet"}}
			res := e.GetEvidence("btc", "mainnet")
			assert.Nil(t, res)
		})

		t.Run("backend mismatch", func(t *testing.T) {
			e := types.EvidenceSlice{&chainscript.Evidence{Backend: "btc", Provider: "testnet"}}
			res := e.GetEvidence("ethereum", "testnet")
			assert.Nil(t, res)
		})

		t.Run("evidence match", func(t *testing.T) {
			e := types.EvidenceSlice{&chainscript.Evidence{Backend: "btc", Provider: "testnet"}}
			res := e.GetEvidence("btc", "testnet")

			require.NotNil(t, res)
			assert.Equal(t, "btc", res.Backend)
			assert.Equal(t, "testnet", res.Provider)
		})
	})

	t.Run("FindEvidences", func(t *testing.T) {
		t.Run("missing evidence", func(t *testing.T) {
			var e types.EvidenceSlice
			err := e.AddEvidence(&chainscript.Evidence{
				Backend:  "btc",
				Provider: "testnet",
			})
			require.NoError(t, err)

			e2 := e.FindEvidences("eth")
			assert.Empty(t, e2)
		})

		t.Run("evidence match", func(t *testing.T) {
			var e types.EvidenceSlice
			require.NoError(t, e.AddEvidence(&chainscript.Evidence{
				Backend:  "btc",
				Provider: "testnet",
			}))
			require.NoError(t, e.AddEvidence(&chainscript.Evidence{
				Backend:  "btc",
				Provider: "mainnet",
			}))

			e2 := e.FindEvidences("btc")
			assert.Len(t, e2, 2)
		})
	})
}
