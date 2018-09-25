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

package evidences_test

import (
	"testing"

	"github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-core/batchfossilizer/evidences"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBatchFossilizerEvidence(t *testing.T) {
	t.Run("unmarshal-invalid-backend", func(t *testing.T) {
		proof := &evidences.BatchProof{Timestamp: 42}
		e, err := proof.Evidence("btc")
		require.NoError(t, err)

		e.Backend = ""
		_, err = evidences.UnmarshalProof(e)
		assert.EqualError(t, err, evidences.ErrInvalidBackend.Error())
	})

	t.Run("unmarshal-missing-provider", func(t *testing.T) {
		proof := &evidences.BatchProof{Timestamp: 42}
		e, err := proof.Evidence("btc")
		require.NoError(t, err)

		e.Provider = ""
		_, err = evidences.UnmarshalProof(e)
		assert.EqualError(t, err, chainscript.ErrMissingProvider.Error())
	})

	t.Run("unmarshal-invalid-version", func(t *testing.T) {
		proof := &evidences.BatchProof{Timestamp: 42}
		e, err := proof.Evidence("btc")
		require.NoError(t, err)

		e.Version = "0.42.0"
		_, err = evidences.UnmarshalProof(e)
		assert.EqualError(t, err, evidences.ErrUnknownVersion.Error())
	})

	t.Run("unmarshal", func(t *testing.T) {
		proof := &evidences.BatchProof{Timestamp: 42}
		e, err := proof.Evidence("btc")
		require.NoError(t, err)

		p, err := evidences.UnmarshalProof(e)
		assert.NoError(t, err)

		assert.Equal(t, int64(42), p.Timestamp)
	})
}
