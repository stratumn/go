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

	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/cs/cstesting"
)

// TestAddEvidences tests what happens when you add evidence
// to a segment.
func (f Factory) TestAddEvidences(t *testing.T) {
	a := f.initAdapter(t)
	defer f.freeAdapter(a)

	e1 := cs.Evidence{Backend: "TMPop", Provider: "1"}
	e2 := cs.Evidence{Backend: "dummy", Provider: "2"}
	e3 := cs.Evidence{Backend: "batch", Provider: "3"}
	e4 := cs.Evidence{Backend: "bcbatch", Provider: "4"}
	e5 := cs.Evidence{Backend: "generic", Provider: "5"}
	evidences := []*cs.Evidence{&e1, &e2, &e3, &e4, &e5}

	l := cstesting.RandomLink()
	linkHash, _ := a.CreateLink(l)

	for _, evidence := range evidences {
		if err := a.AddEvidence(linkHash, evidence); err != nil {
			t.Fatalf("a.AddEvidence(): err: %s", err)
		}
	}

	storedEvidences, err := a.GetEvidences(linkHash)
	if err != nil {
		t.Fatalf("a.GetEvidences(): err: %s", err)
	}
	if len(*storedEvidences) != 5 {
		t.Fatalf("Invalid number of evidences: got %d, want %d",
			len(*storedEvidences), 5)
	}

	for _, evidence := range evidences {
		foundEvidence := storedEvidences.FindEvidences(evidence.Backend)
		if len(foundEvidence) != 1 {
			t.Fatalf("Evidence not found: %v", evidence)
		}
	}
}

// TestAddDuplicateEvidences tests that evidence is ignored when an evidence
// from the same provider has already been added.
func (f Factory) TestAddDuplicateEvidences(t *testing.T) {
	a := f.initAdapter(t)
	defer f.freeAdapter(a)

	e1 := cs.Evidence{Backend: "TMPop", Provider: "42"}
	e2 := cs.Evidence{Backend: "dummy", Provider: "42"}

	l := cstesting.RandomLink()
	linkHash, _ := a.CreateLink(l)
	a.AddEvidence(linkHash, &e1)
	a.AddEvidence(linkHash, &e2)

	storedEvidences, err := a.GetEvidences(linkHash)
	if err != nil {
		t.Fatalf("a.GetEvidences(): err: %s", err)
	}
	if len(*storedEvidences) != 1 {
		t.Fatalf("Invalid number of evidences: got %d, want %d",
			len(*storedEvidences), 1)
	}

	if storedEvidences.GetEvidence("42").Backend != e1.Backend {
		t.Fatalf("Invalid evidence saved: got %s, want %s",
			"42", e1.Backend)
	}
}
