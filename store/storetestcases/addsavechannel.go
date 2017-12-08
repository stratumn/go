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
	"github.com/stratumn/sdk/store"
	"github.com/stretchr/testify/assert"
)

// TestLinkSavedChannel tests that the store correctly notifies listeners when a link is created.
func (f Factory) TestLinkSavedChannel(t *testing.T) {
	a := f.initAdapter(t)
	defer f.freeAdapter(a)

	c := make(chan *store.Event, 1)
	a.AddStoreEventChannel(c)

	link := cstesting.RandomLink()
	_, err := a.CreateLink(link)
	assert.NoError(t, err, "a.CreateLink()")

	got := <-c
	assert.EqualValues(t, store.SavedLinks, got.EventType, "Invalid event type")
	links := got.Data.([]*cs.Link)
	assert.Equal(t, 1, len(links), "Invalid number of links")
	assert.EqualValues(t, link, links[0], "Invalid link")
}

// TestEvidenceAddedChannel tests that the store correctly notifies listeners when some evidence is added.
func (f Factory) TestEvidenceAddedChannel(t *testing.T) {
	a := f.initAdapter(t)
	defer f.freeAdapter(a)

	c := make(chan *store.Event, 10)
	a.AddStoreEventChannel(c)

	link := cstesting.RandomLink()
	linkHash, err := a.CreateLink(link)
	assert.NoError(t, err, "a.CreateLink()")

	evidence := cstesting.RandomEvidence()
	err = a.AddEvidence(linkHash, evidence)
	assert.NoError(t, err, "a.AddEvidence()")

	// Ignore the link saved event
	got := <-c

	// There might be a race between the external evidence added
	// and an evidence produced by a blockchain store (hence the for loop)
	for i := 0; i < 3; i++ {
		got = <-c
		if got.EventType != store.SavedEvidences {
			continue
		}

		evidences := got.Data.(map[string]*cs.Evidence)
		e, found := evidences[linkHash.String()]
		if found && e.Backend == evidence.Backend {
			break
		}
	}

	assert.EqualValues(t, store.SavedEvidences, got.EventType, "Expected saved evidences")
	evidences := got.Data.(map[string]*cs.Evidence)
	assert.EqualValues(t, evidence, evidences[linkHash.String()])
}
