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
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/cs/cstesting"
	"github.com/stratumn/sdk/store"
)

// TestAddDidSaveChannel tests that AddDidSaveChannel functions properly.
func (f Factory) TestAddDidSaveChannel(t *testing.T) {
	a := f.initAdapter(t)
	defer f.free(a)

	c := make(chan *cs.Segment, 1)
	a.AddDidSaveChannel(c)

	s := cstesting.RandomSegment()
	if err := a.SaveSegment(s); err != nil {
		t.Fatalf("a.SaveSegment(): err: %s", err)
	}

	if got, want := <-c, s; !reflect.DeepEqual(want, got) {
		gotJS, _ := json.MarshalIndent(got, "", "  ")
		wantJS, _ := json.MarshalIndent(want, "", "  ")
		t.Errorf("<- c = %s\n want%s", gotJS, wantJS)
	}
}

// TestLinkSavedChannel tests that the store correctly notifies listeners when a link is created.
func (f Factory) TestLinkSavedChannel(t *testing.T) {
	a := f.initAdapterV2(t)
	defer f.freeV2(a)

	c := make(chan *store.Event, 1)
	a.AddStoreEventChannel(c)

	link := cstesting.RandomLink()
	if _, err := a.CreateLink(link); err != nil {
		t.Fatalf("a.CreateLink(); err: %s", err)
	}

	got := <-c
	if got.EventType != store.SavedLink {
		t.Errorf("Expected saved link event, got %v", got)
	}

	if !reflect.DeepEqual(link, got.Details) {
		gotJS, _ := json.MarshalIndent(got.Details, "", "  ")
		wantJS, _ := json.MarshalIndent(link, "", "  ")
		t.Errorf("<- c = %s\n want%s", gotJS, wantJS)
	}
}

// TestEvidenceAddedChannel tests that the store correctly notifies listeners when some evidence is added.
func (f Factory) TestEvidenceAddedChannel(t *testing.T) {
	a := f.initAdapterV2(t)
	defer f.freeV2(a)

	c := make(chan *store.Event, 1)
	a.AddStoreEventChannel(c)

	link := cstesting.RandomLink()
	linkHash, err := a.CreateLink(link)
	if err != nil {
		t.Fatalf("a.CreateLink(); err: %s", err)
	}

	// Ignore the link saved event
	got := <-c

	evidence := cstesting.RandomEvidence()
	if err = a.AddEvidence(linkHash, evidence); err != nil {
		t.Fatalf("a.AddEvidence(); err: %s", err)
	}

	got = <-c

	if got.EventType != store.SavedEvidence {
		t.Errorf("Expected saved evidence event, got %v", got)
	}

	if !reflect.DeepEqual(evidence, got.Details) {
		gotJS, _ := json.MarshalIndent(got.Details, "", "  ")
		wantJS, _ := json.MarshalIndent(evidence, "", "  ")
		t.Errorf("<- c = %s\n want%s", gotJS, wantJS)
	}
}
