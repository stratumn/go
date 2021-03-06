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

package storetestcases

import (
	"context"
	"testing"
	"time"

	"github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-chainscript/chainscripttest"
	"github.com/stratumn/go-core/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStoreEvents tests store channel event notifications.
func (f Factory) TestStoreEvents(t *testing.T) {
	a := f.initAdapter(t)
	defer f.freeAdapter(a)

	c := make(chan *store.Event, 10)
	a.AddStoreEventChannel(c)

	link := chainscripttest.RandomLink(t)
	linkHash, err := a.CreateLink(context.Background(), link)
	assert.NoError(t, err, "a.CreateLink()")

	t.Run("Link saved event should be sent to channel", func(t *testing.T) {
		select {
		case got := <-c:
			assert.EqualValues(t, store.SavedLinks, got.EventType, "Invalid event type")
			links := got.Data.([]*chainscript.Link)
			assert.Equal(t, 1, len(links), "Invalid number of links")
			chainscripttest.LinksEqual(t, link, links[0])
		case <-time.After(10 * time.Second):
			require.Fail(t, "Timeout waiting for link saved event")
		}
	})

	t.Run("Evidence saved event should be sent to channel", func(t *testing.T) {
		ctx := context.Background()
		evidence := chainscripttest.RandomEvidence(t)
		err = a.AddEvidence(ctx, linkHash, evidence)
		assert.NoError(t, err, "a.AddEvidence()")

		var got *store.Event

		// There might be a race between the external evidence added
		// and an evidence produced by a blockchain store (hence the for loop)
		for i := 0; i < 3; i++ {
			select {
			case got = <-c:
			case <-time.After(10 * time.Second):
				require.Fail(t, "Timeout waiting for evidence saved event")
			}

			if got.EventType != store.SavedEvidences {
				continue
			}

			evidences := got.Data.(map[string]*chainscript.Evidence)
			e, found := evidences[linkHash.String()]
			if found && e.Backend == evidence.Backend {
				break
			}
		}

		assert.EqualValues(t, store.SavedEvidences, got.EventType, "Expected saved evidences")
		evidences := got.Data.(map[string]*chainscript.Evidence)
		assert.EqualValues(t, evidence, evidences[linkHash.String()])
	})
}
