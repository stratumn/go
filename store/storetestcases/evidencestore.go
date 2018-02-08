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

package storetestcases

import (
	"testing"

	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/cs/cstesting"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stretchr/testify/assert"
)

// TestEvidenceStore runs all tests for the store.EvidenceStore interface
func (f Factory) TestEvidenceStore(t *testing.T) {
	a := f.initAdapter(t)
	defer f.freeAdapter(a)

	l := cstesting.RandomLink()
	linkHash, _ := a.CreateLink(l)

	s := store.EvidenceStore(a)

	t.Run("Adding evidences to a segment should work", func(t *testing.T) {
		e1 := cs.Evidence{Backend: "TMPop", Provider: "1"}
		e2 := cs.Evidence{Backend: "dummy", Provider: "2"}
		e3 := cs.Evidence{Backend: "batch", Provider: "3"}
		e4 := cs.Evidence{Backend: "bcbatch", Provider: "4"}
		e5 := cs.Evidence{Backend: "generic", Provider: "5"}
		evidences := []*cs.Evidence{&e1, &e2, &e3, &e4, &e5}

		for _, evidence := range evidences {
			err := s.AddEvidence(linkHash, evidence)
			assert.NoError(t, err, "s.AddEvidence()")
		}

		storedEvidences, err := s.GetEvidences(linkHash)
		assert.NoError(t, err, "s.GetEvidences()")
		assert.Equal(t, 5, len(*storedEvidences), "Invalid number of evidences")

		for _, evidence := range evidences {
			foundEvidence := storedEvidences.FindEvidences(evidence.Backend)
			assert.Equal(t, 1, len(foundEvidence), "Evidence not found: %v", evidence)
		}
	})

	t.Run("Duplicate evidences should be discarded", func(t *testing.T) {
		e1 := cs.Evidence{Backend: "TMPop", Provider: "42"}
		e2 := cs.Evidence{Backend: "dummy", Provider: "42"}

		s.AddEvidence(linkHash, &e1)
		s.AddEvidence(linkHash, &e2)

		storedEvidences, err := s.GetEvidences(linkHash)
		assert.NoError(t, err, "s.GetEvidences()")
		assert.Equal(t, 6, len(*storedEvidences), "Invalid number of evidences")
		assert.EqualValues(t, e1.Backend, storedEvidences.GetEvidence("42").Backend, "Invalid evidence backend")
	})
}
