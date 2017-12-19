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

package tmpop

import (
	"sync"

	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/types"
)

// Tendermint doesn't allow us to fire arbitrary events to notify TMStore.
// So instead we store pending events here, and TMStore will query them
// when a new block is produced.
type eventsManager struct {
	pendingEvents []*store.Event
	lock          sync.Mutex
}

func (e *eventsManager) AddSavedLinks(links []*cs.Link) {
	if len(links) > 0 {
		savedEvent := store.NewSavedLinks(links...)

		e.lock.Lock()
		defer e.lock.Unlock()
		e.pendingEvents = append(e.pendingEvents, savedEvent)
	}
}

func (e *eventsManager) AddSavedEvidences(evidences map[*types.Bytes32]*cs.Evidence) {
	if len(evidences) > 0 {
		evidenceEvent := store.NewSavedEvidences()
		for linkHash, evidence := range evidences {
			evidenceEvent.AddSavedEvidence(linkHash, evidence)
		}

		e.lock.Lock()
		defer e.lock.Unlock()
		e.pendingEvents = append(e.pendingEvents, evidenceEvent)
	}
}

func (e *eventsManager) GetPendingEvents() []*store.Event {
	e.lock.Lock()
	defer e.lock.Unlock()

	eventsToDeliver := make([]*store.Event, len(e.pendingEvents))
	copy(eventsToDeliver, e.pendingEvents)
	e.pendingEvents = nil

	return eventsToDeliver
}
