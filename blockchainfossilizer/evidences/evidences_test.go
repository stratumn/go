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

package evidences_test

import (
	"testing"

	"github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-core/blockchainfossilizer/evidences"
	"github.com/stratumn/go-core/monitoring/errorcode"
	"github.com/stratumn/go-core/testutil"
	"github.com/stratumn/go-core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBlockchainFossilizerEvidence(t *testing.T) {
	t.Run("Verify", func(t *testing.T) {
		t.Run("missing transaction id", func(t *testing.T) {
			proof := evidences.New([]byte{42}, nil)
			assert.False(t, proof.Verify([]byte{42}))
		})

		t.Run("data mismatch", func(t *testing.T) {
			proof := evidences.New([]byte{42}, []byte{43})
			assert.False(t, proof.Verify([]byte{43}))
		})

		t.Run("success", func(t *testing.T) {
			proof := evidences.New([]byte{42}, []byte{43})
			assert.True(t, proof.Verify([]byte{42}))
		})
	})

	t.Run("Unmarshal", func(t *testing.T) {
		t.Run("invalid backend", func(t *testing.T) {
			proof := &evidences.BlockchainProof{}
			e, err := proof.Evidence("btc")
			require.NoError(t, err)

			e.Backend = ""
			_, err = evidences.UnmarshalProof(e)

			assert.Equal(t, errorcode.InvalidArgument, err.(*types.Error).Code)
			testutil.AssertWrappedErrorEqual(t, err, evidences.ErrInvalidBackend)
		})

		t.Run("missing provider", func(t *testing.T) {
			proof := &evidences.BlockchainProof{}
			e, err := proof.Evidence("btc")
			require.NoError(t, err)

			e.Provider = ""
			_, err = evidences.UnmarshalProof(e)

			assert.Equal(t, errorcode.InvalidArgument, err.(*types.Error).Code)
			testutil.AssertWrappedErrorEqual(t, err, chainscript.ErrMissingProvider)
		})

		t.Run("invalid version", func(t *testing.T) {
			proof := &evidences.BlockchainProof{}
			e, err := proof.Evidence("btc")
			require.NoError(t, err)

			e.Version = "0.42.0"
			_, err = evidences.UnmarshalProof(e)

			assert.Equal(t, errorcode.InvalidArgument, err.(*types.Error).Code)
			testutil.AssertWrappedErrorEqual(t, err, evidences.ErrUnknownVersion)
		})

		t.Run("success", func(t *testing.T) {
			proof := &evidences.BlockchainProof{TransactionID: []byte{42}}
			e, err := proof.Evidence("btc")
			require.NoError(t, err)

			p, err := evidences.UnmarshalProof(e)
			assert.NoError(t, err)

			assert.Equal(t, types.TransactionID([]byte{42}), p.TransactionID)
		})
	})
}
