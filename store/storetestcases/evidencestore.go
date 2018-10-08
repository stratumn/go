// Copyright 2017-2018 Stratumn SAS. All rights reserved.
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

package storetestcases

import (
	"context"
	"testing"

	"github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-chainscript/chainscripttest"
	"github.com/stratumn/go-core/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEvidenceStore runs all tests for the store.EvidenceStore interface
func (f Factory) TestEvidenceStore(t *testing.T) {
	a := f.initAdapter(t)
	defer f.freeAdapter(a)

	l := chainscripttest.RandomLink(t)
	linkHash, _ := a.CreateLink(context.Background(), l)

	s := store.EvidenceStore(a)

	t.Run("Adding evidences to a segment should work", func(t *testing.T) {
		ctx := context.Background()
		e1, _ := chainscript.NewEvidence("1.0.0", "TMPop", "1", []byte{1})
		e2, _ := chainscript.NewEvidence("1.0.0", "dummy", "2", []byte{2})
		e3, _ := chainscript.NewEvidence("1.0.0", "batch", "3", []byte{3})
		e4, _ := chainscript.NewEvidence("1.0.0", "bcbatch", "4", []byte{4})
		e5, _ := chainscript.NewEvidence("1.0.0", "generic", "5", []byte{5})
		evidences := []*chainscript.Evidence{e1, e2, e3, e4, e5}

		for _, evidence := range evidences {
			err := s.AddEvidence(ctx, linkHash, evidence)
			assert.NoError(t, err, "s.AddEvidence(ctx, )")
		}

		storedEvidences, err := s.GetEvidences(ctx, linkHash)
		assert.NoError(t, err, "s.GetEvidences()")
		assert.Equal(t, 5, len(storedEvidences), "Invalid number of evidences")

		for _, evidence := range evidences {
			foundEvidence := storedEvidences.FindEvidences(evidence.Backend)
			assert.Equal(t, 1, len(foundEvidence), "Evidence not found: %v", evidence)
		}
	})

	t.Run("Duplicate evidences should be discarded", func(t *testing.T) {
		ctx := context.Background()
		e, _ := chainscript.NewEvidence("1.0.0", "TMPop", "42", []byte{42})

		err := s.AddEvidence(ctx, linkHash, e)
		require.NoError(t, err, "s.AddEvidence()")
		// Add duplicate - some stores return an error, others silently ignore
		_ = s.AddEvidence(ctx, linkHash, e)

		storedEvidences, err := s.GetEvidences(ctx, linkHash)
		assert.NoError(t, err, "s.GetEvidences()")
		assert.Equal(t, 6, len(storedEvidences), "Invalid number of evidences")
		assert.EqualValues(t, e.Backend, storedEvidences.GetEvidence("TMPop", "42").Backend, "Invalid evidence backend")
	})
}
