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

// TestStoreEvents tests store channel event notifications.
func (f Factory) TestStoreEvents(t *testing.T) {
	a := f.initAdapter(t)
	defer f.freeAdapter(a)

	c := make(chan *store.Event, 10)
	a.AddStoreEventChannel(c)

	link := cstesting.RandomLink()
	linkHash, err := a.CreateLink(link)
	assert.NoError(t, err, "a.CreateLink()")

	t.Run("Link saved event should be sent to channel", func(t *testing.T) {
		got := <-c
		assert.EqualValues(t, store.SavedLinks, got.EventType, "Invalid event type")
		links := got.Data.([]*cs.Link)
		assert.Equal(t, 1, len(links), "Invalid number of links")
		assert.EqualValues(t, link, links[0], "Invalid link")
	})

	t.Run("Evidence saved event should be sent to channel", func(t *testing.T) {
		evidence := cstesting.RandomEvidence()
		err = a.AddEvidence(linkHash, evidence)
		assert.NoError(t, err, "a.AddEvidence()")

		var got *store.Event

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
	})
}
